package b64

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/spf13/cobra"
)

func NewCmdB64(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "b64 [encode|decode] [input]",
		Aliases: []string{"base64"},
		Short:   "base64 encode/decode",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runB64(f, args)
		},
	}
}

func runB64(f *factory.Factory, args []string) error {
	var input string
	var op string

	switch len(args) {
	case 0:
		input, op = readStdin(f)
		if input == "" && op == "" {
			return fmt.Errorf("no input provided")
		}
	case 1:
		input = args[0]
		op = detectOperation(args[0])
	case 2:
		op = args[0]
		input = args[1]
	default:
		return fmt.Errorf("too many arguments")
	}

	if op != "encode" && op != "decode" {
		return fmt.Errorf("invalid operation: %s", op)
	}

	var result string
	var err error

	if op == "encode" {
		result = base64.StdEncoding.EncodeToString([]byte(input))
	} else {
		result, err = decodeBase64(input)
		if err != nil {
			return fmt.Errorf("decode error: %w", err)
		}
	}

	fmt.Fprintln(f.IO.Out, result)
	return nil
}

func readStdin(f *factory.Factory) (string, string) {
	stat, err := os.Stdin.Stat()
	if err != nil || (stat.Mode()&os.ModeCharDevice) != 0 {
		return "", ""
	}

	data, err := io.ReadAll(f.IO.In)
	if err != nil {
		return "", ""
	}

	input := strings.TrimSpace(string(data))
	op := detectOperation(input)
	return input, op
}

func detectOperation(input string) string {
	trimmed := strings.TrimSpace(input)

	if trimmed == "" {
		return "encode"
	}

	_, err := base64.StdEncoding.DecodeString(trimmed)
	if err == nil && len(trimmed) > 0 {
		if isValidBase64String(trimmed) {
			return "decode"
		}
	}

	return "encode"
}

func isValidBase64String(s string) bool {
	if len(s) == 0 {
		return false
	}

	if len(s)%4 != 0 {
		return false
	}

	for _, ch := range s {
		if (ch >= 'A' && ch <= 'Z') ||
			(ch >= 'a' && ch <= 'z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '+' || ch == '/' || ch == '=' {
			continue
		}
		return false
	}

	return true
}

func decodeBase64(input string) (string, error) {
	trimmed := strings.TrimSpace(input)
	result, err := base64.StdEncoding.DecodeString(trimmed)
	if err != nil {
		return "", err
	}

	if !utf8.Valid(result) {
		return "", fmt.Errorf("decoded output is not valid utf-8")
	}

	return string(result), nil
}
