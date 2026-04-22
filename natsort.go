package natsort

func compareNumericChunk[T ~string](x T, xi int, y T, yi int) (result, xEnd, yEnd int) {
	xEnd = xi
	for xEnd < len(x) && x[xEnd] >= '0' && x[xEnd] <= '9' {
		xEnd++
	}
	yEnd = yi
	for yEnd < len(y) && y[yEnd] >= '0' && y[yEnd] <= '9' {
		yEnd++
	}

	xz, yz := xi, yi
	for xz < xEnd && x[xz] == '0' {
		xz++
	}
	for yz < yEnd && y[yz] == '0' {
		yz++
	}

	xSig, ySig := xEnd-xz, yEnd-yz
	if xSig != ySig {
		if xSig < ySig {
			return -1, xEnd, yEnd
		}
		return 1, xEnd, yEnd
	}
	for i := range xSig {
		if x[xz+i] < y[yz+i] {
			return -1, xEnd, yEnd
		}
		if x[xz+i] > y[yz+i] {
			return 1, xEnd, yEnd
		}
	}
	return 0, xEnd, yEnd
}

func compareAlphaChunk[T ~string](x T, xi int, y T, yi int) (result, xEnd, yEnd int) {
	xEnd = xi
	for xEnd < len(x) && (x[xEnd] < '0' || x[xEnd] > '9') {
		xEnd++
	}
	yEnd = yi
	for yEnd < len(y) && (y[yEnd] < '0' || y[yEnd] > '9') {
		yEnd++
	}

	xLen, yLen := xEnd-xi, yEnd-yi
	minLen := min(xLen, yLen)
	for i := range minLen {
		if x[xi+i] < y[yi+i] {
			return -1, xEnd, yEnd
		}
		if x[xi+i] > y[yi+i] {
			return 1, xEnd, yEnd
		}
	}
	if xLen < yLen {
		return -1, xEnd, yEnd
	}
	if xLen > yLen {
		return 1, xEnd, yEnd
	}
	return 0, xEnd, yEnd
}

// Compare returns -1 if x < y, 0 if x == y, and 1 if x > y according to natural order.
// Numeric segments are compared by value (leading zeros ignored), so "01" == "1" and "file9" < "file10".
func Compare[T ~string](x, y T) int {
	xi, yi := 0, 0
	for {
		if xi >= len(x) && yi >= len(y) {
			return 0
		}
		if xi >= len(x) {
			return -1
		}
		if yi >= len(y) {
			return 1
		}

		xDigit := x[xi] >= '0' && x[xi] <= '9'
		yDigit := y[yi] >= '0' && y[yi] <= '9'

		if xDigit && yDigit {
			r, xn, yn := compareNumericChunk(x, xi, y, yi)
			if r != 0 {
				return r
			}
			xi, yi = xn, yn
		} else if !xDigit && !yDigit {
			r, xn, yn := compareAlphaChunk(x, xi, y, yi)
			if r != 0 {
				return r
			}
			xi, yi = xn, yn
		} else {
			if xDigit {
				return -1
			}
			return 1
		}
	}
}
