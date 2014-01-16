package common

import (
	"errors"
	"fmt"

	"github.com/gonum/matrix/mat64"
)

type DataMismatch struct {
	Input  int
	Output int
	Weight int
}

func (d DataMismatch) Error() string {
	return fmt.Sprintf("reggo: length mismatch. inputs: %v, outputs: %v, weights: %v ", d.Input, d.Output, d.Weight)
}

var InputLengths error = errors.New("reggo: inputs do not all have the same length")
var OutputLengths error = errors.New("reggo: outputs do not all have the same length")
var NoData error = errors.New("reggo: no input data")

// Returns weights
func VerifyInputs(inputs, outputs *mat64.Dense, weights []float64) ([]float64, error) {
	nSamples, _ := inputs.Dims()
	nOutputSamples, _ := outputs.Dims()
	nWeights := len(weights)
	if nSamples != nOutputSamples || (nWeights != 0 && nSamples != nWeights) {
		return weights, DataMismatch{
			Input:  nSamples,
			Output: nOutputSamples,
			Weight: nWeights,
		}
	}
	if len(weights) != 0 {
		return weights, nil
	}
	weights = make([]float64, nSamples)
	for i := range weights {
		weights[i] = 1
	}
	return weights, nil
}
