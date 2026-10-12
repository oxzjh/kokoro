package ort

import (
	"unsafe"

	"github.com/oxzjh/ort"
)

type Kokoro struct {
	engine *ort.Ort
}

func (k *Kokoro) Infer(tokens []int64, styleBlob []float32, speed float32) ([]float32, error) {
	tokensBlob := make([]int64, len(tokens)+2)
	copy(tokensBlob[1:], tokens)
	inputTensors := k.engine.NewInputTensors()
	defer k.engine.FreeInputTensors(inputTensors)

	if err := k.engine.CreateTensorWithShape(unsafe.Pointer(&tokensBlob[0]), len(tokensBlob)<<3, []int64{1, int64(len(tokensBlob))}, ort.TypeInt64, &inputTensors[0]); err != nil {
		return nil, err
	}
	if err := k.engine.CreateTensorWithIndex(unsafe.Pointer(&styleBlob[0]), 1024, 1, ort.TypeFloat, &inputTensors[1]); err != nil {
		return nil, err
	}

	speedBlob := []float32{speed}
	if err := k.engine.CreateTensorWithIndex(unsafe.Pointer(&speedBlob[0]), 4, 2, ort.TypeFloat, &inputTensors[2]); err != nil {
		return nil, err
	}

	result, err := k.engine.Run(inputTensors)
	if err != nil {
		return nil, err
	}
	defer result.Free()

	return result.GetFloat32(0, 0, true)
}

func (k *Kokoro) Close() {
	if k.engine != nil {
		k.engine.Close()
		k.engine = nil
	}
}

func New(model string, threads int) (*Kokoro, error) {
	engine, err := ort.New(model, "kokoro", threads)
	if err != nil {
		return nil, err
	}
	return &Kokoro{engine}, nil
}

func NewWithBuffer(model []byte, threads int) (*Kokoro, error) {
	engine, err := ort.NewWithBuffer(model, "kokoro", threads)
	if err != nil {
		return nil, err
	}
	return &Kokoro{engine}, nil
}
