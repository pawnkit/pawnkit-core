package textedit

import (
	"fmt"

	"github.com/pawnkit/pawnkit-core/source"
)

// SnapshotProvider supplies the current [source.Snapshot] for a file.
// [*source.Registry] does not implement this; back it with whatever
// document store the caller already has.
type SnapshotProvider interface {
	Snapshot(file source.FileID) (source.Snapshot, bool)
}

// MapSnapshotProvider is a trivial [SnapshotProvider] backed by a map.
type MapSnapshotProvider map[source.FileID]source.Snapshot

// Snapshot returns the snapshot for file.
func (m MapSnapshotProvider) Snapshot(file source.FileID) (source.Snapshot, bool) {
	s, ok := m[file]

	return s, ok
}

// FileResult is the computed new content for one file after applying its
// DocumentEdit.
type FileResult struct {
	File    source.FileID
	Content string
}

// ApplyWorkspaceEdit previews an atomic multi-document edit.
func ApplyWorkspaceEdit(snapshots SnapshotProvider, we WorkspaceEdit) ([]FileResult, error) {
	if len(we.Documents) == 0 {
		return nil, nil
	}

	seen := make(map[source.FileID]bool, len(we.Documents))
	docSnapshots := make([]source.Snapshot, len(we.Documents))

	for i, doc := range we.Documents {
		if seen[doc.File] {
			return nil, fmt.Errorf("%w: file %v", ErrDuplicateDocument, doc.File)
		}

		seen[doc.File] = true

		snap, ok := snapshots.Snapshot(doc.File)
		if !ok {
			return nil, fmt.Errorf("%w: file %v", ErrUnknownDocument, doc.File)
		}

		if doc.Version != AnyVersion && doc.Version != snap.Version() {
			return nil, fmt.Errorf(
				"%w: file %v: edit computed against version %d, current version is %d",
				ErrStaleDocument, doc.File, doc.Version, snap.Version(),
			)
		}

		if err := ValidateEdits(doc.Edits); err != nil {
			return nil, fmt.Errorf("file %v: %w", doc.File, err)
		}

		docSnapshots[i] = snap
	}

	results := make([]FileResult, len(we.Documents))

	for i, doc := range we.Documents {
		newContent, err := Apply(docSnapshots[i].Content(), doc.Edits)
		if err != nil {
			return nil, fmt.Errorf("file %v: %w", doc.File, err)
		}

		results[i] = FileResult{File: doc.File, Content: newContent}
	}

	return results, nil
}

// ApplyDocumentEdit is a single-file convenience wrapper around
// [ApplyWorkspaceEdit].
func ApplyDocumentEdit(snap source.Snapshot, doc DocumentEdit) (string, error) {
	results, err := ApplyWorkspaceEdit(
		MapSnapshotProvider{doc.File: snap},
		WorkspaceEdit{Documents: []DocumentEdit{doc}},
	)
	if err != nil {
		return "", err
	}

	return results[0].Content, nil
}
