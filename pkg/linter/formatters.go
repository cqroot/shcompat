package linter

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fatih/color"
)

type Format string

const (
	FormatTTY  Format = "tty"
	FormatJSON Format = "json"
)

func ParseFormat(format string) (Format, error) {
	switch format {
	case "tty":
		return FormatTTY, nil
	case "json":
		return FormatJSON, nil
	default:
		return "", fmt.Errorf("invalid output format: %s", format)
	}
}

func ToTTY(results []CheckResult) string {
	var sb strings.Builder
	for _, r := range results {
		sb.WriteString(color.HiCyanString(r.Rule.Id))
		sb.WriteString(": ")
		sb.WriteString(r.Rule.Error.Error())
		sb.WriteString("\n")

		sb.WriteString("    ")
		sb.WriteString(color.HiBlueString("%s:%d:%d", r.FilePath, r.Line, r.Column))
		sb.WriteString(" ")
		sb.WriteString(strings.TrimSpace(r.Text))

		sb.WriteString("\n\n")
	}
	return sb.String()
}

func ToJSON(results []CheckResult) (string, error) {
	jsonBytes, err := json.Marshal(results)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}
