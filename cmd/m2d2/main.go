// Command m2d2 converts between D2 and Mermaid diagram syntax.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/noamsto/mermaid2d2"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "m2d2:", err)
		os.Exit(1)
	}
}

func run() error {
	out := flag.String("o", "", "output file (default stdout)")
	to := flag.String("to", "", "target format: d2 or mermaid (default: inferred from input extension)")
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() > 1 {
		return fmt.Errorf("expected at most one input file, got %d", flag.NArg())
	}
	inPath := flag.Arg(0)

	src, err := readInput(inPath)
	if err != nil {
		return err
	}

	target, err := resolveTarget(*to, inPath)
	if err != nil {
		return err
	}

	var result string
	switch target {
	case "mermaid":
		result, err = mermaid2d2.D2ToMermaid(string(src))
	case "d2":
		result, err = mermaid2d2.MermaidToD2(string(src))
	}
	if err != nil {
		return err
	}

	return writeOutput(*out, result)
}

func readInput(path string) ([]byte, error) {
	if path == "" || path == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(path)
}

// resolveTarget picks the output format from the -to flag, falling back to the
// input file extension.
func resolveTarget(to, inPath string) (string, error) {
	switch to {
	case "d2", "mermaid":
		return to, nil
	case "":
	default:
		return "", fmt.Errorf("invalid -to %q: want d2 or mermaid", to)
	}
	switch strings.ToLower(filepath.Ext(inPath)) {
	case ".d2":
		return "mermaid", nil
	case ".mmd", ".mermaid":
		return "d2", nil
	default:
		return "", fmt.Errorf("cannot infer target from %q; pass -to d2|mermaid", inPath)
	}
}

func writeOutput(path, result string) error {
	if !strings.HasSuffix(result, "\n") {
		result += "\n"
	}
	if path == "" {
		_, err := io.WriteString(os.Stdout, result)
		return err
	}
	return os.WriteFile(path, []byte(result), 0o644)
}

func usage() {
	fmt.Fprint(os.Stderr, `m2d2 — convert between D2 and Mermaid diagram syntax

Usage:
  m2d2 [flags] [input]

Reads the input file, or stdin when input is "-" or omitted. The target format
is inferred from the input extension (.d2 -> mermaid, .mmd/.mermaid -> d2)
unless -to is given.

Flags:
`)
	flag.PrintDefaults()
}
