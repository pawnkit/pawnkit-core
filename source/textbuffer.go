package source

import (
	"fmt"
	"sort"
	"sync"
	"unicode/utf16"
	"unicode/utf8"
)

const textBufferPieceLimit = 256

type textBlock struct {
	data []byte
}

type textPiece struct {
	block *textBlock
	start int
	end   int
}

// TextBuffer is an immutable, edit-efficient source buffer.
type TextBuffer struct {
	pieces []textPiece
	ends   []int
	length int

	flatOnce sync.Once
	flat     []byte
}

// NewTextBuffer retains content as an immutable source revision.
func NewTextBuffer(content []byte) *TextBuffer {
	block := &textBlock{data: content}
	if len(content) == 0 {
		return &TextBuffer{}
	}
	return newTextBuffer([]textPiece{{block: block, end: len(content)}})
}

// Len returns the source length in bytes.
func (b *TextBuffer) Len() int {
	if b == nil {
		return 0
	}
	return b.length
}

// Apply returns a new revision with one byte range replaced.
func (b *TextBuffer) Apply(start, end Offset, replacement string) (*TextBuffer, error) {
	if b == nil {
		return nil, fmt.Errorf("%w: nil text buffer", ErrInvalidSpan)
	}
	if start < 0 || end < start || int(end) > b.length {
		return nil, fmt.Errorf("%w: replacement range [%d,%d)", ErrInvalidSpan, start, end)
	}
	if !b.validUTF8Boundary(int(start)) || !b.validUTF8Boundary(int(end)) {
		return nil, fmt.Errorf("%w: replacement range [%d,%d)", ErrInvalidUTF8Boundary, start, end)
	}
	if start == end && replacement == "" {
		return b, nil
	}

	pieces := make([]textPiece, 0, len(b.pieces)+3)
	pieces = b.appendRange(pieces, 0, int(start))
	if replacement != "" {
		block := &textBlock{data: []byte(replacement)}
		pieces = appendPiece(pieces, textPiece{block: block, end: len(block.data)})
	}
	pieces = b.appendRange(pieces, int(end), b.length)
	next := newTextBuffer(pieces)
	if len(next.pieces) > textBufferPieceLimit {
		return NewTextBuffer(next.Bytes()), nil
	}
	return next, nil
}

// Bytes returns the complete immutable revision.
// Callers must not modify the returned slice.
func (b *TextBuffer) Bytes() []byte {
	if b == nil || b.length == 0 {
		return nil
	}
	if len(b.pieces) == 1 {
		piece := b.pieces[0]
		return piece.block.data[piece.start:piece.end]
	}
	b.flatOnce.Do(func() {
		b.flat = make([]byte, 0, b.length)
		for _, piece := range b.pieces {
			b.flat = append(b.flat, piece.block.data[piece.start:piece.end]...)
		}
	})
	return b.flat
}

func newTextBuffer(pieces []textPiece) *TextBuffer {
	buffer := &TextBuffer{pieces: pieces, ends: make([]int, len(pieces))}
	for index, piece := range pieces {
		buffer.length += piece.end - piece.start
		buffer.ends[index] = buffer.length
	}
	return buffer
}

func (b *TextBuffer) appendRange(dst []textPiece, start, end int) []textPiece {
	if start == end {
		return dst
	}
	position := 0
	for _, piece := range b.pieces {
		pieceLength := piece.end - piece.start
		pieceEnd := position + pieceLength
		if pieceEnd <= start {
			position = pieceEnd
			continue
		}
		if position >= end {
			break
		}
		localStart := max(start-position, 0)
		localEnd := min(end-position, pieceLength)
		dst = appendPiece(dst, textPiece{
			block: piece.block, start: piece.start + localStart, end: piece.start + localEnd,
		})
		position = pieceEnd
	}
	return dst
}

func appendPiece(pieces []textPiece, piece textPiece) []textPiece {
	if piece.start == piece.end {
		return pieces
	}
	if len(pieces) > 0 {
		last := &pieces[len(pieces)-1]
		if last.block == piece.block && last.end == piece.start {
			last.end = piece.end
			return pieces
		}
	}
	return append(pieces, piece)
}

func (b *TextBuffer) validUTF8Boundary(offset int) bool {
	if offset == b.length {
		return true
	}
	pieceIndex := sort.SearchInts(b.ends, offset+1)
	pieceStart := 0
	if pieceIndex > 0 {
		pieceStart = b.ends[pieceIndex-1]
	}
	piece := b.pieces[pieceIndex]
	value := piece.block.data[piece.start+offset-pieceStart]
	return value&0xC0 != 0x80
}

func (b *TextBuffer) byteAt(offset int) byte {
	pieceIndex := sort.SearchInts(b.ends, offset+1)
	pieceStart := 0
	if pieceIndex > 0 {
		pieceStart = b.ends[pieceIndex-1]
	}
	piece := b.pieces[pieceIndex]
	return piece.block.data[piece.start+offset-pieceStart]
}

func (b *TextBuffer) countCharacters(start, end int, encoding Encoding) int {
	if encoding == UTF8 {
		return end - start
	}
	count := 0
	for start < end {
		r, width := b.decodeRune(start, end)
		start += width
		if encoding == UTF32 {
			count++
			continue
		}
		units := utf16.RuneLen(r)
		if units < 0 {
			units = 1
		}
		count += units
	}
	return count
}

func (b *TextBuffer) advanceCharacters(start, end, count int, encoding Encoding) (int, bool) {
	if count == 0 {
		return 0, true
	}
	if encoding == UTF8 {
		if count > end-start {
			return -1, true
		}
		return count, true
	}

	offset := start
	remaining := count
	for offset < end {
		if remaining == 0 {
			return offset - start, true
		}
		r, width := b.decodeRune(offset, end)
		units := 1
		if encoding == UTF16 {
			units = utf16.RuneLen(r)
			if units < 0 {
				units = 1
			}
			if remaining < units {
				return 0, false
			}
		}
		offset += width
		remaining -= units
	}
	if remaining == 0 {
		return offset - start, true
	}
	return -1, true
}

func (b *TextBuffer) decodeRune(offset, end int) (rune, int) {
	var encoded [utf8.UTFMax]byte
	length := min(end-offset, len(encoded))
	for index := range length {
		encoded[index] = b.byteAt(offset + index)
	}
	return utf8.DecodeRune(encoded[:length])
}
