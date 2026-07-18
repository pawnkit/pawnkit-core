package textedit_test

import (
	"strings"
	"testing"

	"github.com/pawnkit/pawnkit-core/source"
	"github.com/pawnkit/pawnkit-core/textedit"
)

func BenchmarkApplyManyEdits(b *testing.B) {
	reg := source.NewRegistry()
	file := reg.Intern("file:///bench.pwn")

	var sb strings.Builder
	for range 10_000 {
		sb.WriteString("new Float:pos = 0.0;\n")
	}

	content := sb.String()
	lineLen := len("new Float:pos = 0.0;\n")

	edits := make([]textedit.Edit, 0, 10_000)

	for i := range 10_000 {
		start := source.Offset(i*lineLen + 4) // offset of "Float"
		sp, err := source.NewSpan(file, start, start+5)
		if err != nil {
			b.Fatalf("NewSpan error: %v", err)
		}

		edits = append(edits, textedit.Edit{Span: sp, NewText: "float"})
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		if _, err := textedit.Apply(content, edits); err != nil {
			b.Fatalf("Apply error: %v", err)
		}
	}
}
