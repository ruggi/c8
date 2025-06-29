package backend

import (
	"fmt"

	"github.com/ruggi/c8/internal/backend/sdl"
	"github.com/ruggi/c8/internal/backend/terminal"
	"github.com/ruggi/c8/internal/display"
	"github.com/ruggi/c8/internal/input"
	"github.com/ruggi/c8/internal/sound"
)

type Backend interface {
	Name() string
	Update()
	Close()

	display.Manager
	input.Manager
	sound.Manager
}

type Type string

const (
	SDL      Type = "sdl"
	Terminal Type = "terminal"
)

func New(t Type, title string) (Backend, error) {
	switch t {
	case SDL:
		return sdl.New(title)
	case Terminal:
		return terminal.New()
	default:
		return nil, fmt.Errorf("unknown backend type: %s", t)
	}
}
