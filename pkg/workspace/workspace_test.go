package workspace

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func setupFS(t *testing.T) *LocalFileSystem {
	t.Helper()
	root := t.TempDir()
	fs, err := NewLocalFileSystem(root)
	if err != nil {
		t.Fatalf("NewLocalFileSystem: %v", err)
	}
	return fs
}

func TestResolve_AllowsRelativePaths(t *testing.T) {
	fs := setupFS(t)
	if err := fs.SaveFile("notes/case1.cln", "pt 40M"); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err := fs.LoadFile("notes/case1.cln")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got != "pt 40M" {
		t.Errorf("content: got %q", got)
	}
}

func TestResolve_RejectsDotDotEscape(t *testing.T) {
	fs := setupFS(t)
	// "../" sequences are collapsed inside the workspace, never landing
	// outside the root.
	_ = fs.SaveFile("../escape.txt", "should never reach parent")
	parentEscape := filepath.Join(filepath.Dir(fs.absRoot), "escape.txt")
	if _, err := os.Stat(parentEscape); err == nil {
		os.Remove(parentEscape)
		t.Fatalf("path traversal succeeded: wrote to %s", parentEscape)
	}
}

func TestResolve_RejectsAbsoluteCallerPath(t *testing.T) {
	fs := setupFS(t)
	var target string
	if runtime.GOOS == "windows" {
		target = `C:\Windows\System32\nope_clinlang.txt`
	} else {
		target = "/tmp/clinlang_escape_test.txt"
	}
	_ = fs.SaveFile(target, "x")
	if _, err := os.Stat(target); err == nil && !strings.HasPrefix(target, fs.absRoot) {
		os.Remove(target)
		t.Fatalf("absolute path traversal succeeded: wrote to %s", target)
	}
}

func TestRootListingWorks(t *testing.T) {
	fs := setupFS(t)
	if err := fs.SaveFile("a.cln", "x"); err != nil {
		t.Fatal(err)
	}
	if err := fs.SaveFile("sub/b.cln", "y"); err != nil {
		t.Fatal(err)
	}
	entries, err := fs.ListFiles(".")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(entries) < 2 {
		t.Errorf("expected at least 2 entries, got %d", len(entries))
	}
	// Confirm forward-slash normalisation.
	for _, e := range entries {
		if strings.Contains(e.Path, "\\") {
			t.Errorf("entry path contains backslash: %q", e.Path)
		}
	}
}

func TestDeleteRefusesRoot(t *testing.T) {
	fs := setupFS(t)
	if err := fs.DeleteFile("."); err == nil {
		t.Errorf("expected error when deleting workspace root, got nil")
	}
}

func TestRenameWorks(t *testing.T) {
	fs := setupFS(t)
	if err := fs.SaveFile("a.cln", "x"); err != nil {
		t.Fatal(err)
	}
	if err := fs.RenameFile("a.cln", "b/c.cln"); err != nil {
		t.Fatalf("rename: %v", err)
	}
	got, err := fs.LoadFile("b/c.cln")
	if err != nil || got != "x" {
		t.Errorf("after rename: got %q err=%v", got, err)
	}
}
