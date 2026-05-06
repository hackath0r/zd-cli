//go:build tools

// Package tools tracks build-only tool dependencies as recommended by
// https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
//
// Tools listed here are imported only to keep them in go.sum and pin their
// versions; they are never linked into the final binary.
package tools

import (
	_ "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen"
)
