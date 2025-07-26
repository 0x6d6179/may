package open

import (
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
	"strings"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/git"
	"github.com/spf13/cobra"
)

func NewCmdOpen(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open [target]",
		Short: "open repository in browser",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !git.IsGitRepo(".") {
				return fmt.Errorf("not in a git repository")
			}

			runner := &git.Runner{}
			repoURL, err := runner.Run("remote", "get-url", "origin")
			if err != nil {
				return fmt.Errorf("get remote url: %w", err)
			}

			browserURL, err := parseRepoURL(repoURL)
			if err != nil {
				return fmt.Errorf("parse repo url: %w", err)
			}

			target := ""
			if len(args) > 0 {
				target = args[0]
			}

			switch target {
			case "pr", "merge":
				browserURL = appendPRPath(browserURL)
			case "branch":
				branch, err := git.CurrentBranch(runner)
				if err != nil {
					return fmt.Errorf("get current branch: %w", err)
				}
				browserURL = appendBranchPath(browserURL, branch)
			case "issues":
				browserURL = appendIssuesPath(browserURL)
			}

			fmt.Fprintf(f.IO.ErrOut, "opening %s\n", browserURL)

			if err := openBrowser(browserURL); err != nil {
				return fmt.Errorf("open browser: %w", err)
			}

			return nil
		},
	}

	return cmd
}

func parseRepoURL(remoteURL string) (string, error) {
	remoteURL = strings.TrimSpace(remoteURL)

	if strings.HasPrefix(remoteURL, "git@") {
		return parseSshURL(remoteURL)
	}

	if strings.HasPrefix(remoteURL, "http") {
		return parseHTTPURL(remoteURL)
	}

	return "", fmt.Errorf("unrecognized remote format: %s", remoteURL)
}

func parseSshURL(sshURL string) (string, error) {
	parts := strings.SplitN(sshURL, ":", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid ssh url")
	}

	hostPath := parts[1]
	hostPath = strings.TrimSuffix(hostPath, ".git")

	host := strings.SplitN(sshURL, "@", 2)
	if len(host) != 2 {
		return "", fmt.Errorf("invalid ssh url")
	}

	domain := strings.SplitN(host[1], ":", 2)[0]

	return fmt.Sprintf("https://%s/%s", domain, hostPath), nil
}

func parseHTTPURL(httpURL string) (string, error) {
	httpURL = strings.TrimSuffix(httpURL, ".git")
	return httpURL, nil
}

func appendPRPath(baseURL string) string {
	if strings.Contains(baseURL, "gitlab.com") {
		return baseURL + "/-/merge_requests"
	}
	if strings.Contains(baseURL, "bitbucket.org") {
		return baseURL + "/pull-requests"
	}
	return baseURL + "/pulls"
}

func appendBranchPath(baseURL, branch string) string {
	return baseURL + "/tree/" + url.PathEscape(branch)
}

func appendIssuesPath(baseURL string) string {
	if strings.Contains(baseURL, "gitlab.com") {
		return baseURL + "/-/issues"
	}
	return baseURL + "/issues"
}

func openBrowser(urlStr string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", urlStr)
	case "linux":
		cmd = exec.Command("xdg-open", urlStr)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", urlStr)
	default:
		return fmt.Errorf("unsupported os: %s", runtime.GOOS)
	}

	return cmd.Run()
}
