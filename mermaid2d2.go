// Package mermaid2d2 converts between Mermaid and D2 diagram syntax.
//
// D2ToMermaid is implemented for D2's node/container/edge graph, which it maps
// onto a Mermaid flowchart. MermaidToD2 is not yet implemented and returns
// ErrNotImplemented.
package mermaid2d2

import "errors"

// ErrNotImplemented is returned by conversion directions that are not built yet.
var ErrNotImplemented = errors.New("mermaid2d2: conversion not implemented")

// MermaidToD2 converts Mermaid source to D2 source.
//
// Not yet implemented; returns ErrNotImplemented.
func MermaidToD2(src string) (string, error) {
	return "", ErrNotImplemented
}
