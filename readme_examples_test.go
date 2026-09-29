package mermaid2d2

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestREADMEExamplesInSync guards the example fences in the README against
// drift: they quote testdata sources and generated D2 verbatim, so a conversion
// change that is not followed by docs/examples/generate.sh would leave the
// README describing output the tool no longer produces.
func TestREADMEExamplesInSync(t *testing.T) {
	readme := readFile(t, "README.md")

	for _, name := range []string{"flowchart", "sequence", "er", "class", "state"} {
		t.Run(name, func(t *testing.T) {
			mmd := readFile(t, filepath.Join("testdata", name+".mmd"))
			d2 := readFile(t, filepath.Join("docs", "examples", name+".d2"))

			if got, err := MermaidToD2(mmd); err != nil {
				t.Fatalf("MermaidToD2(%s.mmd): %v", name, err)
			} else if strings.TrimSpace(got) != d2 {
				t.Errorf("docs/examples/%s.d2 is stale; rerun docs/examples/generate.sh\n got:\n%s\nwant:\n%s", name, got, d2)
			}
			for _, fence := range []string{"```mermaid\n" + mmd + "\n```", "```d2\n" + d2 + "\n```"} {
				if !strings.Contains(readme, fence) {
					t.Errorf("README is missing the current %s fence:\n%s", name, fence)
				}
			}
		})
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return strings.TrimSpace(string(b))
}
