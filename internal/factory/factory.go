package factory

import (
	"sync"

	"github.com/0x6d6179/may/internal/config"
	"github.com/0x6d6179/may/internal/iostreams"
)

// Factory carries shared dependencies to every command constructor.
// Config is loaded lazily and cached after first call.
type Factory struct {
	IO     *iostreams.IOStreams
	Config func() (*config.Config, error)
}

// New returns a Factory wired to system streams with a lazy config loader.
func New() *Factory {
	var (
		once    sync.Once
		cached  *config.Config
		loadErr error
	)

	return &Factory{
		IO: iostreams.System(),
		Config: func() (*config.Config, error) {
			once.Do(func() {
				cached, loadErr = config.Load()
			})
			return cached, loadErr
		},
	}
}
