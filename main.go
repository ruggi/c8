package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ruggi/c8/internal/backend"
	"github.com/ruggi/c8/internal/emulator"
	"github.com/urfave/cli"
)

var config struct {
	romFile    string
	backend    string
	cpuRate    int
	renderRate int
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
		&cli.IntFlag{
			Name:        "c,cpu-rate",
			Usage:       "The CPU rate in Hz",
			Destination: &config.cpuRate,
			Value:       600,
		},
		&cli.IntFlag{
			Name:        "r,render-rate",
			Usage:       "The render rate in Hz",
			Destination: &config.renderRate,
			Value:       60,
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
	e.Load(rom)

	return e.Run(b, config.cpuRate, config.renderRate)
}
