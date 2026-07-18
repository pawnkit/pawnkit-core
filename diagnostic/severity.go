package diagnostic

import "fmt"

// Severity classifies a diagnostic. Its zero value is invalid.
type Severity uint8

// Diagnostic severity levels.
const (
	SeverityError Severity = iota + 1
	SeverityWarning
	SeverityInfo
	SeverityHint
)

// IsValid reports whether s is a named severity.
func (s Severity) IsValid() bool {
	switch s {
	case SeverityError, SeverityWarning, SeverityInfo, SeverityHint:
		return true
	default:
		return false
	}
}

func (s Severity) String() string {
	switch s {
	case SeverityError:
		return "error"
	case SeverityWarning:
		return "warning"
	case SeverityInfo:
		return "info"
	case SeverityHint:
		return "hint"
	default:
		return fmt.Sprintf("Severity(%d)", uint8(s))
	}
}

// Tag marks a diagnostic with additional machine-readable metadata.
type Tag string

const (
	// TagDeprecated marks deprecated source.
	TagDeprecated Tag = "deprecated"
	// TagUnnecessary marks unnecessary source.
	TagUnnecessary Tag = "unnecessary"
)

// FixKind classifies the review needed for a fix.
type FixKind uint8

// Fix review levels.
const (
	FixSafe FixKind = iota + 1
	FixReviewRequired
)

// IsValid reports whether k is a named fix kind.
func (k FixKind) IsValid() bool {
	switch k {
	case FixSafe, FixReviewRequired:
		return true
	default:
		return false
	}
}

func (k FixKind) String() string {
	switch k {
	case FixSafe:
		return "safe"
	case FixReviewRequired:
		return "review-required"
	default:
		return fmt.Sprintf("FixKind(%d)", uint8(k))
	}
}
