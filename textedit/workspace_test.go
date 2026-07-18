package textedit_test

import (
	"errors"
	"testing"

	"github.com/pawnkit/pawnkit-core/source"
	"github.com/pawnkit/pawnkit-core/textedit"
)

func mustSnapshot(t *testing.T, file source.FileID, version int32, content string) source.Snapshot {
	t.Helper()

	snap, err := source.NewSnapshot(file, version, content)
	if err != nil {
		t.Fatalf("NewSnapshot error: %v", err)
	}

	return snap
}

func TestApplyWorkspaceEditSingleFile(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f := reg.Intern("file:///a.pwn")

	snap := mustSnapshot(t, f, 1, "hello world")
	provider := textedit.MapSnapshotProvider{f: snap}

	we := textedit.WorkspaceEdit{
		Documents: []textedit.DocumentEdit{
			{File: f, Version: 1, Edits: []textedit.Edit{
				{Span: span(t, f, 6, 11), NewText: "pawn"},
			}},
		},
	}

	results, err := textedit.ApplyWorkspaceEdit(provider, we)
	if err != nil {
		t.Fatalf("ApplyWorkspaceEdit error: %v", err)
	}

	if len(results) != 1 || results[0].Content != "hello pawn" {
		t.Fatalf("results = %+v, want [{%v hello pawn}]", results, f)
	}
}

func TestApplyWorkspaceEditMultiFileTransactional(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f1 := reg.Intern("file:///a.pwn")
	f2 := reg.Intern("file:///b.pwn")

	provider := textedit.MapSnapshotProvider{
		f1: mustSnapshot(t, f1, 1, "aaaa"),
		f2: mustSnapshot(t, f2, 1, "bbbb"),
	}

	we := textedit.WorkspaceEdit{
		Documents: []textedit.DocumentEdit{
			{File: f1, Version: 1, Edits: []textedit.Edit{{Span: span(t, f1, 0, 4), NewText: "AAAA"}}},
			{File: f2, Version: 1, Edits: []textedit.Edit{{Span: span(t, f2, 0, 4), NewText: "BBBB"}}},
		},
	}

	results, err := textedit.ApplyWorkspaceEdit(provider, we)
	if err != nil {
		t.Fatalf("ApplyWorkspaceEdit error: %v", err)
	}

	byFile := map[source.FileID]string{}
	for _, r := range results {
		byFile[r.File] = r.Content
	}

	if byFile[f1] != "AAAA" || byFile[f2] != "BBBB" {
		t.Fatalf("results = %+v", byFile)
	}
}

func TestApplyWorkspaceEditAllOrNothing(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f1 := reg.Intern("file:///a.pwn")
	f2 := reg.Intern("file:///b.pwn")

	provider := textedit.MapSnapshotProvider{
		f1: mustSnapshot(t, f1, 1, "aaaa"),
		f2: mustSnapshot(t, f2, 5, "bbbb"), // current version is 5.
	}

	we := textedit.WorkspaceEdit{
		Documents: []textedit.DocumentEdit{
			{File: f1, Version: 1, Edits: []textedit.Edit{{Span: span(t, f1, 0, 4), NewText: "AAAA"}}},
			{File: f2, Version: 1, Edits: []textedit.Edit{{Span: span(t, f2, 0, 4), NewText: "BBBB"}}}, // stale.
		},
	}

	results, err := textedit.ApplyWorkspaceEdit(provider, we)
	if !errors.Is(err, textedit.ErrStaleDocument) {
		t.Fatalf("err = %v, want ErrStaleDocument", err)
	}

	if results != nil {
		t.Fatalf("expected nil results on failure, got %+v (f1 must not be silently applied)", results)
	}
}

func TestApplyWorkspaceEditRejectsUnknownDocument(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f := reg.Intern("file:///a.pwn")

	provider := textedit.MapSnapshotProvider{}

	we := textedit.WorkspaceEdit{
		Documents: []textedit.DocumentEdit{
			{File: f, Version: textedit.AnyVersion, Edits: []textedit.Edit{{Span: span(t, f, 0, 0), NewText: "x"}}},
		},
	}

	if _, err := textedit.ApplyWorkspaceEdit(provider, we); !errors.Is(err, textedit.ErrUnknownDocument) {
		t.Fatalf("err = %v, want ErrUnknownDocument", err)
	}
}

func TestApplyWorkspaceEditRejectsDuplicateDocument(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f := reg.Intern("file:///a.pwn")

	snap := mustSnapshot(t, f, 1, "content")
	provider := textedit.MapSnapshotProvider{f: snap}

	we := textedit.WorkspaceEdit{
		Documents: []textedit.DocumentEdit{
			{File: f, Version: 1, Edits: []textedit.Edit{{Span: span(t, f, 0, 0), NewText: "a"}}},
			{File: f, Version: 1, Edits: []textedit.Edit{{Span: span(t, f, 1, 1), NewText: "b"}}},
		},
	}

	if _, err := textedit.ApplyWorkspaceEdit(provider, we); !errors.Is(err, textedit.ErrDuplicateDocument) {
		t.Fatalf("err = %v, want ErrDuplicateDocument", err)
	}
}

func TestApplyWorkspaceEditAnyVersionSkipsStaleCheck(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f := reg.Intern("file:///a.pwn")

	snap := mustSnapshot(t, f, 42, "content")
	provider := textedit.MapSnapshotProvider{f: snap}

	we := textedit.WorkspaceEdit{
		Documents: []textedit.DocumentEdit{
			{File: f, Version: textedit.AnyVersion, Edits: []textedit.Edit{{Span: span(t, f, 0, 7), NewText: "changed"}}},
		},
	}

	results, err := textedit.ApplyWorkspaceEdit(provider, we)
	if err != nil {
		t.Fatalf("ApplyWorkspaceEdit error: %v", err)
	}

	if results[0].Content != "changed" {
		t.Fatalf("Content = %q, want %q", results[0].Content, "changed")
	}
}

func TestApplyWorkspaceEditEmpty(t *testing.T) {
	t.Parallel()

	results, err := textedit.ApplyWorkspaceEdit(textedit.MapSnapshotProvider{}, textedit.WorkspaceEdit{})
	if err != nil || results != nil {
		t.Fatalf("ApplyWorkspaceEdit(empty) = (%+v, %v), want (nil, nil)", results, err)
	}
}

func TestApplyDocumentEdit(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f := reg.Intern("file:///a.pwn")

	snap := mustSnapshot(t, f, 1, "hello world")

	got, err := textedit.ApplyDocumentEdit(snap, textedit.DocumentEdit{
		File:    f,
		Version: 1,
		Edits:   []textedit.Edit{{Span: span(t, f, 0, 5), NewText: "howdy"}},
	})
	if err != nil {
		t.Fatalf("ApplyDocumentEdit error: %v", err)
	}

	if got != "howdy world" {
		t.Fatalf("got %q", got)
	}
}

func TestApplyWorkspaceEditPreviewDoesNotMutateSnapshot(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f := reg.Intern("file:///a.pwn")

	snap := mustSnapshot(t, f, 1, "original")
	provider := textedit.MapSnapshotProvider{f: snap}

	we := textedit.WorkspaceEdit{
		Documents: []textedit.DocumentEdit{
			{File: f, Version: 1, Edits: []textedit.Edit{{Span: span(t, f, 0, 8), NewText: "modified"}}},
		},
	}

	if _, err := textedit.ApplyWorkspaceEdit(provider, we); err != nil {
		t.Fatalf("preview error: %v", err)
	}

	// The provider's snapshot must remain untouched: this package never
	// mutates or persists, only computes.
	again, _ := provider.Snapshot(f)
	if again.Content() != "original" {
		t.Fatalf("snapshot content changed after preview: %q", again.Content())
	}
}
