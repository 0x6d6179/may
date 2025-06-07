package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/version"
	"github.com/spf13/cobra"
)

func NewCmdUpdate(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "update may to the latest release",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(f)
		},
	}
}

func runUpdate(f *factory.Factory) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://api.github.com/repos/0x6d6179/may/releases/latest", nil)
	if err != nil {
		fmt.Fprintf(f.IO.ErrOut, "update check failed: %v\n", err)
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(f.IO.ErrOut, "update check failed: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		fmt.Fprintf(f.IO.ErrOut, "update check failed: %v\n", err)
		return err
	}

	if release.TagName == "" || release.TagName <= version.Version {
		fmt.Fprintf(f.IO.ErrOut, "already up to date\n")
		return nil
	}

	artifactName := fmt.Sprintf("may_%s_%s", runtime.GOOS, runtime.GOARCH)

	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == artifactName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}
	if downloadURL == "" {
		fmt.Fprintf(f.IO.ErrOut, "no release asset for %s\n", artifactName)
		return fmt.Errorf("no release asset for %s", artifactName)
	}

	execPath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(f.IO.ErrOut, "download failed: %v\n", err)
		return err
	}

	tempPath := execPath + ".new"

	if err := downloadFile(ctx, tempPath, downloadURL); err != nil {
		fmt.Fprintf(f.IO.ErrOut, "download failed: %v\n", err)
		_ = os.Remove(tempPath)
		return err
	}

	if err := os.Chmod(tempPath, 0755); err != nil {
		fmt.Fprintf(f.IO.ErrOut, "download failed: %v\n", err)
		_ = os.Remove(tempPath)
		return err
	}

	if err := os.Rename(tempPath, execPath); err != nil {
		fmt.Fprintf(f.IO.ErrOut, "download failed: %v\n", err)
		_ = os.Remove(tempPath)
		return err
	}

	fmt.Fprintf(f.IO.ErrOut, "✓ updated to %s\n", release.TagName)
	return nil
}

func downloadFile(ctx context.Context, dest, url string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %s", resp.Status)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}
