// Command protocol round-trips a diagnostic across a process boundary.
package main

import (
	"encoding/json"
	"fmt"

	"github.com/pawnkit/pawnkit-core/diagnostic"
	"github.com/pawnkit/pawnkit-core/protocol"
	"github.com/pawnkit/pawnkit-core/source"
)

func main() {
	// Producer process.
	producer := source.NewRegistry()
	file := producer.Intern(source.FileURI("/gamemodes/main.pwn"))

	sp, err := source.NewSpan(file, 10, 25)
	if err != nil {
		panic(err)
	}

	d := diagnostic.New("pawnlint:deprecated-func", "pawnlint", diagnostic.SeverityWarning, "SetPlayerColor is deprecated", sp)

	wire, err := protocol.EncodeDiagnostic(protocol.EncodeOptions{Resolver: producer}, d)
	if err != nil {
		panic(err)
	}

	data, err := json.MarshalIndent(wire, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(data))

	// Consumer process.
	var decodedWire protocol.Diagnostic
	if err := json.Unmarshal(data, &decodedWire); err != nil {
		panic(err)
	}

	consumer := source.NewRegistry()

	back, err := protocol.DecodeDiagnostic(consumer, decodedWire)
	if err != nil {
		panic(err)
	}

	uri, _ := consumer.URI(back.Primary.File)
	fmt.Printf("\ndecoded in a different process: %s at %v\n", uri, back.Primary)
}
