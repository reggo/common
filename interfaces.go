package common

import (
	"github.com/gonum/matrix/mat64"
)

type Rower interface {
	Row([]float64, int) []float64
}

type RowMatrix interface {
	mat64.Matrix
	Rower
}

type MutableRowMatrix interface {
	Rower
	mat64.Mutable
	SetRow(int, []float64) int
}
