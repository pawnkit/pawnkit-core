package diagnostic

import "fmt"

// Zero value is invalid; use one of the named constants.
type Severity uint8

const (
	SeverityError Severity = iota + 1
	SeverityWarning
	SeverityInfo
	SeverityHint
)

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
	TagDeprecated  Tag = "deprecated"
	TagUnnecessary Tag = "unnecessary"
)

// Zero value is invalid; use FixSafe or FixReviewRequired.
type FixKind uint8

const (
	FixSafe FixKind = iota + 1
	FixReviewRequired
)

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
