package natsort

import (
	"cmp"
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

		// Custom string type via generics (tested implicitly; direct type test in TestCompareCustomType)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Compare(tt.x, tt.y)
			assert.Equal(t, tt.want, got)
		})
	}
}

type myString string

func TestCompareCustomType(t *testing.T) {
	assert.Equal(t, -1, Compare(myString("file1"), myString("file2")))
	assert.Equal(t, 0, Compare(myString("file1"), myString("file1")))
	assert.Equal(t, 1, Compare(myString("file2"), myString("file1")))
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
