// Command textedit demonstrates computing a single-file edit and a
// multi-file workspace edit, including the atomic all-or-nothing behaviour
// of ApplyWorkspaceEdit when one document turns out to be stale.
package main

import (
	"errors"
	"fmt"

	"github.com/pawnkit/pawnkit-core/source"
	"github.com/pawnkit/pawnkit-core/textedit"
)

func main() {
	reg := source.NewRegistry()
	main := reg.Intern(source.FileURI("/gamemodes/main.pwn"))
	includeFile := reg.Intern(source.FileURI("/gamemodes/util.inc"))

	mainSnap, err := source.NewSnapshot(main, 1, "SetPlayerColor(playerid, 0xFF0000FF);\n")
	if err != nil {
		panic(err)
	}

	includeSnap, err := source.NewSnapshot(includeFile, 1, "native SetPlayerColor(playerid, color);\n")
	if err != nil {
		panic(err)
	}

	provider := textedit.MapSnapshotProvider{
		main:        mainSnap,
		includeFile: includeSnap,
	}

	rename := func(file source.FileID, oldStart, oldEnd source.Offset, newName string) textedit.DocumentEdit {
		sp, spanErr := source.NewSpan(file, oldStart, oldEnd)
		if spanErr != nil {
			panic(spanErr)
		}

		return textedit.DocumentEdit{
			File:    file,
			Version: 1,
			Edits:   []textedit.Edit{{Span: sp, NewText: newName}},
		}
	}

	we := textedit.WorkspaceEdit{Documents: []textedit.DocumentEdit{
		rename(main, 0, 14, "SetPlayerColour"),
		rename(includeFile, 7, 21, "SetPlayerColour"),
	}}

	results, err := textedit.ApplyWorkspaceEdit(provider, we)
	if err != nil {
		panic(err)
	}

	for _, r := range results {
		uri, _ := reg.URI(r.File)
		fmt.Printf("%s ->\n  %s", uri, r.Content)
	}

	// Now simulate a stale document: main.pwn changed to version 2 behind
	// our back. The whole workspace edit is rejected, not partially
	// applied.
	provider[main], err = source.NewSnapshot(main, 2, mainSnap.Content())
	if err != nil {
		panic(err)
	}

	if _, err := textedit.ApplyWorkspaceEdit(provider, we); errors.Is(err, textedit.ErrStaleDocument) {
		fmt.Println("stale document correctly rejected; util.inc was NOT partially edited")
	}
}
