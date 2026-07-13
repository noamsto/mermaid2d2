#!/usr/bin/env bash
# Regenerate the side-by-side example SVGs in this directory.
#
# For each fixture it renders the Mermaid source to an SVG, converts the source
# to D2 with m2d2, and renders that D2 to an SVG — so the pair shown in the
# README is a real, reproducible round-trip through this tool.
#
# Requires: mermaid-cli (mmdc) and d2 on PATH, or run via `nix run`:
#   nix run nixpkgs#mermaid-cli -- ... / nix run nixpkgs#d2 -- ...
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo="$(cd "$here/../.." && pwd)"

mmdc="${MMDC:-mmdc}"
d2="${D2:-d2}"

for name in flowchart sequence er; do
	src="$repo/testdata/$name.mmd"
	"$mmdc" -i "$src" -o "$here/$name.mermaid.svg" -b white
	go -C "$repo" run ./cmd/m2d2 -to d2 "$src" >"$here/$name.d2"
	"$d2" "$here/$name.d2" "$here/$name.d2.svg"
done
