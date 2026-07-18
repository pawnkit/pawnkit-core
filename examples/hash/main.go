// Command hash demonstrates building a composite cache key from a file's
// content hash, a configuration hash, and a tool version.
package main

import (
	"fmt"

	"github.com/pawnkit/pawnkit-core/hash"
)

type lintConfig struct {
	Ruleset string
	Strict  bool
}

func main() {
	source := []byte("stock Add(a, b) { return a + b; }\n")

	contentKey := hash.Content(source)

	configKey, err := hash.JSON(lintConfig{Ruleset: "recommended", Strict: true})
	if err != nil {
		panic(err)
	}

	cacheKey := hash.Combine(contentKey, configKey, "pawnlint@1.4.0")

	fmt.Printf("content hash: %s\n", contentKey)
	fmt.Printf("config hash:  %s\n", configKey)
	fmt.Printf("cache key:    %s\n", cacheKey)
}
