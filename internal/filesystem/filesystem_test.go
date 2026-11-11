package filesystem

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveInput(t *testing.T) {
	dir := t.TempDir()
	for _, ext := range []string{".typ"} {
		path := filepath.Join(dir, "source"+ext)
		if err := os.WriteFile(path, []byte("ok"), 0o600); err != nil {
			t.Fatal(err)
		}
		got, err := ResolveInput(path)
		if err != nil {
			t.Fatal(err)
		}
		if !filepath.IsAbs(got) || filepath.Clean(got) != got {
			t.Fatalf("path is not clean and absolute: %q", got)
		}
	}
	for _, path := range []string{"", filepath.Join(dir, "source.txt"), filepath.Join(dir, "source.typst"), filepath.Join(dir, "missing.typ"), dir} {
		if _, err := ResolveInput(path); err == nil {
			t.Errorf("ResolveInput(%q) unexpectedly succeeded", path)
		}
	}
}

func TestPrepareOutput(t *testing.T) {
	dir := t.TempDir()
	got, err := PrepareOutput(filepath.Join(dir, "new", "out.pdf"))
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(got) || filepath.Clean(got) != got {
		t.Fatalf("path is not clean and absolute: %q", got)
	}
	if info, err := os.Stat(filepath.Dir(got)); err != nil || !info.IsDir() {
		t.Fatalf("output directory was not created: %v", err)
	}
	if _, err := PrepareOutput(""); err == nil {
		t.Fatal("empty output path should fail")
	}
	if _, err := PrepareOutput(filepath.Join(dir, "out.txt")); err == nil {
		t.Fatal("unsupported extension should fail")
	}
	if err := os.Mkdir(filepath.Join(dir, "directory.pdf"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := PrepareOutput(filepath.Join(dir, "directory.pdf")); err == nil {
		t.Fatal("directory output should fail")
	}
	parentFile := filepath.Join(dir, "not-a-directory")
	if err := os.WriteFile(parentFile, []byte("file"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := PrepareOutput(filepath.Join(parentFile, "out.pdf")); err == nil {
		t.Fatal("output under a regular file should fail")
	}
}

func TestCreateStagesAndCleanup(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source.typ")
	if err := os.WriteFile(source, []byte("source"), 0o600); err != nil {
		t.Fatal(err)
	}
	template, cleanupTemplate, err := CreateTemplateStage(source)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Dir(template) != dir || !strings.HasSuffix(template, ".typ") {
		t.Fatalf("unexpected template stage path: %q", template)
	}
	if err := os.WriteFile(template, []byte("rendered"), 0o600); err != nil {
		t.Fatal(err)
	}
	cleanupTemplate()
	if _, err := os.Stat(template); !os.IsNotExist(err) {
		t.Fatalf("template stage was not cleaned: %v", err)
	}

	destination := filepath.Join(dir, "out.pdf")
	stage, cleanupOutput, err := CreateOutputStage(destination)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Dir(stage) != dir || !strings.HasSuffix(stage, ".pdf") {
		t.Fatalf("unexpected output stage path: %q", stage)
	}
	if info, err := os.Stat(stage); err != nil || !info.Mode().IsRegular() {
		t.Fatalf("output stage was not reserved for the compiler: %v", err)
	}
	if err := os.WriteFile(stage, []byte("pdf"), 0o600); err != nil {
		t.Fatal(err)
	}
	cleanupOutput()
	if _, err := os.Stat(stage); !os.IsNotExist(err) {
		t.Fatalf("output stage was not cleaned: %v", err)
	}
}

func TestCreateStagesRequireExistingParent(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "missing")
	if _, _, err := CreateTemplateStage(filepath.Join(dir, "source.typ")); err == nil {
		t.Fatal("template staging should fail when its parent does not exist")
	}
	if _, _, err := CreateOutputStage(filepath.Join(dir, "out.pdf")); err == nil {
		t.Fatal("output staging should fail when its parent does not exist")
	}
}
