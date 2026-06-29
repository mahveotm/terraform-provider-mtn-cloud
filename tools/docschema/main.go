package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"

	version "github.com/hashicorp/go-version"
	install "github.com/hashicorp/hc-install"
	"github.com/hashicorp/hc-install/fs"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/hc-install/src"
)

func main() {
	outPath := flag.String("out", "", "output provider schema JSON")
	providerName := flag.String("provider-name", "mtncloud", "short provider name")
	providerSource := flag.String("provider-source", "registry.terraform.io/mahveotm/mtncloud", "fully-qualified provider source address")
	providerVersion := flag.String("provider-version", "0.0.1", "local provider version used for schema extraction")
	terraformVersion := flag.String("terraform-version", "1.5.7", "Terraform version to download when terraform is not already installed")
	flag.Parse()

	if *outPath == "" {
		fmt.Fprintln(os.Stderr, "usage: docschema -out schema.json")
		os.Exit(2)
	}

	tmpDir, err := os.MkdirTemp("", "mtncloud-docs-schema")
	if err != nil {
		exitf("create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	schema, err := buildProviderSchema(tmpDir, *providerName, *providerSource, *providerVersion, *terraformVersion)
	if err != nil {
		exitf("%v", err)
	}
	if err := writeAliasedSchema(*outPath, schema, *providerSource, *providerName); err != nil {
		exitf("%v", err)
	}
}

func buildProviderSchema(tmpDir, providerName, providerSource, providerVersion, terraformVersion string) ([]byte, error) {
	pluginDir := filepath.Join(
		tmpDir,
		"plugins",
		filepath.FromSlash(providerSource),
		providerVersion,
		runtime.GOOS+"_"+runtime.GOARCH,
	)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return nil, fmt.Errorf("create plugin dir: %w", err)
	}

	binary := filepath.Join(pluginDir, "terraform-provider-"+providerName+"_v"+providerVersion)
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	if err := run(exec.Command("go", "build", "-o", binary, ".")); err != nil {
		return nil, fmt.Errorf("build provider: %w", err)
	}

	terraformrc := filepath.Join(tmpDir, "terraformrc")
	if err := os.WriteFile(terraformrc, []byte(fmt.Sprintf(`provider_installation {
  dev_overrides {
    %q = %q
  }
  direct {}
}
`, providerSource, pluginDir)), 0600); err != nil {
		return nil, fmt.Errorf("write terraformrc: %w", err)
	}

	workDir := filepath.Join(tmpDir, "work")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, fmt.Errorf("create terraform work dir: %w", err)
	}
	if err := os.WriteFile(filepath.Join(workDir, "main.tf"), []byte(fmt.Sprintf(`terraform {
  required_providers {
    %[1]s = {
      source  = "mahveotm/%[1]s"
      version = %[2]q
    }
  }
}

provider %[1]q {}
`, providerName, providerVersion)), 0600); err != nil {
		return nil, fmt.Errorf("write terraform config: %w", err)
	}

	tfBin, err := terraformBinary(tmpDir, terraformVersion)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(tfBin, "-chdir="+workDir, "providers", "schema", "-json")
	cmd.Env = append(os.Environ(), "TF_CLI_CONFIG_FILE="+terraformrc)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("terraform providers schema: %w\n%s", err, string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("terraform providers schema: %w", err)
	}
	return output, nil
}

func terraformBinary(tmpDir, terraformVersion string) (string, error) {
	parsedVersion, err := version.NewVersion(terraformVersion)
	if err != nil {
		return "", fmt.Errorf("parse terraform version %q: %w", terraformVersion, err)
	}

	installer := install.NewInstaller()
	tfBin, err := installer.Ensure(context.Background(), []src.Source{
		&fs.AnyVersion{Product: &product.Terraform},
		&releases.ExactVersion{
			InstallDir: tmpDir,
			Product:    product.Terraform,
			Version:    parsedVersion,
		},
	})
	if err != nil {
		return "", fmt.Errorf("find or install terraform: %w", err)
	}
	return tfBin, nil
}

func writeAliasedSchema(outPath string, schemaJSON []byte, providerSource, providerName string) error {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(schemaJSON, &top); err != nil {
		return fmt.Errorf("parse schema: %w", err)
	}

	var providers map[string]json.RawMessage
	if err := json.Unmarshal(top["provider_schemas"], &providers); err != nil {
		return fmt.Errorf("parse provider_schemas: %w", err)
	}

	schema, ok := providers[providerSource]
	if !ok {
		keys := make([]string, 0, len(providers))
		for key := range providers {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		return fmt.Errorf("provider schema %q not found; available: %v", providerSource, keys)
	}
	providers[providerName] = schema

	encodedProviders, err := json.Marshal(providers)
	if err != nil {
		return fmt.Errorf("encode provider_schemas: %w", err)
	}
	top["provider_schemas"] = encodedProviders

	output, err := json.MarshalIndent(top, "", "  ")
	if err != nil {
		return fmt.Errorf("encode schema: %w", err)
	}
	output = append(output, '\n')

	if err := os.WriteFile(outPath, output, 0600); err != nil {
		return fmt.Errorf("write schema: %w", err)
	}
	return nil
}

func run(cmd *exec.Cmd) error {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func exitf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
