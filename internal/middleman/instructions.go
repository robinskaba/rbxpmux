package middleman

import (
	"bufio"
	"os"
	"strings"
	"unicode/utf8"
)

func mergeInputStrings(copy, remove string) string {
	output := strings.Builder{}

	if copy != "" {
		output.WriteString("copy\n")
		output.WriteString(copy)
		output.WriteString("\n")

		if remove != "" {
			output.WriteString("\n")
		}
	}
	if remove != "" {
		output.WriteString("remove\n")
		output.WriteString(remove)
		output.WriteString("\n")
	}

	return output.String()
}

func parseTXTFile(path string) (string, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	var copy strings.Builder
	var remove strings.Builder
	buf := &copy

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		err = scanner.Err()
		if err != nil {
			return "", "", err
		}
		line := scanner.Text()
		switch strings.ToLower(line) {
		case "copy":
			buf = &copy
		case "remove":
			buf = &remove
		default:
			buf.WriteString(line)
			buf.WriteString("\n")
		}
	}
	return strings.TrimSpace(copy.String()), strings.TrimSpace(remove.String()), nil
}

func InstructionsToString(copyStr, removeStr, title string) string {
	output := strings.Builder{}

	if copyStr != "" {
		output.WriteString("copy\n")
		output.WriteString(copyStr)
		output.WriteString("\n")

		if removeStr != "" {
			output.WriteString("\n")
		}
	}
	if removeStr != "" {
		output.WriteString("remove\n")
		output.WriteString(removeStr)
		output.WriteString("\n")
	}

	return output.String()
}

func BuildNoteDescription(copy, remove string) string {
	const MAX_NOTE_SIZE = 1000

	input := mergeInputStrings(copy, remove)
	out := strings.Builder{}

	// strip input string of comments and whitespaces to save size
	lines := strings.Split(input, "\n")
	for l := range lines {
		line := lines[l]

		// strip comments
		cmtIdx := strings.Index(line, "--")
		if cmtIdx != -1 {
			line = line[:cmtIdx]
		}

		// strip white chars
		line = strings.TrimSuffix(line, "\t")
		line = strings.TrimSuffix(line, " ")

		if line == "" {
			continue
		}

		out.WriteString(line)
		out.WriteString("\n")
	}

	chars := utf8.RuneCountInString(out.String())
	if chars < 1000 {
		return input
	}

	return input[:997] + ".."
}
