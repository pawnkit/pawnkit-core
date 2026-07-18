// Package hash provides small, canonical content and configuration hashing
// helpers for building cache keys. Every helper returns a versioned,
// self-describing string (e.g. "sha256:<hex>") so an algorithm change is
// self-describing rather than a silent format break.
package hash
