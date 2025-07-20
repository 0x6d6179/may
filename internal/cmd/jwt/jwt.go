package jwt

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/spf13/cobra"
)

func NewCmdJwt(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jwt [token]",
		Short: "decode jwt tokens",
		RunE: func(cmd *cobra.Command, args []string) error {
			var token string

			stat, _ := os.Stdin.Stat()
			isPiped := (stat.Mode() & os.ModeCharDevice) == 0

			if isPiped {
				data, err := io.ReadAll(f.IO.In)
				if err != nil {
					return fmt.Errorf("read stdin: %w", err)
				}
				token = strings.TrimSpace(string(data))
			} else if len(args) > 0 {
				token = args[0]
			} else {
				return cmd.Help()
			}

			return decodeJWT(f, token)
		},
	}

	return cmd
}

func decodeJWT(f *factory.Factory, token string) error {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return fmt.Errorf("invalid jwt: expected 3 parts, got %d", len(parts))
	}

	header, err := decodeSegment(parts[0])
	if err != nil {
		return fmt.Errorf("decode header: %w", err)
	}

	payload, err := decodeSegment(parts[1])
	if err != nil {
		return fmt.Errorf("decode payload: %w", err)
	}

	fmt.Fprintf(f.IO.Out, "header:\n")
	if err := printJSON(f, header); err != nil {
		return fmt.Errorf("format header: %w", err)
	}

	fmt.Fprintf(f.IO.Out, "\npayload:\n")
	if err := printJSON(f, payload); err != nil {
		return fmt.Errorf("format payload: %w", err)
	}

	fmt.Fprintf(f.IO.Out, "\n")
	return nil
}

func decodeSegment(seg string) ([]byte, error) {
	data, err := base64.RawURLEncoding.DecodeString(seg)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func printJSON(f *factory.Factory, data []byte) error {
	var obj interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}

	pretty, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return err
	}

	fmt.Fprintf(f.IO.Out, "%s\n", pretty)
	return nil
}
