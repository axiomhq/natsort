package natsort

import (
	"cmp"
	"fmt"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompare(t *testing.T) {
	tests := []struct {
		name string
		x    string
		y    string
		want int
	}{
		// Equal strings
		{"equal empty", "", "", 0},
		{"equal alpha", "abc", "abc", 0},
		{"equal numeric", "123", "123", 0},
		{"equal mixed", "abc123", "abc123", 0},

		// Pure alphabetic ordering
		{"alpha less", "abc", "abd", -1},
		{"alpha greater", "abd", "abc", 1},
		{"alpha prefix less", "abc", "abcd", -1},
		{"alpha prefix greater", "abcd", "abc", 1},

		// Pure numeric ordering (natural: compare as integers)
		{"num less", "1", "2", -1},
		{"num greater", "2", "1", 1},
		{"num leading zeros equal value", "01", "1", 0},
		{"num leading zeros less", "009", "10", -1},
		{"num large vs small", "100", "9", 1},

		// Mixed alpha+numeric natural ordering
		{"nat less numeric part", "file1", "file2", -1},
		{"nat greater numeric part", "file2", "file1", 1},
		{"nat equal", "file1", "file1", 0},
		{"nat leading zeros", "file01", "file1", 0},
		{"nat large num beats small", "file10", "file9", 1},
		{"nat alpha differs first", "abc1", "abd1", -1},

		// Varying chunk counts
		{"more chunks greater", "a1b", "a1", 1},
		{"fewer chunks less", "a1", "a1b", -1},
		{"numeric then alpha", "1a", "1b", -1},

		// Common filename patterns
		{"version less", "v1.2.3", "v1.2.4", -1},
		{"version greater", "v1.10.0", "v1.9.0", 1},
		{"chapter ordering", "chapter2", "chapter10", -1},
		{"chapter equal", "chapter10", "chapter10", 0},

		// Edge: numeric vs non-numeric chunk at same position
		{"num chunk vs alpha chunk", "1", "a", -1},
		{"alpha chunk vs num chunk", "a", "1", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Compare(tt.x, tt.y)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCompareUnicode(t *testing.T) {
	tests := []struct {
		name string
		x    string
		y    string
		want int
	}{
		// Multi-byte greek letters compared as alpha chunks
		{"greek equal", "αβγ", "αβγ", 0},
		{"greek less", "αβγ", "αβδ", -1}, // γ (U+03B3) < δ (U+03B4), same in byte order
		{"greek greater", "αβδ", "αβγ", 1},

		// Unicode alpha prefix with ASCII digits: natural sort applies to the digit segment
		{"greek prefix numeric less", "α1", "α2", -1},
		{"greek prefix numeric equal", "α1", "α1", 0},
		{"greek prefix natural sort", "α9", "α10", -1}, // 9 < 10 as integers

		// CJK characters with ASCII digits
		{"cjk natural sort less", "文件9", "文件10", -1},
		{"cjk natural sort equal", "文件10", "文件10", 0},

		// 4-byte emoji with ASCII digits
		{"emoji numeric less", "🎉1", "🎉2", -1},
		{"emoji numeric equal", "🎉1", "🎉1", 0},
		{"emoji natural sort", "🎉9", "🎉10", -1},

		// Non-ASCII digit-like characters are treated as alpha (not numeric segments).
		// Arabic-Indic digit ١ (U+0661, bytes 0xD9 0xA1) is not in '0'-'9'.
		// Mixed chunk: ASCII digit "1" (0x31) < non-ASCII "١" (0xD9...) in byte order.
		{"ascii digit vs arabic digit", "1", "١", -1},
		{"arabic digit vs ascii digit", "١", "1", 1},

		// Fullwidth digit １ (U+FF11, bytes 0xEF 0xBC 0x91) is also non-ASCII, treated as alpha.
		{"fullwidth digit vs ascii digit", "１", "1", 1},

		// Unicode non-digit before an ASCII digit segment:
		// "file١" is one alpha chunk (the Arabic digit is non-ASCII so not a chunk boundary),
		// vs "file2" which splits into alpha "file" + digit "2".
		// "file١" alpha chunk is longer than "file" alpha chunk after matching "file" prefix,
		// so "file١" > "file" -> result 1.
		{"unicode absorbs non-ascii digit into alpha chunk", "file١", "file2", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Compare(tt.x, tt.y))
		})
	}
}

func TestCompareEnsureNoAllocs(t *testing.T) {
	cases := []struct{ x, y string }{
		{"abcdefgh", "abcdefgh"},
		{"abcdefgh", "abcdefgi"},
		{"123456", "123456"},
		{"123456", "123457"},
		{"file1", "file2"},
		{"chapter100section200", "chapter100section201"},
		{"v1.10.3-alpha", "v1.10.3-beta"},
	}
	for _, c := range cases {
		allocs := testing.AllocsPerRun(100, func() {
			Compare(c.x, c.y)
		})
		assert.Equal(t, 0.0, allocs, "expected 0 allocs for Compare(%q, %q)", c.x, c.y)
	}
}

type myString string

func TestCompareCustomType(t *testing.T) {
	assert.Equal(t, -1, Compare(myString("file1"), myString("file2")))
	assert.Equal(t, 0, Compare(myString("file1"), myString("file1")))
	assert.Equal(t, 1, Compare(myString("file2"), myString("file1")))
}

func ExampleCompare_sortFunc() {
	files := []string{"file10", "file2", "file1", "file20", "file3"}
	slices.SortFunc(files, Compare)
	fmt.Println(files)
	// Output:
	// [file1 file2 file3 file10 file20]
}

func ExampleCompare_sortFuncVersions() {
	versions := []string{"v1.10.0", "v1.9.0", "v1.2.0", "v1.11.0", "v1.1.0"}
	slices.SortFunc(versions, Compare)
	fmt.Println(versions)
	// Output:
	// [v1.1.0 v1.2.0 v1.9.0 v1.10.0 v1.11.0]
}

func ExampleCompare() {
	fmt.Println(Compare("file9", "file10"))  // numeric: 9 < 10
	fmt.Println(Compare("file10", "file9"))  // numeric: 10 > 9
	fmt.Println(Compare("abc", "abd"))       // alpha: c < d
	fmt.Println(Compare("file1", "file1"))   // equal
	// Output:
	// -1
	// 1
	// -1
	// 0
}

func ExampleCompare_leadingZeros() {
	fmt.Println(Compare("01", "1"))   // equal: leading zeros ignored
	fmt.Println(Compare("009", "10")) // 9 < 10
	// Output:
	// 0
	// -1
}

func ExampleCompare_versions() {
	fmt.Println(Compare("v1.9.0", "v1.10.0"))  // 9 < 10
	fmt.Println(Compare("v1.10.0", "v1.9.0"))  // 10 > 9
	fmt.Println(Compare("v1.2.3", "v1.2.4"))   // 3 < 4
	// Output:
	// -1
	// 1
	// -1
}

func BenchmarkCompare(b *testing.B) {
	cases := []struct {
		name string
		x, y string
	}{
		{"pure alpha equal", "abcdefgh", "abcdefgh"},
		{"pure alpha unequal", "abcdefgh", "abcdefgi"},
		{"pure numeric equal", "123456", "123456"},
		{"pure numeric unequal", "123456", "123457"},
		{"mixed short", "file1", "file2"},
		{"mixed long", "chapter100section200", "chapter100section201"},
		{"version string", "v1.10.3-alpha", "v1.10.3-beta"},
	}

	for _, c := range cases {
		b.Run("natural/"+c.name, func(b *testing.B) {
			for range b.N {
				Compare(c.x, c.y)
			}
		})
		b.Run("std/"+c.name, func(b *testing.B) {
			for range b.N {
				cmp.Compare(c.x, c.y)
			}
		})
	}
}
