// Command sml validates and renders declarative terminal view files.
package main

import (
    "errors"
    "flag"
    "fmt"
    "os"
    "path/filepath"
    "strconv"
    "strings"

    sml "github.com/AmpedWasTaken/SML"
)

type bindings []string

func (b *bindings) String() string { return strings.Join(*b, ",") }
func (b *bindings) Set(s string) error {
    k, _, ok := strings.Cut(s, "=")
    if !ok || k == "" || strings.ContainsAny(k, " \t\r\n") {
        return errors.New("-set expects key=value")
    }
    *b = append(*b, s)
    return nil
}

func main() { os.Exit(run(os.Args[1:])) }

func usage() {
    fmt.Fprintln(os.Stderr, "Usage: sml <validate|render> [-assets DIR] [-width N] [-set key=value] PATH")
    fmt.Fprintln(os.Stderr, "validate accepts a .sml file or an assets directory; render accepts a .sml file")
}

func run(args []string) int {
    if len(args) < 1 || (args[0] != "validate" && args[0] != "render") {
        usage()
        return 2
    }
    action := args[0]
    fs := flag.NewFlagSet(action, flag.ContinueOnError)
    fs.SetOutput(os.Stderr)
    assets := fs.String("assets", "", "asset root with components/, art/ and media/")
    width := fs.Int("width", 80, "rendered width (12 to 240)")
    var raw bindings
    fs.Var(&raw, "set", "binding key=value (repeatable)")
    if err := fs.Parse(args[1:]); err != nil { return 2 }
    if fs.NArg() != 1 { usage(); return 2 }
    path := fs.Arg(0)
    info, err := os.Stat(path)
    if err != nil { fmt.Fprintln(os.Stderr, "sml:", err); return 1 }
    if info.IsDir() {
        if action != "validate" {
            fmt.Fprintln(os.Stderr, "sml: render requires a .sml file")
            return 2
        }
        views, err := sml.LoadDir(path)
        if err != nil { fmt.Fprintln(os.Stderr, "sml:", err); return 1 }
        fmt.Printf("Validated %d view(s) in %s\n", len(views), path)
        return 0
    }
    if filepath.Ext(path) != ".sml" {
        fmt.Fprintln(os.Stderr, "sml: expected a .sml file")
        return 2
    }
    data, err := os.ReadFile(path)
    if err != nil { fmt.Fprintln(os.Stderr, "sml:", err); return 1 }
    view, err := sml.Parse(data)
    if err != nil { fmt.Fprintln(os.Stderr, "sml:", err); return 1 }
    if *assets != "" {
        if _, err := sml.LoadDir(*assets); err != nil {
            fmt.Fprintln(os.Stderr, "sml: invalid assets:", err)
            return 1
        }
        view = view.WithAssets(*assets)
    }
    if action == "validate" {
        fmt.Printf("Valid SML view: %s\n", view.Name)
        return 0
    }
    if *width < 12 || *width > 240 {
        fmt.Fprintln(os.Stderr, "sml: -width must be between 12 and 240")
        return 2
    }
    values := make(map[string]string, len(raw))
    for _, item := range raw {
        key, value, _ := strings.Cut(item, "=")
        if _, err := strconv.Unquote(value); err == nil && strings.HasPrefix(value, "\"") {
            // Quoted shell arguments can be passed through as one value.
            value, _ = strconv.Unquote(value)
        }
        values[key] = value
    }
    lines, err := view.RenderChecked(values, *width)
    if err != nil { fmt.Fprintln(os.Stderr, "sml:", err); return 1 }
    for _, line := range lines { fmt.Println(line) }
    return 0
}
