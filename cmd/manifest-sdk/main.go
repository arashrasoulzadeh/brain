package main

import (
	"log"
	"os"

	"brain/internal/plugin"
)

func main() {
	source, err := plugin.ManifestSDKTypes("plugins")
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile("plugins/sdk-typescript/src/manifest-operations.generated.ts", []byte(source), 0o644); err != nil {
		log.Fatal(err)
	}
}
