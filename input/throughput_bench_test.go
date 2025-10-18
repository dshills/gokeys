package input

import (
	"testing"
)

// BenchmarkParseASCII measures single-byte ASCII character parsing performance.
// This provides a baseline for comparison with multi-byte UTF-8 parsing.
func BenchmarkParseASCII(b *testing.B) {
	parser := NewSequenceParser()
	seq := []byte{'a'}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = parser.Parse(seq)
	}
}

// BenchmarkParseUTF8_2byte measures 2-byte UTF-8 character parsing performance.
func BenchmarkParseUTF8_2byte(b *testing.B) {
	parser := NewSequenceParser()
	seq := []byte{0xc3, 0xa9} // é (U+00E9)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = parser.Parse(seq)
	}
}

// BenchmarkParseUTF8_3byte measures 3-byte UTF-8 character parsing performance.
func BenchmarkParseUTF8_3byte(b *testing.B) {
	parser := NewSequenceParser()
	seq := []byte{0xe6, 0x97, 0xa5} // 日 (U+65E5)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = parser.Parse(seq)
	}
}

// BenchmarkParseUTF8_4byte measures 4-byte UTF-8 character parsing performance.
func BenchmarkParseUTF8_4byte(b *testing.B) {
	parser := NewSequenceParser()
	seq := []byte{0xf0, 0x9f, 0x98, 0x80} // 😀 (U+1F600)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = parser.Parse(seq)
	}
}
