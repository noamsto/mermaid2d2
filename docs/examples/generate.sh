#!/usr/bin/env bash
# Regenerate the example D2 sources and SVGs in this directory.
#
# For each fixture it converts the Mermaid source to D2 with m2d2 and renders
# that D2 to an SVG — so what the README shows is a real, reproducible run of
# this tool. --dark-theme bakes a prefers-color-scheme media query into the SVG
# so it follows the reader's theme, and --layout=elk places a composite state's
# terminal node below its container rather than off to one side, as dagre does. The Mermaid side needs no render: GitHub
# draws the ```mermaid fences in the README itself.
#
# Requires: d2 on PATH, or run via `nix run nixpkgs#d2 -- ...`
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo="$(cd "$here/../.." && pwd)"

d2="${D2:-d2}"

for name in flowchart sequence er class state; do
	src="$repo/testdata/$name.mmd"
	go -C "$repo" run ./cmd/m2d2 -to d2 "$src" >"$here/$name.d2"
	"$d2" --layout=elk --dark-theme=200 "$here/$name.d2" "$here/$name.d2.svg"
done
