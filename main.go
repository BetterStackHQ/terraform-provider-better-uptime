package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/altinity/terraform-provider-betteruptime/internal/provider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

// Format Terraform examples/.
//go:generate terraform fmt -recursive ./examples/

// Generate Provider Documentation (see https://www.terraform.io/docs/registry/providers/docs.html for details).
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

// Value injected during build process via -ldflags "-X main.version=...".
var version string

func main() {
	var printVersionAndExit bool
	var debugMode bool

	flag.BoolVar(&printVersionAndExit, "version", false, "print version")
	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	if printVersionAndExit {
		fmt.Println(version)
		return
	}

	// Do not include a timestamp (terraform includes it already).
	log.SetFlags(log.Lshortfile)
	// Log request/response bodies (that may contain sensitive data) if (and only if)
	// TF_PROVIDER_BETTERUPTIME_LOG_INSECURE env var is set to 1.
	if os.Getenv("TF_PROVIDER_BETTERUPTIME_LOG_INSECURE") != "1" {
		log.SetOutput(io.Discard)
	}

	opts := &plugin.ServeOpts{ProviderFunc: func() *schema.Provider {
		return provider.New(provider.WithVersion(version))
	}}

	if debugMode {
		err := plugin.Debug(context.Background(), "registry.terraform.io/altinity/betteruptime", opts)
		if err != nil {
			log.Fatal(err.Error())
		}
		return
	}

	plugin.Serve(opts)
}
