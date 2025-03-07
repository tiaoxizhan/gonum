// Copyright ©2015 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package f64_test

import (
	"math"
	"math/rand/v2"
	"testing"

	"gonum.org/v1/gonum/floats/scalar"
)

const (
	msgVal      = "%v: unexpected value at %v Got: %v Expected: %v"
	msgGuard    = "%v: Guard violated in %s vector %v %v"
	msgReadOnly = "%v: modified read-only %v argument"
)

var (
	nan = math.NaN()
	inf = math.Inf(1)
)

// newGuardedVector allocates a new slice and returns it as three subslices.
// v is a strided vector that contains elements of data at indices i*inc and
// NaN elsewhere. frontGuard and backGuard are filled with NaN values, and
// their backing arrays are directly adjacent to v in memory. The three slices
// can be used to detect invalid memory reads and writes.
func newGuardedVector(data []float64, inc int) (v, frontGuard, backGuard []float64) {
	if inc < 0 {
		inc = -inc
	}
	guard := 2 * inc
	size := (len(data)-1)*inc + 1
	whole := make([]float64, size+2*guard)
	v = whole[guard : len(whole)-guard]
	for i := range whole {
		whole[i] = math.NaN()
	}
	for i, d := range data {
		v[i*inc] = d
	}
	return v, whole[:guard], whole[len(whole)-guard:]
}

// allNaN returns true if x contains only NaN values, and false otherwise.
func allNaN(x []float64) bool {
	for _, v := range x {
		if !math.IsNaN(v) {
			return false
		}
	}
	return true
}

// equalStrided returns true if the strided vector x contains elements of the
// dense vector ref at indices i*inc, false otherwise.
func equalStrided(ref, x []float64, inc int) bool {
	if inc < 0 {
		inc = -inc
	}
	for i, v := range ref {
		if !scalar.Same(x[i*inc], v) {
			return false
		}
	}
	return true
}

// nonStridedWrite returns false if all elements of x at non-stride indices are
// equal to NaN, true otherwise.
func nonStridedWrite(x []float64, inc int) bool {
	if inc < 0 {
		inc = -inc
	}
	for i, v := range x {
		if i%inc != 0 && !math.IsNaN(v) {
			return true
		}
	}
	return false
}

// guardVector copies the source vector (vec) into a new slice with guards.
// Guards guarded[:gdLn] and guarded[len-gdLn:] will be filled with sigil value gdVal.
func guardVector(vec []float64, gdVal float64, gdLn int) (guarded []float64) {
	guarded = make([]float64, len(vec)+gdLn*2)
	copy(guarded[gdLn:], vec)
	for i := 0; i < gdLn; i++ {
		guarded[i] = gdVal
		guarded[len(guarded)-1-i] = gdVal
	}
	return guarded
}

// isValidGuard will test for violated guards, generated by guardVector.
func isValidGuard(vec []float64, gdVal float64, gdLn int) bool {
	for i := 0; i < gdLn; i++ {
		if !scalar.Same(vec[i], gdVal) || !scalar.Same(vec[len(vec)-1-i], gdVal) {
			return false
		}
	}
	return true
}

// guardIncVector copies the source vector (vec) into a new incremented slice with guards.
// End guards will be length gdLen.
// Internal and end guards will be filled with sigil value gdVal.
func guardIncVector(vec []float64, gdVal float64, inc, gdLen int) (guarded []float64) {
	if inc < 0 {
		inc = -inc
	}
	inrLen := len(vec) * inc
	guarded = make([]float64, inrLen+gdLen*2)
	for i := range guarded {
		guarded[i] = gdVal
	}
	for i, v := range vec {
		guarded[gdLen+i*inc] = v
	}
	return guarded
}

// checkValidIncGuard will test for violated guards, generated by guardIncVector
func checkValidIncGuard(t *testing.T, vec []float64, gdVal float64, inc, gdLen int) {
	srcLn := len(vec) - 2*gdLen
	for i := range vec {
		switch {
		case scalar.Same(vec[i], gdVal):
			// Correct value
		case (i-gdLen)%inc == 0 && (i-gdLen)/inc < len(vec):
			// Ignore input values
		case i < gdLen:
			t.Errorf("Front guard violated at %d %v", i, vec[:gdLen])
		case i > gdLen+srcLn:
			t.Errorf("Back guard violated at %d %v", i-gdLen-srcLn, vec[gdLen+srcLn:])
		default:
			t.Errorf("Internal guard violated at %d %v", i-gdLen, vec[gdLen:gdLen+srcLn])
		}
	}
}

// sameApprox tests for nan-aware equality within tolerance.
func sameApprox(a, b, tol float64) bool {
	return scalar.Same(a, b) || scalar.EqualWithinAbsOrRel(a, b, tol, tol)
}

var ( // Offset sets for testing alignment handling in Unitary assembly functions.
	align1 = []int{0, 1}
	align2 = newIncSet(0, 1)
	align3 = newIncToSet(0, 1)
)

type incSet struct {
	x, y int
}

// genInc will generate all (x,y) combinations of the input increment set.
func newIncSet(inc ...int) []incSet {
	n := len(inc)
	is := make([]incSet, n*n)
	for x := range inc {
		for y := range inc {
			is[x*n+y] = incSet{inc[x], inc[y]}
		}
	}
	return is
}

type incToSet struct {
	dst, x, y int
}

// genIncTo will generate all (dst,x,y) combinations of the input increment set.
func newIncToSet(inc ...int) []incToSet {
	n := len(inc)
	is := make([]incToSet, n*n*n)
	for i, dst := range inc {
		for x := range inc {
			for y := range inc {
				is[i*n*n+x*n+y] = incToSet{dst, inc[x], inc[y]}
			}
		}
	}
	return is
}

var benchSink []float64

func randomSlice(n, inc int) []float64 {
	if inc < 0 {
		inc = -inc
	}
	x := make([]float64, (n-1)*inc+1)
	for i := range x {
		x[i] = rand.Float64()
	}
	return x
}

func randSlice(n, inc int, r *rand.Rand) []float64 {
	if inc < 0 {
		inc = -inc
	}
	x := make([]float64, (n-1)*inc+1)
	for i := range x {
		x[i] = r.Float64()
	}
	return x
}
