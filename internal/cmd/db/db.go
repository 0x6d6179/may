package db

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/spf13/cobra"
)

func NewCmdDb(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db [name]",
		Short: "connect to database from .env urls",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			envVars := findDatabaseURLs()
			if len(envVars) == 0 {
				return fmt.Errorf("no database urls found in .env")
			}

			if len(args) == 0 {
				for name, val := range envVars {
					masked := maskURL(val)
					fmt.Fprintf(f.IO.Out, "  %s  %s\n", name, masked)
				}
				return nil
			}

			key := strings.ToUpper(args[0])
			if !strings.HasSuffix(key, "_URL") {
				key += "_URL"
			}

			connStr, ok := envVars[key]
			if !ok {
				for k, v := range envVars {
					if strings.Contains(strings.ToLower(k), strings.ToLower(args[0])) {
						connStr = v
						key = k
						ok = true
						break
					}
				}
			}
			if !ok {
				return fmt.Errorf("no database url matching %q", args[0])
			}

			return connect(f, key, connStr)
		},
	}
	return cmd
}

func findDatabaseURLs() map[string]string {
	result := make(map[string]string)
	envFiles := []string{".env", ".env.local", ".env.development"}

	for _, name := range envFiles {
		path := filepath.Join(".", name)
		f, err := os.Open(path)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			val = strings.Trim(val, `"'`)

			upper := strings.ToUpper(key)
			if strings.Contains(upper, "DATABASE") || strings.Contains(upper, "DB") ||
				strings.Contains(upper, "POSTGRES") || strings.Contains(upper, "MYSQL") ||
				strings.Contains(upper, "MONGO") || strings.Contains(upper, "REDIS") {
				if strings.Contains(val, "://") {
					result[key] = val
				}
			}
		}
		f.Close()
	}

	return result
}

func maskURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return "***"
	}
	if u.User != nil {
		u.User = url.UserPassword(u.User.Username(), "***")
	}
	return u.String()
}

func connect(f *factory.Factory, key, connStr string) error {
	u, err := url.Parse(connStr)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}

	var bin string
	var args []string

	switch u.Scheme {
	case "postgres", "postgresql":
		bin = "psql"
		args = []string{connStr}
	case "mysql":
		bin = "mysql"
		host := u.Hostname()
		port := u.Port()
		if port == "" {
			port = "3306"
		}
		dbName := strings.TrimPrefix(u.Path, "/")
		args = []string{"-h", host, "-P", port, "-u", u.User.Username()}
		if p, ok := u.User.Password(); ok {
			args = append(args, fmt.Sprintf("-p%s", p))
		}
		if dbName != "" {
			args = append(args, dbName)
		}
	case "mongodb", "mongodb+srv":
		bin = "mongosh"
		args = []string{connStr}
	case "redis", "rediss":
		bin = "redis-cli"
		args = []string{"-u", connStr}
	default:
		return fmt.Errorf("unsupported database scheme: %s", u.Scheme)
	}

	if _, err := exec.LookPath(bin); err != nil {
		return fmt.Errorf("%s not found — install it to connect", bin)
	}

	fmt.Fprintf(f.IO.ErrOut, "connecting to %s via %s...\n", key, bin)

	c := exec.Command(bin, args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}
