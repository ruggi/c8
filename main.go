package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ruggi/c8/internal/backend"
	"github.com/ruggi/c8/internal/emulator"
	"github.com/urfave/cli"
)

var config struct {
	romFile string
	backend string
}

func main() {
	app := cli.NewApp()
	app.Name = "C8"
	app.Usage = "A humble CHIP-8 emulator"
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "f,rom-file",
			Usage:       "The filename of the Chip-8 ROM to run",
			Destination: &config.romFile,
			Required:    true,
		},
		&cli.StringFlag{
			Name:        "b,backend",
			Usage:       "The backend to use (sdl, terminal)",
			Destination: &config.backend,
			Value:       string(backend.SDL),
		},
	}
	app.Action = run

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func run(ctx *cli.Context) error {
	rom, err := os.ReadFile(config.romFile)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	b, err := backend.New(backend.Type(config.backend), ctx.App.Name)
	if err != nil {
		return fmt.Errorf("error initializing draw: %w", err)
	}
	defer b.Close()

	e := emulator.New(b)
	e.LoadROM(rom)

	cpuHz := 600 // hz
	cpuInterval := time.Duration(1_000_000_000/cpuHz) * time.Nanosecond

	displayHz := 60 // hz
	displayInterval := time.Duration(1_000_000_000/displayHz) * time.Nanosecond

	previousTick := time.Now()
	previousRender := time.Now()

	for {
		now := time.Now()

		b.Update()

		if now.Sub(previousTick) >= cpuInterval {
			e.Tick()
			previousTick = now
		}

		if now.Sub(previousRender) >= displayInterval {
			err := b.Render(e.FB())
			if err != nil {
				return fmt.Errorf("render (%s): %w", b.Name(), err)
			}
			previousRender = now
		}

		// Check if sound should be playing
		if e.ShouldBuzz() {
			err := b.Buzz()
			if err != nil {
				return fmt.Errorf("sound error: %w", err)
			}
		}
	}
}
