package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
)

func main() {
	inPath := flag.String("in", "", "input terraform providers schema JSON")
	outPath := flag.String("out", "", "output schema JSON")
	source := flag.String("source", "", "fully-qualified provider source address")
	alias := flag.String("alias", "", "short provider alias expected by tfplugindocs")
	flag.Parse()

	if *inPath == "" || *outPath == "" || *source == "" || *alias == "" {
		fmt.Fprintln(os.Stderr, "usage: docschemaalias -in schema.json -out schema-short.json -source registry.terraform.io/namespace/name -alias name")
		os.Exit(2)
	}

	input, err := os.ReadFile(*inPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read schema: %v\n", err)
		os.Exit(1)
	}

	var top map[string]json.RawMessage
	if err := json.Unmarshal(input, &top); err != nil {
		fmt.Fprintf(os.Stderr, "parse schema: %v\n", err)
		os.Exit(1)
	}

	var providers map[string]json.RawMessage
	if err := json.Unmarshal(top["provider_schemas"], &providers); err != nil {
		fmt.Fprintf(os.Stderr, "parse provider_schemas: %v\n", err)
		os.Exit(1)
	}

	schema, ok := providers[*source]
	if !ok {
		keys := make([]string, 0, len(providers))
		for key := range providers {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		fmt.Fprintf(os.Stderr, "provider schema %q not found; available: %v\n", *source, keys)
		os.Exit(1)
	}
	providers[*alias] = schema

	encodedProviders, err := json.Marshal(providers)
	if err != nil {
		fmt.Fprintf(os.Stderr, "encode provider_schemas: %v\n", err)
		os.Exit(1)
	}
	top["provider_schemas"] = encodedProviders

	output, err := json.MarshalIndent(top, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "encode schema: %v\n", err)
		os.Exit(1)
	}
	output = append(output, '\n')

	if err := os.WriteFile(*outPath, output, 0600); err != nil {
		fmt.Fprintf(os.Stderr, "write schema: %v\n", err)
		os.Exit(1)
	}
}
