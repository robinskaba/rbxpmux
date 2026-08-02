package editor

import (
	"bytes"
	"fmt"
	"log/slog"

	"github.com/robloxapi/rbxfile/rbxl"
)

type InstructionType byte

const (
	COPY   InstructionType = 1
	REMOVE InstructionType = 2
)

type Instruction struct {
	Type    InstructionType
	Content string
}

func EditPlaces(originFile []byte, targetFiles map[int][]byte, instructions []Instruction) (map[int][]byte, error) {
	result := make(map[int][]byte)
	origin, _, err := rbxl.Decoder{}.Decode(bytes.NewReader(originFile))
	if err != nil {
		return result, fmt.Errorf("decoding origin to rbxfile issue: %w", err)
	}

	for id := range targetFiles {
		slog.Info("copying place", "id", id)
		target, _, err := rbxl.Decoder{}.Decode(bytes.NewReader(targetFiles[id]))
		if err != nil {
			return result, fmt.Errorf("decoding target to rbxfile issue: %w", err)
		}

		err = executePlaceInstructions(origin, target, instructions)
		if err != nil {
			return result, err
		}

		var outBuf bytes.Buffer
		_, err = rbxl.Encoder{}.Encode(&outBuf, target)
		if err != nil {
			return result, err
		}
		result[id] = outBuf.Bytes()
	}
	return result, nil
}
