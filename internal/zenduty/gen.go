// Package zenduty wraps the Zenduty / Xurrent IMR REST API.
//
// The bulk of this package is generated from the upstream OpenAPI spec at
// api/openapi.yaml. The generated file is named zenduty.gen.go and contains
// strongly-typed request/response models and a ClientWithResponses type.
//
// Hand-written supporting code lives alongside the generated file in
// auth.go, retry.go, and client.go, providing:
//
//   - the Authorization: Token <key> header (NOT Bearer)
//   - retry/backoff on 429 and 5xx
//   - a thin New() constructor that wires everything together for the CLI
//
// To regenerate after a spec update:
//
//	make generate
package zenduty

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config ../../api/oapi-codegen.yaml ../../api/openapi.yaml
