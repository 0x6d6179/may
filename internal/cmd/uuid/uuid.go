package uuid

import (
	"crypto/rand"
	"fmt"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/spf13/cobra"
)

func NewCmdUuid(f *factory.Factory) *cobra.Command {
	var count int
	var upper bool

	cmd := &cobra.Command{
		Use:   "uuid",
		Short: "generate uuids",
		RunE: func(cmd *cobra.Command, args []string) error {
			for i := 0; i < count; i++ {
				id := generateV4()
				if upper {
					id = toUpper(id)
				}
				fmt.Fprintln(f.IO.Out, id)
			}
			return nil
		},
	}

	cmd.Flags().IntVarP(&count, "count", "n", 1, "number of uuids to generate")
	cmd.Flags().BoolVar(&upper, "upper", false, "output uppercase")

	return cmd
}

func generateV4() string {
	var uuid [16]byte
	_, _ = rand.Read(uuid[:])
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	uuid[8] = (uuid[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])
}

func toUpper(s string) string {
	b := []byte(s)
	for i := range b {
		if b[i] >= 'a' && b[i] <= 'f' {
			b[i] -= 32
		}
	}
	return string(b)
}
