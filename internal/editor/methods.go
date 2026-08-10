package editor

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/robloxapi/rbxfile"
)

// wrapper around rbxfile.Instance to provide access to parent instance
type instance struct {
	root   *rbxfile.Root
	parent *rbxfile.Instance
	object *rbxfile.Instance
}

// return array of children for both services and classic instances
func (ins *instance) getChildren() []*rbxfile.Instance {
	if ins.parent == nil {
		return ins.root.Instances
	}
	return ins.parent.Children
}

func (ins *instance) getObjIdx() (int, error) {
	objectIdx := slices.Index(ins.getChildren(), ins.object)
	if objectIdx == -1 {
		return 0, fmt.Errorf("object is not inside the parent") // should not happen at all
	}
	return objectIdx, nil
}

var ErrInstanceNotFound error = fmt.Errorf("specified instance not found")
var ErrServiceRemoval error = fmt.Errorf("services cannot be removed")
var ErrServiceInsert error = fmt.Errorf("services cannot be inserted as they are always present")
var ErrMultipleMatches error = fmt.Errorf("multiple matches of instance found in parent") // instead the instruction should specify to copy the whole parent
var ErrMissingInOrigin error = fmt.Errorf("missing in origin")

func parentInstruction(instruction string) string {
	segments := strings.Split(instruction, ".")
	segments = segments[:len(segments)-1]
	return strings.Join(segments, ".")
}

func findInstance(root *rbxfile.Root, path string) (*instance, error) {
	segments := strings.Split(path, ".")

	matches := []*rbxfile.Instance{}
	children := root.Instances

	for _, segment := range segments {
		var match *rbxfile.Instance
		for i := range children {
			child := children[i]
			val, _ := child.Properties["Name"]
			if val.String() == segment {
				if match != nil {
					return nil, ErrMultipleMatches
				}
				match = child
			}
		}
		if match != nil {
			children = match.Children
			matches = append(matches, match)
		} else {
			return nil, ErrInstanceNotFound
		}
	}

	var parent *rbxfile.Instance
	if len(matches) >= 2 {
		parent = matches[len(matches)-2]
	}
	return &instance{
		object: matches[len(matches)-1],
		parent: parent,
		root:   root,
	}, nil
}

func removeInstance(instance *instance) error {
	if instance.parent == nil {
		return ErrServiceRemoval
	}
	objIdx, err := instance.getObjIdx()
	if err != nil {
		return err
	}
	instance.parent.Children = append(instance.parent.Children[:objIdx], instance.parent.Children[objIdx+1:]...)
	slog.Info("removed instance", "object name", instance.object.Properties["Name"])
	return nil
}

func insertInstance(object *rbxfile.Instance, parent *rbxfile.Instance) error {
	if parent == nil { // if no parent then object is a service since parent does not equal the root(game)
		return ErrServiceInsert
	}
	parent.Children = append(parent.Children, object)
	return nil
}

func replaceInstance(instance *instance, with *rbxfile.Instance) error {
	objIdx, err := instance.getObjIdx()
	if err != nil {
		return err
	}
	children := instance.getChildren()
	oldObj := children[objIdx]
	with.Reference = oldObj.Reference
	children[objIdx] = with
	return nil
}

func performCopy(origin *rbxfile.Root, target *rbxfile.Root, instruction string) error {
	originInstance, err := findInstance(origin, instruction)
	if err != nil {
		if errors.Is(err, ErrInstanceNotFound) {
			return ErrMissingInOrigin
		}
		return fmt.Errorf("origin instance not found: %w", err)
	}

	destinationInstance, err := findInstance(target, instruction)
	if err != nil {
		if errors.Is(err, ErrInstanceNotFound) {
			// INSERT INSTANCE
			destinationParent, err := findInstance(target, parentInstruction(instruction))
			if err != nil {
				return fmt.Errorf("destination parent not found: %w", err)
			}
			err = insertInstance(originInstance.object.Copy(), destinationParent.object)
			if err != nil {
				return fmt.Errorf("failed to insert: %w", err)
			}
			slog.Info("instance inserted", "object name", originInstance.object.Properties["Name"])
		} else {
			return err
		}
	} else {
		// REPLACE INSTANCE
		err := replaceInstance(destinationInstance, originInstance.object.Copy())
		if err != nil {
			return fmt.Errorf("failed to replace: %w", err)
		}
		slog.Info("instance replaced", "object name", originInstance.object.Properties["Name"])
	}
	return nil
}

func performRemove(target *rbxfile.Root, instruction string) error {
	// REMOVE INSTANCE
	instance, err := findInstance(target, instruction)
	if err != nil {
		return err
	}
	err = removeInstance(instance)
	if err != nil {
		return fmt.Errorf("failed to remove: %w", err)
	}
	return nil
}

func executePlaceInstructions(origin *rbxfile.Root, target *rbxfile.Root, instructions []Instruction) error {
	for i := range instructions {
		ins := instructions[i]

		ins.Content = strings.TrimSuffix(ins.Content, ".")
		ins.Content = strings.TrimPrefix(ins.Content, "game.")
		ins.Content = strings.ReplaceAll(ins.Content, "[", ".")
		ins.Content = strings.ReplaceAll(ins.Content, "]", "")
		ins.Content = strings.ReplaceAll(ins.Content, "'", "")
		ins.Content = strings.ReplaceAll(ins.Content, "\"", "")

		if strings.HasPrefix(ins.Content, "workspace") {
			ins.Content = strings.Replace(ins.Content, "workspace", "Workspace", 1)
		}

		switch ins.Type {
		case COPY:
			err := performCopy(origin, target, ins.Content)
			if err != nil {
				return fmt.Errorf("COPY %s: %w", ins.Content, err)
			}
		case REMOVE:
			err := performRemove(target, ins.Content)
			if err != nil {
				return fmt.Errorf("REMOVE %s: %w", ins.Content, err)
			}
		}
	}
	return nil
}
