package main

import (
    "os"
    "path/filepath"
    "testing"
)

func TestCLIValidateAndRender(t *testing.T) {
    dir := t.TempDir()
    file := filepath.Join(dir,"hello.sml")
    if err := os.WriteFile(file, []byte(`<screen name="hello"><text>Hello {{name}}</text></screen>`), 0600); err != nil { t.Fatal(err) }
    if got := run([]string{"validate",file}); got != 0 { t.Fatalf("validate exit=%d",got) }
    if got := run([]string{"render","-set","name=developer",file}); got != 0 { t.Fatalf("render exit=%d",got) }
    if got := run([]string{"render","-width","0",file}); got == 0 { t.Fatal("accepted invalid width") }
}

func TestCLIRejectsInvalidMarkup(t *testing.T) {
    dir := t.TempDir()
    file := filepath.Join(dir,"bad.sml")
    if err := os.WriteFile(file, []byte(`<screen name="bad"><execute/></screen>`), 0600); err != nil { t.Fatal(err) }
    if got := run([]string{"validate",file}); got == 0 { t.Fatal("accepted unsupported element") }
}

func TestCLIRenderFailureIsNonzero(t *testing.T) {
    dir := t.TempDir()
    file := filepath.Join(dir, "missing.sml")
    if err := os.WriteFile(file, []byte(`<screen name="missing"><include name="unavailable"/></screen>`), 0600); err != nil { t.Fatal(err) }
    if got := run([]string{"render",file}); got == 0 { t.Fatal("render error reported as success") }
}
