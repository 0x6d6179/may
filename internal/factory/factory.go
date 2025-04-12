package factory

import "github.com/0x6d6179/may/internal/iostreams"

type Factory struct {
	IO *iostreams.IOStreams
}

func New() *Factory {
	return &Factory{
		IO: iostreams.System(),
	}
}
