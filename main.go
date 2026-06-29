package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/mahveotm/terraform-provider-mtncloud/internal/provider"
)

// Generate provider documentation under docs/ from the schema descriptions and
// the examples/ directory. Run with `go generate ./...` or `make docs`.
//go:generate ./scripts/generate-docs.sh

var version = "dev"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "run provider with debugger support")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/mahveotm/mtncloud",
		Debug:   debug,
	}

	if err := providerserver.Serve(context.Background(), provider.New(version), opts); err != nil {
		log.Fatal(err)
	}
}
