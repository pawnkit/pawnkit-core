package source

import (
	"fmt"
	"net/url"
	"strings"
)

// URI is a stable, serializable identifier for a source file, independent
// of any single process. Unlike [FileID], safe to persist, log, or send
// across a wire.
//
// The zero value ("") is invalid.
type URI string

// IsValid reports whether u is nonempty.
func (u URI) IsValid() bool {
	return u != ""
}

func (u URI) String() string {
	return string(u)
}

// FileURI builds a file:// URI from a filesystem path. path may use "/" or
// "\" as a separator; the result always uses "/". Windows drive letters
// (e.g. "C:\foo") are encoded as "file:///C:/foo".
//
// FileURI does not consult the filesystem and does not resolve "." or ".."
// segments.
func FileURI(path string) URI {
	p := strings.ReplaceAll(path, `\`, "/")

	switch {
	case len(p) >= 2 && isDriveLetter(p[0]) && p[1] == ':':
		p = "/" + p
	case !strings.HasPrefix(p, "/"):
		p = "/" + p
	}

	u := url.URL{Scheme: "file", Path: p}

	return URI(u.String())
}

// Filename recovers a filesystem path from a file:// URI.
func (u URI) Filename() (string, error) {
	if !u.IsValid() {
		return "", ErrInvalidURI
	}

	parsed, err := url.Parse(string(u))
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidURI, err)
	}

	if parsed.Scheme != "file" {
		return "", fmt.Errorf("%w: not a file URI: %s", ErrInvalidURI, u)
	}

	p := parsed.Path
	if len(p) >= 3 && p[0] == '/' && isDriveLetter(p[1]) && p[2] == ':' {
		p = p[1:]
	}

	return p, nil
}

func isDriveLetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}
