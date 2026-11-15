package editor

import (
	"bytes"
	"fmt"
	"log/slog"

	"github.com/robloxapi/rbxfile"
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

// decodePlace decodes a place file. Decoder warnings mean that some data could
// not be read (e.g. a property type unknown to rbxfile) and would be silently
// dropped when the place is encoded again, so they are treated as errors to
// avoid publishing a damaged place.
func decodePlace(file []byte) (*rbxfile.Root, error) {
	root, warn, err := rbxl.Decoder{}.Decode(bytes.NewReader(file))
	if err != nil {
		return nil, err
	}
	if warn != nil {
		slog.Error("place decoded with warnings", "warnings", warn.Error())
		return nil, fmt.Errorf("place contains data that cannot be read and would be lost by publishing (the place file format may have been changed by a Roblox update): %w", warn)
	}
	return root, nil
}

func EditPlaces(originFile []byte, targetFiles map[int][]byte, instructions []Instruction) (map[int][]byte, error) {
	result := make(map[int][]byte)
	origin, err := decodePlace(originFile)
	if err != nil {
		return result, fmt.Errorf("decoding origin to rbxfile issue: %w", err)
	}

	for id := range targetFiles {
		slog.Info("copying place", "id", id)
		target, err := decodePlace(targetFiles[id])
		if err != nil {
			return result, fmt.Errorf("decoding target %d to rbxfile issue: %w", id, err)
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
