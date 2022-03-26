package run

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vercel/turborepo/cli/internal/fs"
)

func Test_manuallyHashPackage(t *testing.T) {
	rootIgnore := strings.Join([]string{
		"ignoreme",
		"ignorethisdir/",
	}, "\n")
	pkgIgnore := strings.Join([]string{
		"pkgignoreme",
		"pkgignorethisdir/",
	}, "\n")
	root, err := os.MkdirTemp("", "turbo-manual-file-hashing-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	pkgName := "libA"
	files := map[string]struct {
		contents string
		hash     string
	}{
		"top-level-file":              {"top-level-file-contents", ""},
		"other-dir/other-dir-file":    {"other-dir-file-contents", ""},
		"ignoreme":                    {"anything", ""},
		"libA/some-file":              {"some-file-contents", "7e59c6a6ea9098c6d3beb00e753e2c54ea502311"},
		"libA/some-dir/other-file":    {"some-file-contents", "7e59c6a6ea9098c6d3beb00e753e2c54ea502311"},
		"libA/ignoreme":               {"anything", ""},
		"libA/ignorethisdir/anything": {"anything", ""},
		"libA/pkgignoreme":            {"anything", ""},
		"libA/pkgignorethisdir/file":  {"anything", ""},
	}

	rootIgnoreFile, err := os.Create(filepath.Join(root, ".gitignore"))
	if err != nil {
		t.Fatalf("failed to create .gitignore: %v", err)
	}
	_, err = rootIgnoreFile.WriteString(rootIgnore)
	if err != nil {
		t.Fatalf("failed to write contents to .gitignore: %v", err)
	}
	rootIgnoreFile.Close()
	pkgIgnoreFilename := filepath.Join(root, pkgName, ".gitignore")
	err = fs.EnsureDir(pkgIgnoreFilename)
	if err != nil {
		t.Fatalf("failed to ensure directories for %v: %v", pkgIgnoreFilename, err)
	}
	pkgIgnoreFile, err := os.Create(pkgIgnoreFilename)
	if err != nil {
		t.Fatalf("failed to create libA/.gitignore: %v", err)
	}
	_, err = pkgIgnoreFile.WriteString(pkgIgnore)
	if err != nil {
		t.Fatalf("failed to write contents to libA/.gitignore: %v", err)
	}
	pkgIgnoreFile.Close()
	for path, spec := range files {
		filename := filepath.Join(root, path)
		err = fs.EnsureDir(filename)
		if err != nil {
			t.Fatalf("failed to ensure directories for %v: %v", filename, err)
		}
		f, err := os.Create(filename)
		if err != nil {
			t.Fatalf("failed to create file: %v: %v", filename, err)
		}
		_, err = f.WriteString(spec.contents)
		if err != nil {
			t.Fatalf("failed to write contents to %v: %v", filename, err)
		}
		f.Close()
	}

	hashes, err := manuallyHashPackage(&fs.PackageJSON{
		Dir: filepath.Join(root, pkgName),
	}, []string{})
	if err != nil {
		t.Fatalf("failed to calculate manual hashes: %v", err)
	}
	prefix := pkgName + "/"
	prefixLen := len(prefix)
	for path, spec := range files {
		if strings.HasPrefix(path, prefix) {
			got, ok := hashes[path[prefixLen:]]
			if !ok {
				if spec.hash != "" {
					t.Errorf("did not find hash for %v, but wanted one", path)
				}
			} else if got != spec.hash {
				t.Errorf("hash of %v, got %v want %v", path, got, spec.hash)
			}
		}
	}
}
