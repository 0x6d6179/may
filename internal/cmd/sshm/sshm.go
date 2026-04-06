package sshm

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

func NewCmdSshm(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sshm",
		Aliases: []string{"ssh"},
		Short:   "ssh connection manager",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(f)
		},
	}

	cmd.AddCommand(newCmdList(f))
	cmd.AddCommand(newCmdConnect(f))
	cmd.AddCommand(newCmdNew(f))
	cmd.AddCommand(newCmdEdit(f))
	cmd.AddCommand(newCmdDelete(f))
	cmd.AddCommand(newCmdKeygen(f))
	cmd.AddCommand(newCmdCopy(f))

	return cmd
}

func newCmdList(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "list saved connections",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(f)
		},
	}
}

func runList(f *factory.Factory) error {
	s, err := loadStore()
	if err != nil {
		return err
	}

	if len(s.Connections) == 0 {
		fmt.Fprintln(f.IO.ErrOut, "no ssh connections saved — use: may sshm new")
		return nil
	}

	maxName := 0
	for _, c := range s.Connections {
		if len(c.Name) > maxName {
			maxName = len(c.Name)
		}
	}

	for _, c := range s.Connections {
		port := ""
		if c.Port != 0 && c.Port != 22 {
			port = fmt.Sprintf(":%d", c.Port)
		}
		key := ""
		if c.KeyPath != "" {
			key = fmt.Sprintf("  key:%s", filepath.Base(c.KeyPath))
		}
		fmt.Fprintf(f.IO.Out, "  %-*s  %s@%s%s%s\n", maxName, c.Name, c.User, c.Host, port, key)
	}
	return nil
}

func newCmdConnect(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "connect <name>",
		Short: "connect to a saved ssh host",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := loadStore()
			if err != nil {
				return err
			}

			conn, _ := findConnection(s, args[0])
			if conn == nil {
				return fmt.Errorf("connection %q not found", args[0])
			}

			if err := validateConnection(conn); err != nil {
				return fmt.Errorf("invalid connection %q: %w", conn.Name, err)
			}

			sshArgs := buildSSHArgs(conn)
			fmt.Fprintf(f.IO.ErrOut, "connecting to %s (%s@%s)...\n", conn.Name, conn.User, conn.Host)

			c := exec.Command("ssh", sshArgs...)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			return c.Run()
		},
	}
}

func newCmdNew(f *factory.Factory) *cobra.Command {
	var (
		host      string
		user      string
		port      int
		keyPath   string
		proxyJump string
		genKey    bool
	)

	cmd := &cobra.Command{
		Use:   "new <name>",
		Short: "add a new ssh connection",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if host == "" {
				return fmt.Errorf("--host is required")
			}
			if user == "" {
				user = os.Getenv("USER")
				if user == "" {
					user = "root"
				}
			}

			s, err := loadStore()
			if err != nil {
				return err
			}

			if c, _ := findConnection(s, name); c != nil {
				return fmt.Errorf("connection %q already exists — use: may sshm edit %s", name, name)
			}

			if genKey {
				generated, err := generateKey(name)
				if err != nil {
					return fmt.Errorf("generate key: %w", err)
				}
				keyPath = generated
				fmt.Fprintf(f.IO.ErrOut, "✓ generated key: %s\n", keyPath)
			}

			conn := Connection{
				Name:      name,
				Host:      host,
				User:      user,
				Port:      port,
				KeyPath:   keyPath,
				ProxyJump: proxyJump,
			}

			if err := validateConnection(&conn); err != nil {
				return fmt.Errorf("invalid connection: %w", err)
			}

			s.Connections = append(s.Connections, conn)
			if err := saveStore(s); err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "✓ saved connection: %s (%s@%s)\n", name, user, host)
			fmt.Fprintf(f.IO.ErrOut, "  connect with: may sshm connect %s\n", name)
			return nil
		},
	}

	cmd.Flags().StringVarP(&host, "host", "H", "", "hostname or ip address")
	cmd.Flags().StringVarP(&user, "user", "u", "", "ssh username (default: $USER)")
	cmd.Flags().IntVarP(&port, "port", "p", 0, "ssh port (default: 22)")
	cmd.Flags().StringVarP(&keyPath, "key", "k", "", "path to ssh private key")
	cmd.Flags().StringVarP(&proxyJump, "proxy", "J", "", "proxy jump host")
	cmd.Flags().BoolVar(&genKey, "gen-key", false, "generate a new ed25519 key for this connection")

	return cmd
}

func newCmdEdit(f *factory.Factory) *cobra.Command {
	var (
		host      string
		user      string
		port      int
		keyPath   string
		proxyJump string
	)

	cmd := &cobra.Command{
		Use:   "edit <name>",
		Short: "edit an existing connection",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := loadStore()
			if err != nil {
				return err
			}

			conn, _ := findConnection(s, args[0])
			if conn == nil {
				return fmt.Errorf("connection %q not found", args[0])
			}

			if cmd.Flags().Changed("host") {
				conn.Host = host
			}
			if cmd.Flags().Changed("user") {
				conn.User = user
			}
			if cmd.Flags().Changed("port") {
				conn.Port = port
			}
			if cmd.Flags().Changed("key") {
				conn.KeyPath = keyPath
			}
			if cmd.Flags().Changed("proxy") {
				conn.ProxyJump = proxyJump
			}

			if err := validateConnection(conn); err != nil {
				return fmt.Errorf("invalid connection: %w", err)
			}

			if err := saveStore(s); err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "✓ updated connection: %s\n", conn.Name)
			return nil
		},
	}

	cmd.Flags().StringVarP(&host, "host", "H", "", "hostname or ip address")
	cmd.Flags().StringVarP(&user, "user", "u", "", "ssh username")
	cmd.Flags().IntVarP(&port, "port", "p", 0, "ssh port")
	cmd.Flags().StringVarP(&keyPath, "key", "k", "", "path to ssh private key")
	cmd.Flags().StringVarP(&proxyJump, "proxy", "J", "", "proxy jump host")

	return cmd
}

func newCmdDelete(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "delete <name>",
		Aliases: []string{"rm", "remove"},
		Short:   "delete a saved connection",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := loadStore()
			if err != nil {
				return err
			}

			_, idx := findConnection(s, args[0])
			if idx < 0 {
				return fmt.Errorf("connection %q not found", args[0])
			}

			s.Connections = append(s.Connections[:idx], s.Connections[idx+1:]...)
			if err := saveStore(s); err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "✓ deleted connection: %s\n", args[0])
			return nil
		},
	}
}

func newCmdKeygen(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "keygen <name>",
		Short: "generate an ed25519 ssh key pair",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			keyPath, err := generateKey(name)
			if err != nil {
				return err
			}

			pubPath := keyPath + ".pub"
			pubData, err := os.ReadFile(pubPath)
			if err != nil {
				return fmt.Errorf("read public key: %w", err)
			}

			fmt.Fprintf(f.IO.ErrOut, "✓ generated key pair:\n")
			fmt.Fprintf(f.IO.ErrOut, "  private: %s\n", keyPath)
			fmt.Fprintf(f.IO.ErrOut, "  public:  %s\n", pubPath)
			fmt.Fprintf(f.IO.ErrOut, "\npublic key:\n")
			fmt.Fprintf(f.IO.Out, "%s", string(pubData))

			s, err := loadStore()
			if err != nil {
				return nil
			}
			conn, _ := findConnection(s, name)
			if conn != nil && conn.KeyPath == "" {
				conn.KeyPath = keyPath
				_ = saveStore(s)
				fmt.Fprintf(f.IO.ErrOut, "\n✓ attached key to connection: %s\n", name)
			}

			return nil
		},
	}
}

func newCmdCopy(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "copy-id <name>",
		Short: "copy public key to remote host",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := loadStore()
			if err != nil {
				return err
			}

			conn, _ := findConnection(s, args[0])
			if conn == nil {
				return fmt.Errorf("connection %q not found", args[0])
			}

			if conn.KeyPath == "" {
				return fmt.Errorf("no key configured for %q — use: may sshm keygen %s", conn.Name, conn.Name)
			}

			pubPath := conn.KeyPath + ".pub"
			if _, err := os.Stat(pubPath); err != nil {
				return fmt.Errorf("public key not found: %s", pubPath)
			}

			copyArgs := []string{"-i", pubPath}
			if conn.Port != 0 && conn.Port != 22 {
				copyArgs = append(copyArgs, "-p", strconv.Itoa(conn.Port))
			}
			copyArgs = append(copyArgs, fmt.Sprintf("%s@%s", conn.User, conn.Host))

			fmt.Fprintf(f.IO.ErrOut, "copying public key to %s@%s...\n", conn.User, conn.Host)

			c := exec.Command("ssh-copy-id", copyArgs...)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			return c.Run()
		},
	}
}

// validateConnection checks connection fields for unsafe values that could
// be used for SSH argument injection via hand-edited YAML configs.
func validateConnection(conn *Connection) error {
	if conn.Host == "" {
		return fmt.Errorf("host is required")
	}
	if strings.ContainsAny(conn.Host, " \t\n;|&$`") {
		return fmt.Errorf("host contains invalid characters")
	}
	if strings.ContainsAny(conn.User, " \t\n;|&$`") {
		return fmt.Errorf("user contains invalid characters")
	}
	if conn.ProxyJump != "" && strings.ContainsAny(conn.ProxyJump, "\n;|&$`") {
		return fmt.Errorf("proxy jump contains invalid characters")
	}
	if conn.RemoteCmd != "" && strings.ContainsAny(conn.RemoteCmd, "\n;|&$`") {
		return fmt.Errorf("remote command contains invalid characters")
	}
	if conn.Port < 0 || conn.Port > 65535 {
		return fmt.Errorf("port must be between 0 and 65535")
	}
	return nil
}

func buildSSHArgs(conn *Connection) []string {
	var args []string

	if conn.Port != 0 && conn.Port != 22 {
		args = append(args, "-p", strconv.Itoa(conn.Port))
	}
	if conn.KeyPath != "" {
		args = append(args, "-i", conn.KeyPath)
	}
	if conn.ProxyJump != "" {
		args = append(args, "-J", conn.ProxyJump)
	}
	if conn.ExtraFlags != "" {
		args = append(args, strings.Fields(conn.ExtraFlags)...)
	}

	args = append(args, fmt.Sprintf("%s@%s", conn.User, conn.Host))

	if conn.RemoteCmd != "" {
		args = append(args, conn.RemoteCmd)
	}

	return args
}

func generateKey(name string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	keyDir := filepath.Join(home, ".ssh")
	if err := os.MkdirAll(keyDir, 0o700); err != nil {
		return "", err
	}

	keyFile := filepath.Join(keyDir, fmt.Sprintf("may_%s", name))
	if _, err := os.Stat(keyFile); err == nil {
		return keyFile, nil
	}

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", fmt.Errorf("generate ed25519 key: %w", err)
	}

	privBytes, err := ssh.MarshalPrivateKey(priv, "")
	if err != nil {
		return "", fmt.Errorf("marshal private key: %w", err)
	}

	if err := os.WriteFile(keyFile, pem.EncodeToMemory(privBytes), 0o600); err != nil {
		return "", fmt.Errorf("write private key: %w", err)
	}

	pubKey, err := ssh.NewPublicKey(pub)
	if err != nil {
		return "", fmt.Errorf("create public key: %w", err)
	}

	pubBytes := ssh.MarshalAuthorizedKey(pubKey)
	pubFile := keyFile + ".pub"
	if err := os.WriteFile(pubFile, pubBytes, 0o644); err != nil {
		return "", fmt.Errorf("write public key: %w", err)
	}

	return keyFile, nil
}
