package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/yesz1520/autowonder-switch/internal/taskhooks"
	"path/filepath"
	"strings"
)

type deliveryIntent struct {
	Name    string `json:"name"`
	Mode    string `json:"mode"`
	Branch  string `json:"branch"`
	Message string `json:"message"`
}
type codeReply struct {
	Success      bool             `json:"success"`
	Summary      string           `json:"summary"`
	Repositories []deliveryIntent `json:"repositories"`
}

func deliverySchema() map[string]any {
	str := map[string]string{"type": "string"}
	item := map[string]any{"type": "object", "additionalProperties": false, "properties": map[string]any{"name": str, "mode": map[string]any{"type": "string", "enum": []string{"patch", "commit", "push"}}, "branch": str, "message": str}, "required": []string{"name", "mode", "branch", "message"}}
	return map[string]any{"type": "object", "additionalProperties": false, "properties": map[string]any{"success": map[string]string{"type": "boolean"}, "summary": str, "repositories": map[string]any{"type": "array", "items": item}}, "required": []string{"success", "summary", "repositories"}}
}

const deliveryInstructions = `Edit and test repository files in the working directory. The executor performs Git delivery after your turn. Do not create .git directories, commit, push or edit the executor state yourself. Report one delivery intent per repository by its exact name. Derive mode from the work item: patch for review, analysis or edits without a requested commit, commit for a requested local commit, push for a requested remote branch update. Permissions alone never request a push. Do not downgrade an explicitly requested delivery, even if permissions are missing. Use the exact requested or supplied delivery branch and a meaningful commit message for commit/push. If a required target is ambiguous, or the task failed, set success=false and explain. Describe completed edits and tests; the executor verifies Git outcomes separately.`

type revision struct {
	Name       string `json:"name"`
	Base       string `json:"baseCommit"`
	Head       string `json:"headCommit"`
	Branch     string `json:"branch,omitempty"`
	RemoteHead string `json:"remoteHeadCommit"`
}

func (c checkout) revision() revision { return revision{c.name, c.base, c.head, c.branch, c.verified} }
func (c checkout) validate(ctx context.Context, i deliveryIntent) error {
	if i.Mode == "patch" {
		return nil
	}
	if i.Mode != "commit" && i.Mode != "push" {
		return errors.New("invalid delivery mode")
	}
	if !c.spec.AllowCommit {
		return errors.New("commit denied by repository policy")
	}
	if i.Mode == "push" && (!c.spec.AllowPush || !c.spec.AllowNetwork || c.remote == "") {
		return errors.New("push requires an authorized remote, network and push permission")
	}
	if i.Branch == "" || i.Branch == "HEAD" || i.Branch == "@" || strings.HasPrefix(i.Branch, "-") || strings.HasPrefix(i.Branch, "refs/") || strings.TrimSpace(i.Message) == "" {
		return errors.New("explicit branch and commit message required")
	}
	if _, err := git(ctx, "check-ref-format", "refs/heads/"+i.Branch); err != nil {
		return errors.New("invalid delivery branch")
	}
	if c.spec.DeliveryBranch != "" && c.spec.DeliveryBranch != i.Branch {
		return errors.New("branch differs from package delivery branch")
	}
	if len(c.spec.AllowedBranchPatterns) > 0 {
		for _, pattern := range c.spec.AllowedBranchPatterns {
			prefix, wild := strings.CutSuffix(pattern, "*")
			if pattern == i.Branch || wild && strings.HasPrefix(i.Branch, prefix) {
				return nil
			}
		}
		return errors.New("branch outside repository policy")
	}
	return nil
}
func (c *checkout) commit(ctx context.Context, i deliveryIntent) error {
	tree, err := c.git(ctx, "write-tree")
	if err != nil {
		return err
	}
	parentTree, err := c.git(ctx, "rev-parse", c.head+"^{tree}")
	if err != nil {
		return err
	}
	if string(tree) != string(parentTree) {
		id, err := c.git(ctx, "commit-tree", strings.TrimSpace(string(tree)), "-p", c.head, "-m", i.Message)
		if err != nil {
			return err
		}
		c.head = strings.TrimSpace(string(id))
		c.verified = ""
	}
	c.branch = i.Branch
	if _, err = c.git(ctx, "update-ref", "refs/heads/"+i.Branch, c.head); err != nil {
		return err
	}
	_, err = c.git(ctx, "update-ref", "--no-deref", "HEAD", c.head)
	return err
}
func (c *checkout) publish(ctx context.Context) error {
	return c.publishWithHooks(ctx, nil)
}
func (c *checkout) publishWithHooks(ctx context.Context, h *taskhooks.Set) error {
	ref := "refs/heads/" + c.branch
	run := func(args ...string) ([]byte, error) {
		return gitCommand(ctx, true, filepath.IsAbs(c.remote), append([]string{"--git-dir", c.gitDir}, args...)...)
	}
	check := func() bool {
		out, e := run("ls-remote", "--refs", "--", c.remote, ref)
		return e == nil && strings.TrimSpace(string(out)) == c.head+"\t"+ref
	}
	if !check() {
		if c.verified == c.head {
			return fmt.Errorf("previously verified branch %s changed since delivery", ref)
		}
		if err := h.Run(ctx, "beforePush", c.workDir, map[string]any{"branch": c.branch, "commit": c.head}); err != nil {
			return err
		}
		_, pushErr := run("push", "--no-follow-tags", "--", c.remote, c.head+":"+ref)
		if !check() {
			return fmt.Errorf("remote branch %s does not match %s (push error: %v)", ref, c.head, pushErr)
		}
	}
	c.verified = c.head
	return nil
}

type preparedRepo struct {
	Name, Base, Head, Remote, Branch, Verified string
	Spec                                       repository
	Intent                                     deliveryIntent
}
type preparedDelivery struct {
	PackageID, Checksum        string
	ResultID, Session, Summary string
	Repos                      []preparedRepo
}

func prepareDelivery(ctx context.Context, dir, resultID, session, text string, repos []checkout, h *taskhooks.Set) (preparedDelivery, error) {
	d := preparedDelivery{ResultID: resultID, Session: session, Summary: text}
	if len(repos) == 0 {
		return d, nil
	}
	var response codeReply
	if err := json.Unmarshal([]byte(text), &response); err != nil {
		return d, errors.New("missing structured code delivery result")
	}
	if !response.Success || strings.TrimSpace(response.Summary) == "" {
		return d, fmt.Errorf("code task incomplete: %s", response.Summary)
	}
	d.Summary = response.Summary
	intents := map[string]deliveryIntent{}
	for _, i := range response.Repositories {
		if _, ok := intents[i.Name]; ok {
			return d, errors.New("duplicate repository delivery")
		}
		intents[i.Name] = i
	}
	if len(intents) != len(repos) {
		return d, errors.New("incomplete repository delivery")
	}
	for _, c := range repos {
		i, ok := intents[c.name]
		if !ok {
			return d, fmt.Errorf("no delivery intent for %s", c.name)
		}
		if err := c.validate(ctx, i); err != nil {
			return d, fmt.Errorf("%s: %w", c.name, err)
		}
	}
	for n := range repos {
		c := &repos[n]
		i := intents[c.name]
		patch, err := c.patch(ctx)
		if err != nil {
			return d, err
		}
		if err = writeFile(filepath.Join(dir, "change-"+c.name+".patch"), patch); err != nil {
			return d, err
		}
		if i.Mode != "patch" {
			if err = h.Run(ctx, "beforeCommit", c.workDir, map[string]any{"branch": i.Branch}); err != nil {
				return d, err
			}
			if err = c.commit(ctx, i); err != nil {
				return d, err
			}
		}
		d.Repos = append(d.Repos, preparedRepo{c.name, c.base, c.head, c.remote, c.branch, c.verified, c.spec, i})
	}
	return d, nil
}
func (p preparedRepo) checkout(dir string) checkout {
	return checkout{name: p.Name, base: p.Base, head: p.Head, remote: p.Remote, branch: p.Branch, verified: p.Verified, spec: p.Spec, gitDir: filepath.Join(dir, "git-"+p.Name), workDir: filepath.Join(dir, "workspace", p.Name)}
}

func loadDelivery(dir string) (preparedDelivery, bool, error) {
	var d preparedDelivery
	ok, err := cached(filepath.Join(dir, "delivery.json"), &d)
	return d, ok, err
}
