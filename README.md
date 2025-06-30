# C8

This is a humble CHIP-8 emulator made as a lazy Sunday project.

Implementation based off of [this spec](http://devernay.free.fr/hacks/chip8/C8TECH10.HTM#Fx0A) with some slight tweaks for SUPER CHIP. The emulator has been tested against most of the ROMs in https://github.com/Timendus/chip8-test-suite on macOS.

## Run it

```text
go build .
./c8 -r <your-rom-file>
```

Help for the program is available using the `--h (--help)` flag.

## Backend

### SDL

![SDL backend](/_assets/sdl.png)

By default the emulator uses SDL as the input/output backend.
You'll need to have the SDL library installed on your machine in order to compile/run C8.

### Terminal

![Terminal backend](/_assets/terminal.png)

Alternatively you can use the terminal backend:

```text
./c8 -r <your-rom-file> -b terminal
```

Since there are no key release events in terminals, they are simulated with a 250ms timeout.

## Performance

By default the CPU will simulate running at 600Hz, while the rendering will happen at 60Hz.

You can customize both values when launching the program with the `-c` and `-r` flags:

```text
   -c value, --cpu-rate value     The CPU rate in Hz (default: 600)
   -r value, --render-rate value  The render rate in Hz (default: 60)
```
