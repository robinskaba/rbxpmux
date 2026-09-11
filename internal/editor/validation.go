package editor

import (
	"github.com/robloxapi/rbxfile"
)

// extract all references to a map where values are the path to them
func extractReferencePaths(root *rbxfile.Root) map[string]string {
	refMap := make(map[string]string)
	for _, inst := range root.Instances {
		extractReferencePathsRecursive(inst, "", refMap)
	}
	return refMap
}

func extractReferencePathsRecursive(inst *rbxfile.Instance, currentPath string, refMap map[string]string) {
	name := inst.ClassName
	if nameProp, ok := inst.Properties["Name"]; ok {
		name = nameProp.String()
	}

	path := name
	if currentPath != "" {
		path = currentPath + "." + name
	}

	if val, ok := inst.Properties["UniqueId"]; ok {
		if idStr := val.String(); idStr != "" {
			refMap[idStr] = path
		}
	}

	for _, child := range inst.Children {
		extractReferencePathsRecursive(child, path, refMap)
	}
}

// returns true if the ID does not exist in the root
func IsReferenceUnique(root *rbxfile.Root, refID string) bool {
	if refID == "" {
		return true
	}
	for _, inst := range root.Instances {
		if hasReference(inst, refID) {
			return false
		}
	}
	return true
}

func hasReference(inst *rbxfile.Instance, refID string) bool {
	if inst.Reference == refID {
		return true
	}
	for _, child := range inst.Children {
		if hasReference(child, refID) {
			return true
		}
	}
	return false
}

// checks for duplicate references
func CheckDuplicateReferences(incoming *rbxfile.Instance, targetRefs map[string]string, destPath string, isReplace bool) (string, bool) {
	if val, ok := incoming.Properties["UniqueId"]; ok {
		if idStr := val.String(); idStr != "" {
			if existingPath, ok := targetRefs[idStr]; ok {
				isSafe := isReplace && (existingPath == destPath || len(existingPath) > len(destPath) && existingPath[:len(destPath)+1] == destPath+".")
				if !isSafe {
					return existingPath, true
				}
			}
		}
	}
	for _, child := range incoming.Children {
		if dupPath, hasDup := CheckDuplicateReferences(child, targetRefs, destPath, isReplace); hasDup {
			return dupPath, true
		}
	}
	return "", false
}
