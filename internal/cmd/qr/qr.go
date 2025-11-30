package qr

import (
	"fmt"
	"io"
	"strings"

	"github.com/0x6d6179/may/internal/factory"
	qrcode "github.com/skip2/go-qrcode"
	"github.com/spf13/cobra"
)

func NewCmdQr(f *factory.Factory) *cobra.Command {
	var invert bool

	cmd := &cobra.Command{
		Use:   "qr [text]",
		Short: "generate qr codes",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var text string
			if len(args) > 0 {
				text = args[0]
			} else {
				data, err := io.ReadAll(f.IO.In)
				if err != nil {
					return fmt.Errorf("read stdin: %w", err)
				}
				text = strings.TrimSpace(string(data))
			}

			if text == "" {
				return fmt.Errorf("no input provided")
			}

			q, err := qrcode.New(text, qrcode.Medium)
			if err != nil {
				return fmt.Errorf("generate qr: %w", err)
			}

			fmt.Fprint(f.IO.ErrOut, q.ToSmallString(invert))
			return nil
		},
	}

	cmd.Flags().BoolVar(&invert, "invert", false, "invert colors (for light terminals)")
	return cmd
}
