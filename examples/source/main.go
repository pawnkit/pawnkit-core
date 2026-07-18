// Command source demonstrates the basics of the source package: interning
// a file, building a snapshot, and converting between byte offsets and
// line/character positions in UTF-16 (the LSP default).
package main

import (
	"fmt"

	"github.com/pawnkit/pawnkit-core/source"
)

func main() {
	reg := source.NewRegistry()
	file := reg.Intern(source.FileURI("/gamemodes/main.pwn"))

	content := "stock Add(a, b) {\n\treturn a + b;\n}\n"

	snap, err := source.NewSnapshot(file, 1, content)
	if err != nil {
		panic(err)
	}

	// "Add" starts at byte offset 6.
	sp, err := snap.Span(6, 9)
	if err != nil {
		panic(err)
	}

	text, err := snap.Text(sp)
	if err != nil {
		panic(err)
	}

	rng, err := snap.Lines().Range(sp, source.UTF16)
	if err != nil {
		panic(err)
	}

	uri, _ := reg.URI(file)

	fmt.Printf("file:    %s\n", uri)
	fmt.Printf("text:    %q\n", text)
	fmt.Printf("range:   %s (UTF-16 line/character)\n", rng)
}
