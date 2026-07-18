package source

import "sync"

// FileID is an opaque, process-local identifier for a source file; never
// persist or share it across processes. Use [Registry.URI] for identity
// that must survive serialization.
//
// The zero value is invalid. Valid FileIDs are only produced by
// [Registry.Intern].
type FileID uint32

func (id FileID) IsValid() bool {
	return id != 0
}

// Registry interns URIs into small, comparable FileIDs and provides the
// reverse lookup. Safe for concurrent use; the zero value is not usable,
// use [NewRegistry].
type Registry struct {
	mu    sync.RWMutex
	byURI map[URI]FileID
	byID  []URI // byID[0] is unused; FileID 0 is invalid.
}

func NewRegistry() *Registry {
	return &Registry{
		byURI: make(map[URI]FileID),
		byID:  make([]URI, 1),
	}
}

// Intern returns the FileID for uri, allocating a new one on first use.
// Intern panics if uri is invalid (empty).
func (r *Registry) Intern(uri URI) FileID {
	if !uri.IsValid() {
		panic("source: Intern called with invalid URI")
	}

	r.mu.RLock()
	if id, ok := r.byURI[uri]; ok {
		r.mu.RUnlock()
		return id
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	if id, ok := r.byURI[uri]; ok {
		return id
	}

	id := FileID(len(r.byID)) //nolint:gosec // registries never approach uint32 overflow in practice.
	r.byID = append(r.byID, uri)
	r.byURI[uri] = id

	return id
}

// Lookup returns the FileID already assigned to uri, if any, without
// allocating a new one.
func (r *Registry) Lookup(uri URI) (FileID, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.byURI[uri]

	return id, ok
}

// URI returns the URI that id was interned from.
func (r *Registry) URI(id FileID) (URI, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !id.IsValid() || int(id) >= len(r.byID) {
		return "", false
	}

	return r.byID[id], true
}

func (r *Registry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.byID) - 1
}
