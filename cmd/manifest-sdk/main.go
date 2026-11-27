package main

import (
	"log"
	"os"

	"brain/internal/plugin"
)

func generate(root, output string) error {
	source, err := plugin.ManifestSDKTypes(root)
	if err != nil {
		return err
	}
	return os.WriteFile(output, []byte(source), 0o644)
}

func main() {
	if err := generate("plugins", "plugins/sdk-typescript/src/manifest-operations.generated.d.ts"); err != nil {
		log.Fatal(err)
	}
}
