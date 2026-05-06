// Command json2yaml converts a JSON document to YAML, used by `make
// openapi-pull` to vendor the upstream Zenduty / Xurrent IMR OpenAPI spec
// as YAML for human-readable diffs.
//
// Usage:
//
//	json2yaml <input.json> <output.yaml>
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: json2yaml <input.json> <output.yaml>")
		os.Exit(2)
	}
	in, err := os.ReadFile(os.Args[1])
	must(err)

	var doc any
	must(json.Unmarshal(in, &doc))

	out, err := os.Create(os.Args[2])
	must(err)
	defer out.Close()

	enc := yaml.NewEncoder(out)
	enc.SetIndent(2)
	must(enc.Encode(doc))
	must(enc.Close())
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "json2yaml:", err)
		os.Exit(1)
	}
}
