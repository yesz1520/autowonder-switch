package autowonder

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSignatureEnvelope(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1000, 0)
	expiry := now.Add(time.Hour).UTC().Format(time.RFC3339)
	f := TaskDispatch{PackageID: "pkg_701", Checksum: "sha256:" + strings.Repeat("a", 64), Issuer: "autowonder-server", SignatureAlgorithm: "ed25519", SignatureRef: fmt.Sprintf("sha256:%x", sha256.Sum256(der)), SignaturePublicKey: base64.StdEncoding.EncodeToString(der), ExpiresAt: expiry}
	payload := fmt.Sprintf("autowonder.taskPackage.signature.v1\n%s\n%s\n%s\n%s\n%s\n", f.Issuer, f.SignatureRef, expiry, f.PackageID, f.Checksum)
	f.Signature = base64.StdEncoding.EncodeToString(ed25519.Sign(priv, []byte(payload)))
	if err = f.verifyEnvelope(now); err != nil {
		t.Fatal(err)
	}
	if err = f.verifyEnvelope(now.Add(time.Hour)); err == nil {
		t.Fatal("accepted expired signature")
	}
	f.PackageID = "pkg_702"
	if err = f.verifyEnvelope(now); err == nil {
		t.Fatal("accepted altered ID")
	}
}

// Run with the unmodified 1.3.0 server fixture from src/test/resources/contract.
func TestUpstreamArchive(t *testing.T) {
	path := os.Getenv("AUTOWONDER_CONTRACT_FIXTURE")
	if path == "" {
		t.Skip("set AUTOWONDER_CONTRACT_FIXTURE to upstream server-task-package-v1.json")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var v struct{ ArchiveBase64, SHA256, PackageID, Issuer, SignatureRef, Signature, SignatureAlgorithm, SignaturePublicKey, ExpiresAt string }
	if err = json.Unmarshal(b, &v); err != nil {
		t.Fatal(err)
	}
	archive, err := base64.StdEncoding.DecodeString(v.ArchiveBase64)
	if err != nil {
		t.Fatal(err)
	}
	checksum := fmt.Sprintf("sha256:%x", sha256.Sum256(archive))
	if strings.TrimPrefix(checksum, "sha256:") != strings.TrimPrefix(v.SHA256, "sha256:") {
		t.Fatal("upstream archive digest mismatch")
	}
	f := TaskDispatch{Checksum: checksum, PackageID: v.PackageID, Issuer: v.Issuer, SignatureRef: v.SignatureRef, Signature: v.Signature, SignatureAlgorithm: v.SignatureAlgorithm, SignaturePublicKey: v.SignaturePublicKey, ExpiresAt: v.ExpiresAt}
	expiry, err := time.Parse(time.RFC3339, v.ExpiresAt)
	if err != nil {
		t.Fatal(err)
	}
	if err = f.verifyEnvelope(expiry.Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}
	tmp := filepath.Join(t.TempDir(), "package.zip")
	if err := os.WriteFile(tmp, archive, 0600); err != nil {
		t.Fatal(err)
	}
	manifest, err := ReadTaskPackage(tmp)
	if err != nil || manifest.PackageID != v.PackageID || len(manifest.FileDigests) == 0 {
		t.Fatal(manifest.PackageID, err)
	}
}
