package hash

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/spf13/cobra"
)

func NewCmdHash(f *factory.Factory) *cobra.Command {
	var (
		algorithm string
		filePath  string
	)

	cmd := &cobra.Command{
		Use:   "hash [string]",
		Short: "hash strings or files",
		RunE: func(cmd *cobra.Command, args []string) error {
			var input []byte
			var err error

			if filePath != "" {
				input, err = os.ReadFile(filePath)
				if err != nil {
					return fmt.Errorf("read file: %w", err)
				}
			} else if len(args) > 0 {
				input = []byte(args[0])
			} else {
				stat, _ := os.Stdin.Stat()
				isPiped := (stat.Mode() & os.ModeCharDevice) == 0
				if isPiped {
					input, err = io.ReadAll(os.Stdin)
					if err != nil {
						return fmt.Errorf("read stdin: %w", err)
					}
				} else {
					return fmt.Errorf("provide input: hash <string>, hash --file <path>, or echo <string> | may hash")
				}
			}

			digest, err := hashInput(input, algorithm)
			if err != nil {
				return err
			}

			fmt.Fprintln(f.IO.Out, digest)
			return nil
		},
	}

	cmd.Flags().StringVarP(&algorithm, "algorithm", "a", "sha256", "hash algorithm: md5, sha1, sha256, sha512")
	cmd.Flags().StringVarP(&filePath, "file", "f", "", "file to hash")

	return cmd
}

func hashInput(input []byte, algorithm string) (string, error) {
	var hash []byte

	switch algorithm {
	case "md5":
		h := md5.Sum(input)
		hash = h[:]
	case "sha1":
		h := sha1.Sum(input)
		hash = h[:]
	case "sha256":
		h := sha256.Sum256(input)
		hash = h[:]
	case "sha512":
		h := sha512.Sum512(input)
		hash = h[:]
	default:
		return "", fmt.Errorf("unsupported algorithm: %s", algorithm)
	}

	return hex.EncodeToString(hash), nil
}
