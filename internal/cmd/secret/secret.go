package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/spf13/cobra"
)

func NewCmdSecret(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secret",
		Short: "encrypt or decrypt secrets for safe sharing",
	}

	cmd.AddCommand(newCmdEncrypt(f))
	cmd.AddCommand(newCmdDecrypt(f))

	return cmd
}

func newCmdEncrypt(f *factory.Factory) *cobra.Command {
	var passphrase string
	cmd := &cobra.Command{
		Use:   "encrypt <value>",
		Short: "encrypt a value with a passphrase",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if passphrase == "" {
				return fmt.Errorf("--passphrase is required")
			}

			input, err := readInput(args)
			if err != nil {
				return err
			}

			encrypted, err := encrypt([]byte(input), passphrase)
			if err != nil {
				return fmt.Errorf("encrypt: %w", err)
			}

			fmt.Fprintln(f.IO.Out, encrypted)
			return nil
		},
	}
	cmd.Flags().StringVarP(&passphrase, "passphrase", "p", "", "passphrase for encryption")
	return cmd
}

func newCmdDecrypt(f *factory.Factory) *cobra.Command {
	var passphrase string
	cmd := &cobra.Command{
		Use:   "decrypt <value>",
		Short: "decrypt a value with a passphrase",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if passphrase == "" {
				return fmt.Errorf("--passphrase is required")
			}

			input, err := readInput(args)
			if err != nil {
				return err
			}

			decrypted, err := decrypt(input, passphrase)
			if err != nil {
				return fmt.Errorf("decrypt: %w", err)
			}

			fmt.Fprintln(f.IO.Out, decrypted)
			return nil
		},
	}
	cmd.Flags().StringVarP(&passphrase, "passphrase", "p", "", "passphrase for decryption")
	return cmd
}

func readInput(args []string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("read stdin: %w", err)
		}
		return strings.TrimSpace(string(data)), nil
	}

	return "", fmt.Errorf("provide a value as argument or pipe via stdin")
}

func deriveKey(passphrase string) []byte {
	h := sha256.Sum256([]byte(passphrase))
	return h[:]
}

func encrypt(plaintext []byte, passphrase string) (string, error) {
	block, err := aes.NewCipher(deriveKey(passphrase))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decrypt(encoded string, passphrase string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("invalid base64: %w", err)
	}

	block, err := aes.NewCipher(deriveKey(passphrase))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed — wrong passphrase?")
	}

	return string(plaintext), nil
}
