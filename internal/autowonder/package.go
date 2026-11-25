package autowonder

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// DownloadPackage checks the envelope if present, then atomically installs
// the ZIP once its bytes match the SHA-256 declared in the dispatch.
func (c *Client) DownloadPackage(ctx context.Context, task *TaskDispatch, dest string) error {
	if task == nil || task.ChecksumAlgorithm != "sha256" || task.ChecksumScope != "zip_archive" || !strings.HasPrefix(task.Checksum, "sha256:") {
		return errors.New("unsupported package checksum")
	}
	if err := task.verifyEnvelope(time.Now()); err != nil {
		return err
	}
	return c.downloadVerified(ctx, task.DownloadURL, task.Checksum, dest, func(temp string) error {
		info, err := os.Stat(temp)
		if err != nil {
			return err
		}
		if task.Size > 0 && info.Size() != task.Size {
			return errors.New("unexpected archive size")
		}
		manifest, err := ReadTaskPackage(temp)
		if err != nil {
			return err
		}
		if manifest.DispatchID != strconv.FormatInt(task.DispatchID, 10) || manifest.PackageID != task.PackageID {
			return errors.New("manifest does not match dispatch")
		}
		return nil
	})
}
func (t *TaskDispatch) verifyEnvelope(now time.Time) error {
	if t.Signature == "" && t.SignatureRef == "" && t.SignatureAlgorithm == "" && t.SignaturePublicKey == "" && t.Issuer == "" && t.ExpiresAt == "" {
		return nil
	}
	if t.SignatureAlgorithm != "ed25519" || t.PackageID == "" || t.Issuer == "" {
		return errors.New("incomplete or unsupported package signature")
	}
	expires, err := time.Parse(time.RFC3339Nano, t.ExpiresAt)
	if err != nil || !now.Before(expires) {
		return errors.New("invalid package signature expiry")
	}
	der, err := base64.StdEncoding.DecodeString(t.SignaturePublicKey)
	if err != nil {
		return errors.New("invalid package public key")
	}
	if fmt.Sprintf("sha256:%x", sha256.Sum256(der)) != t.SignatureRef {
		return errors.New("package key fingerprint mismatch")
	}
	key, err := x509.ParsePKIXPublicKey(der)
	if err != nil {
		return errors.New("invalid package public key encoding")
	}
	pub, ok := key.(ed25519.PublicKey)
	if !ok {
		return errors.New("unexpected package public key type")
	}
	sig, err := base64.StdEncoding.DecodeString(t.Signature)
	if err != nil {
		return errors.New("invalid package signature encoding")
	}
	payload := strings.Join([]string{"autowonder.taskPackage.signature.v1", t.Issuer, t.SignatureRef, t.ExpiresAt, t.PackageID, strings.ToLower(t.Checksum), ""}, "\n")
	if !ed25519.Verify(pub, []byte(payload), sig) {
		return errors.New("invalid package signature")
	}
	return nil
}
