package weather

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/spf13/cobra"
)

func NewCmdWeather(f *factory.Factory) *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "weather [city]",
		Short: "show weather forecast",
		RunE: func(cmd *cobra.Command, args []string) error {
			var city string
			if len(args) > 0 {
				city = args[0]
			}

			weather, err := fetchWeather(city, format)
			if err != nil {
				return fmt.Errorf("could not fetch weather: %w", err)
			}

			fmt.Fprintln(f.IO.Out, weather)
			return nil
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "short", "output format: short or full")

	return cmd
}

func fetchWeather(city string, format string) (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	var endpoint string
	if format == "full" {
		if city == "" {
			endpoint = "https://wttr.in/?0ATnq"
		} else {
			endpoint = fmt.Sprintf("https://wttr.in/%s?0ATnq", url.PathEscape(city))
		}
	} else {
		if city == "" {
			endpoint = "https://wttr.in/?format=3"
		} else {
			endpoint = fmt.Sprintf("https://wttr.in/%s?format=3", url.PathEscape(city))
		}
	}

	resp, err := client.Get(endpoint)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
