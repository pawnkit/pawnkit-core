package source_test

import (
	"strings"
	"testing"

	"github.com/pawnkit/pawnkit-core/source"
)

func largeSyntheticFile(lines int) string {
	var b strings.Builder

	pattern := []string{
		"stock OnPlayerConnect(playerid) {\n",
		"\treturn 1;\r\n",
		"} // done\n",
		"// commentaire en français: café, naïve, résumé\n",
		"new Float:pos[3];\n",
	}

	for i := range lines {
		b.WriteString(pattern[i%len(pattern)])
	}

	return b.String()
}

func BenchmarkNewLineIndex(b *testing.B) {
	content := largeSyntheticFile(50_000)
	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		_ = source.NewLineIndex(content)
	}
}

func BenchmarkLineIndexApply(b *testing.B) {
	content := largeSyntheticFile(50_000)
	index := source.NewLineIndex(content)
	offset := source.Offset(len(content) / 2)
	b.ReportAllocs()
	b.SetBytes(int64(len(content)))
	b.ResetTimer()

	for b.Loop() {
		if _, err := index.Apply(offset, offset, "x"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPositionUTF16(b *testing.B) {
	content := largeSyntheticFile(50_000)
	idx := source.NewLineIndex(content)

	offsets := make([]source.Offset, 0, 1000)
	for i := range 1000 {
		offsets = append(offsets, source.Offset(i*len(content)/1000))
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; b.Loop(); i++ {
		off := offsets[i%len(offsets)]
		if !idx.ValidOffset(off) {
			continue
		}

		if _, err := idx.Position(off, source.UTF16); err != nil {
			b.Fatalf("Position error: %v", err)
		}
	}
}

func BenchmarkPositionUTF8(b *testing.B) {
	content := largeSyntheticFile(50_000)
	idx := source.NewLineIndex(content)

	offsets := make([]source.Offset, 0, 1000)
	for i := range 1000 {
		offsets = append(offsets, source.Offset(i*len(content)/1000))
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; b.Loop(); i++ {
		off := offsets[i%len(offsets)]
		if !idx.ValidOffset(off) {
			continue
		}

		if _, err := idx.Position(off, source.UTF8); err != nil {
			b.Fatalf("Position error: %v", err)
		}
	}
}

func BenchmarkLineAt(b *testing.B) {
	content := largeSyntheticFile(50_000)
	idx := source.NewLineIndex(content)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; b.Loop(); i++ {
		off := source.Offset((i * 97) % len(content))
		if !idx.ValidOffset(off) {
			continue
		}

		if _, err := idx.LineAt(off); err != nil {
			b.Fatalf("LineAt error: %v", err)
		}
	}
}
