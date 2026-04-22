package natsort

import (
	"cmp"
	"regexp"
	"strconv"
)

var natsortChunks = regexp.MustCompile(`(\d+|\D+)`)

// Compare returns -1 if x < y according to natural order
func Compare[T ~string](x, y T) int {
	chunksX := natsortChunks.FindAllString(string(x), -1)
	chunksY := natsortChunks.FindAllString(string(y), -1)

	nChunksX := len(chunksX)
	nChunksY := len(chunksY)

	if nChunksX == 0 {
		return cmp.Compare(nChunksX, nChunksY)
	}

	for i := range chunksX {
		if i >= nChunksY {
			return 1
		}

		xInt, aErr := strconv.Atoi(chunksX[i])
		yInt, bErr := strconv.Atoi(chunksY[i])

		// If both chunks are numeric, compare them as integers
		if aErr == nil && bErr == nil {
			if xInt == yInt {

				switch {
				case i == nChunksX-1 && i == nChunksY-1:
					// both sides have the same number of chunks
					return 0
				case i == nChunksX-1:
					// We reached the last chunk of A, thus B is greater than A
					return -1
				case i == nChunksY-1:
					// We reached the last chunk of B, thus A is greater than B
					return 1
				}

				continue
			}
			return cmp.Compare(xInt, yInt)
		}

		// So far both strings are equal, continue to next chunk
		if chunksX[i] == chunksY[i] {
			switch {
			case i == nChunksX-1 && i == nChunksY-1:
				// both sides have the same number of chunks
				return 0
			case i == nChunksX-1:
				// We reached the last chunk of A, thus B is greater than A
				return -1
			case i == nChunksY-1:
				// We reached the last chunk of B, thus A is greater than B
				return 1
			}

			continue
		}

		return cmp.Compare(chunksX[i], chunksY[i])
	}

	return 1
}
