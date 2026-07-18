// Command diagnostic demonstrates building a Diagnostic with a related
// location, a tag, and a safe fix, then validating it.
package main

import (
	"fmt"

	"github.com/pawnkit/pawnkit-core/diagnostic"
	"github.com/pawnkit/pawnkit-core/source"
	"github.com/pawnkit/pawnkit-core/textedit"
)

func main() {
	reg := source.NewRegistry()
	file := reg.Intern(source.FileURI("/gamemodes/main.pwn"))

	primary, err := source.NewSpan(file, 10, 25)
	if err != nil {
		panic(err)
	}

	declaredAt, err := source.NewSpan(file, 0, 5)
	if err != nil {
		panic(err)
	}

	d := diagnostic.New(
		"pawnlint:deprecated-func",
		"pawnlint",
		diagnostic.SeverityWarning,
		"SetPlayerColor is deprecated",
		primary,
	)
	d.Related = []diagnostic.RelatedLocation{{Span: declaredAt, Message: "declared here"}}
	d.Tags = []diagnostic.Tag{diagnostic.TagDeprecated}
	d.Help = "use SetPlayerColour (British spelling) instead"
	d.SafeFixes = []diagnostic.Fix{{
		Message: "Replace with SetPlayerColour",
		Kind:    diagnostic.FixSafe,
		Edit: textedit.WorkspaceEdit{Documents: []textedit.DocumentEdit{
			{File: file, Version: textedit.AnyVersion, Edits: []textedit.Edit{
				{Span: primary, NewText: "SetPlayerColour(playerid, color)"},
			}},
		}},
	}}

	if err := d.Validate(); err != nil {
		panic(err)
	}

	fmt.Printf("[%s] %s: %s\n", d.Severity, d.Code, d.Message)
	fmt.Printf("  deprecated: %v\n", d.HasTag(diagnostic.TagDeprecated))
	fmt.Printf("  safe fixes: %d\n", len(d.SafeFixes))
}
