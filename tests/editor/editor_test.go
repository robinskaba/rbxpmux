package editor_test

import (
	"bytes"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/robinskaba/rbxpmux/internal/editor"
	"github.com/robloxapi/rbxfile"
	"github.com/robloxapi/rbxfile/rbxl"
)

// Helper function to find an instance in a root by its path
func findInstance(root *rbxfile.Root, path string) (*rbxfile.Instance, error) {
	segments := strings.Split(path, ".")
	var current *rbxfile.Instance
	children := root.Instances

	for _, segment := range segments {
		found := false
		for _, child := range children {
			val, _ := child.Properties["Name"]
			if val.String() == segment {
				current = child
				children = current.Children
				found = true
				break
			}
		}
		if !found {
			return nil, errors.New("instance not found")
		}
	}
	return current, nil
}

// Helper to verify two instances are structurally and procedurally identical
func verifyInstancesMatch(t *testing.T, expected, actual *rbxfile.Instance) {
	if expected.ClassName != actual.ClassName {
		t.Errorf("ClassName mismatch: expected %s, got %s", expected.ClassName, actual.ClassName)
	}

	// Compare properties strictly
	for k, expectedVal := range expected.Properties {
		actualVal, ok := actual.Properties[k]
		if !ok {
			t.Errorf("Property %s missing in actual instance", k)
			continue
		}
		// DeepEqual handles different typed values coming from rbxfile properties properly
		if !reflect.DeepEqual(expectedVal, actualVal) {
			t.Errorf("Property %s mismatch: expected %v, got %v", k, expectedVal, actualVal)
		}
	}

	// Ensure children count matches exactly
	if len(expected.Children) != len(actual.Children) {
		t.Errorf("Children count mismatch: expected %d, got %d", len(expected.Children), len(actual.Children))
	} else {
		// Deeply verify children match recursively
		for i, expectedChild := range expected.Children {
			actualChild := actual.Children[i]
			verifyInstancesMatch(t, expectedChild, actualChild)
		}
	}
}

func TestEditPlaces(t *testing.T) {
	originFile, err := os.ReadFile("origin.rbxl")
	if err != nil {
		t.Skipf("origin.rbxl not found in the tests/editor directory, skipping test: %v", err)
	}
	targetFile, err := os.ReadFile("place_a.rbxl")
	if err != nil {
		t.Skipf("place_a.rbxl not found in the tests/editor directory, skipping test: %v", err)
	}

	// Pre-decode origin to serve as our source of truth for 'verify' assertions
	originRoot, _, err := rbxl.Decoder{}.Decode(bytes.NewReader(originFile))
	if err != nil {
		t.Fatalf("failed to decode origin.rbxl: %v", err)
	}

	tests := []struct {
		name         string
		instructions []editor.Instruction
		wantErr      bool
		// A callback to verify the output data has correctly been modified
		verify func(t *testing.T, resultRoot *rbxfile.Root)
	}{
		{
			name: "Remove an existing instance",
			instructions: []editor.Instruction{
				// Pre-requisite: 'FolderToRemove' must exist in place_a.rbxl
				{Type: editor.REMOVE, Content: "ServerStorage.FolderToRemove"},
			},
			wantErr: false,
			verify: func(t *testing.T, resultRoot *rbxfile.Root) {
				_, err := findInstance(resultRoot, "ServerStorage.FolderToRemove")
				if err == nil {
					t.Errorf("expected ServerStorage/FolderToRemove to be removed, but it was found")
				}
			},
		},
		{
			name: "Copy a missing instance (insert) and verify contents match perfectly",
			instructions: []editor.Instruction{
				// Pre-requisite: 'NewFeature' must exist in origin.rbxl but not in place_a.rbxl
				{Type: editor.COPY, Content: "ServerScriptService.NewFeature"},
			},
			wantErr: false,
			verify: func(t *testing.T, resultRoot *rbxfile.Root) {
				actualInstance, err := findInstance(resultRoot, "ServerScriptService.NewFeature")
				if err != nil {
					t.Fatalf("expected ServerScriptService.NewFeature to be inserted, but got error: %v", err)
				}
				expectedInstance, err := findInstance(originRoot, "ServerScriptService.NewFeature")
				if err != nil {
					t.Fatalf("failed to find origin instance for verification: %v", err)
				}
				verifyInstancesMatch(t, expectedInstance, actualInstance)
			},
		},
		{
			name: "Copy an existing instance (replace) and verify contents match perfectly",
			instructions: []editor.Instruction{
				// Pre-requisite: 'FolderToReplace' must exist in BOTH origin.rbxl and place_a.rbxl (with different contents to actually test replacement)
				{Type: editor.COPY, Content: "ServerStorage.FolderToReplace"},
			},
			wantErr: false,
			verify: func(t *testing.T, resultRoot *rbxfile.Root) {
				actualInstance, err := findInstance(resultRoot, "ServerStorage.FolderToReplace")
				if err != nil {
					t.Fatalf("expected ServerStorage.FolderToReplace to exist, but got error: %v", err)
				}
				expectedInstance, err := findInstance(originRoot, "ServerStorage.FolderToReplace")
				if err != nil {
					t.Fatalf("failed to find origin instance for verification: %v", err)
				}
				verifyInstancesMatch(t, expectedInstance, actualInstance)
			},
		},
		{
			name: "Fail to remove non-existent instance",
			instructions: []editor.Instruction{
				{Type: editor.REMOVE, Content: "ServerStorage.NonExistentFolder123"},
			},
			wantErr: true,
		},
		{
			name: "Fail to copy non-existent origin instance",
			instructions: []editor.Instruction{
				{Type: editor.COPY, Content: "ServerStorage.NonExistentOriginFolder123"},
			},
			wantErr: true,
		},
		{
			name: "Fail to copy (insert) into non-existent target parent",
			instructions: []editor.Instruction{
				// Pre-requisite: 'MissingParent/NewChild' exists in origin.rbxl, but 'MissingParent' is absent in place_a.rbxl
				{Type: editor.COPY, Content: "ServerStorage.MissingParent.NewChild"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targetFiles := map[int][]byte{
				1: targetFile,
			}

			results, err := editor.EditPlaces(originFile, targetFiles, tt.instructions)
			if (err != nil) != tt.wantErr {
				t.Errorf("EditPlaces() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.verify != nil {
				resultData, ok := results[1]
				if !ok {
					t.Fatalf("expected result for target 1")
				}

				resultRoot, _, err := rbxl.Decoder{}.Decode(bytes.NewReader(resultData))
				if err != nil {
					t.Fatalf("failed to decode result: %v", err)
				}

				tt.verify(t, resultRoot)
			}
		})
	}
}
