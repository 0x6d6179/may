package ip

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/spf13/cobra"
)

func NewCmdIp(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ip",
		Short: "show local and public ip addresses",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunIp(cmd.Context(), f, cmd)
		},
	}

	cmd.Flags().BoolP("local", "l", false, "show local ip only")
	cmd.Flags().BoolP("public", "p", false, "show public ip only")

	return cmd
}

func RunIp(ctx context.Context, f *factory.Factory, cmd *cobra.Command) error {
	showLocal, _ := cmd.Flags().GetBool("local")
	showPublic, _ := cmd.Flags().GetBool("public")

	if !showLocal && !showPublic {
		showLocal = true
		showPublic = true
	}

	var localIP string
	var publicIP string
	var localErr error
	var publicErr error

	if showLocal {
		localIP, localErr = getLocalIP()
	}

	if showPublic {
		publicIP, publicErr = getPublicIP(ctx)
	}

	if showLocal && showPublic {
		if localErr == nil {
			fmt.Fprintf(f.IO.Out, "local   %s\n", localIP)
		} else {
			fmt.Fprintf(f.IO.Out, "local   error: %v\n", localErr)
		}
		if publicErr == nil {
			fmt.Fprintf(f.IO.Out, "public  %s\n", publicIP)
		} else {
			fmt.Fprintf(f.IO.Out, "public  error: %v\n", publicErr)
		}
	} else if showLocal {
		if localErr != nil {
			return localErr
		}
		fmt.Fprintln(f.IO.Out, localIP)
	} else if showPublic {
		if publicErr != nil {
			return publicErr
		}
		fmt.Fprintln(f.IO.Out, publicIP)
	}

	return nil
}

func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}

		if ipNet.IP.IsLoopback() {
			continue
		}

		if ipv4 := ipNet.IP.To4(); ipv4 != nil {
			return ipv4.String(), nil
		}
	}

	return "", fmt.Errorf("no local ipv4 address found")
}

func getPublicIP(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.ipify.org", nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("api.ipify.org returned status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	ip := strings.TrimSpace(string(data))
	return ip, nil
}
