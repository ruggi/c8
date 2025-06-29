package emulator

import (
	"sync"

	"github.com/ruggi/c8/internal/display"
	"github.com/ruggi/c8/internal/input"
)

const romStart = 0x200

type Emulator struct {
	memory [4096]uint8
	pc     uint16

	stack [16]uint16
	sp    uint8

	registers [16]uint8

	index uint16

	delayTimer uint8
	soundTimer uint8

	input input.Manager
	fb    display.Framebuffer
	fbMu  sync.RWMutex

	waitingForKey bool
	keyWaitTarget uint8
}

func New(input input.Manager) *Emulator {
	c := &Emulator{
		input: input,
	}

	c.memory = [4096]uint8{
		// seed memory with the font sprites
		0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
		0x20, 0x60, 0x20, 0x20, 0x70, // 1
		0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
		0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
		0x90, 0x90, 0xF0, 0x10, 0x10, // 4
		0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
		0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
		0xF0, 0x10, 0x20, 0x40, 0x40, // 7
		0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
		0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
		0xF0, 0x90, 0xF0, 0x90, 0x90, // A
		0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
		0xF0, 0x80, 0x80, 0x80, 0xF0, // C
		0xE0, 0x90, 0x90, 0x90, 0xE0, // D
		0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
		0xF0, 0x80, 0xF0, 0x80, 0x80, // F
	}

	return c
}

func (c *Emulator) FB() display.Framebuffer {
	c.fbMu.RLock()
	defer c.fbMu.RUnlock()

	return c.fb
}

func (c *Emulator) LoadROM(data []byte) {
	for i := range data {
		c.memory[romStart+i] = data[i]
	}
}

func (c *Emulator) Tick() {
	op1 := c.memory[c.pc]
	op2 := c.memory[c.pc+1]
	c.pcUP()

	opcode := uint16(op1)<<8 | uint16(op2)

	if c.delayTimer > 0 {
		c.delayTimer--
	}
	if c.soundTimer > 0 {
		c.soundTimer--
	}

	ins := parseInstruction(opcode)
	ins.run(c)

}

func (c *Emulator) pcUP() {
	c.pc += 2
}

func (c *Emulator) pcDown() {
	c.pc -= 2
}

// ShouldBuzz returns true if the sound timer is active (greater than 0)
func (c *Emulator) ShouldBuzz() bool {
	return c.soundTimer > 0
}

func (c *Emulator) flag(value bool) {
	if value {
		c.registers[0xF] = 1
	} else {
		c.registers[0xF] = 0
	}
}
