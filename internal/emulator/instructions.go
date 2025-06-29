package emulator

import (
	"math/rand"

	"github.com/ruggi/c8/internal/display"
)

type instruction interface {
	run(c *Emulator)
}

type instructionInput struct {
	opcode uint16
	lsb    uint8
	nnn    uint16
	nn     uint8
	n      uint8
	x      uint8
	y      uint8
}

func parseInstruction(ins uint16) instruction {
	in := &instructionInput{
		opcode: ins >> 12,
		lsb:    uint8(ins & 0x000F),
		nnn:    ins & 0x0FFF,
		nn:     uint8(ins & 0x00FF),
		n:      uint8(ins & 0x000F),
		x:      uint8((ins & 0x0F00) >> 8),
		y:      uint8((ins & 0x00F0) >> 4),
	}

	switch in.opcode {
	case 0x0:
		switch in.nnn {
		case 0x0E0:
			return op00E0{}
		case 0x0EE:
			return op00EE{}
		default:
			return op0NNN{in: in}
		}
	case 0x1:
		return op1NNN{in: in}
	case 0x2:
		return op2NNN{in: in}
	case 0x3:
		return op3XNNN{in: in}
	case 0x4:
		return op4XNN{in: in}
	case 0x5:
		return op5XY0{in: in}
	case 0x6:
		return op6XNN{in: in}
	case 0x7:
		return op7XNN{in: in}
	case 0x8:
		switch in.lsb {
		case 0:
			return op8XY0{in: in}
		case 1:
			return op8XY1{in: in}
		case 2:
			return op8XY2{in: in}
		case 3:
			return op8XY3{in: in}
		case 4:
			return op8XY4{in: in}
		case 5:
			return op8XY5{in: in}
		case 6:
			return op8XY6{in: in}
		case 7:
			return op8XY7{in: in}
		case 0xE:
			return op8XYE{in: in}
		}
	case 0x9:
		return op9XY0{in: in}
	case 0xA:
		return opANNN{in: in}
	case 0xB:
		return opBNNN{in: in}
	case 0xC:
		return opCXNN{in: in}
	case 0xD:
		return opDXYN{in: in}
	case 0xE:
		switch in.nn {
		case 0x9E:
			return opEX9E{in: in}
		case 0xA1:
			return opEXA1{in: in}
		}
	case 0xF:
		switch in.nn {
		case 0x07:
			return opFX07{in: in}
		case 0x0A:
			return opFX0A{in: in}
		case 0x15:
			return opFX15{in: in}
		case 0x18:
			return opFX18{in: in}
		case 0x1E:
			return opFX1E{in: in}
		case 0x29:
			return opFX29{in: in}
		case 0x33:
			return opFX33{in: in}
		case 0x55:
			return opFX55{in: in}
		case 0x65:
			return opFX65{in: in}
		}
	}

	return opUnknown{}
}

// opUnknown is unknown ¯\_(ツ)_/¯
type opUnknown struct{}

func (u opUnknown) run(c *Emulator) {}

type op0NNN struct {
	in *instructionInput
}

func (o op0NNN) run(c *Emulator) {
	// 0x0NNN - Machine language routine (no-op in most implementations)
	// Some implementations call a machine language routine at address NNN
	// For compatibility, we'll make this a no-op
}

// op00E0 clears the display
type op00E0 struct{}

func (op00E0) run(c *Emulator) {
	c.fbMu.Lock()
	defer c.fbMu.Unlock()

	for i := range c.fb {
		for j := range c.fb[i] {
			c.fb[i][j] = false
		}
	}
}

// op00EE returns from a subroutine
type op00EE struct{}

func (op00EE) run(c *Emulator) {
	c.sp--
	c.pc = c.stack[c.sp]
}

// op1NNN jumps to address NNN
type op1NNN struct {
	in *instructionInput
}

func (o op1NNN) run(c *Emulator) {
	c.pc = o.in.nnn
}

// op2NNN calls a subroutine at address NNN
type op2NNN struct {
	in *instructionInput
}

func (o op2NNN) run(c *Emulator) {
	c.stack[c.sp] = c.pc
	c.sp++
	c.pc = o.in.nnn
}

// op3XNNN skips the next instruction if Vx == nn
type op3XNNN struct {
	in *instructionInput
}

func (o op3XNNN) run(c *Emulator) {
	if c.registers[o.in.x] == o.in.nn {
		c.pcUP()
	}
}

// op4XNN skips the next instruction if Vx != nn
type op4XNN struct {
	in *instructionInput
}

func (o op4XNN) run(c *Emulator) {
	if c.registers[o.in.x] != o.in.nn {
		c.pcUP()
	}
}

// op5XY0 skips the next instruction if Vx == Vy
type op5XY0 struct {
	in *instructionInput
}

func (o op5XY0) run(c *Emulator) {
	if c.registers[o.in.x] == c.registers[o.in.y] {
		c.pcUP()
	}
}

// op6XNN sets Vx to nn
type op6XNN struct {
	in *instructionInput
}

func (o op6XNN) run(c *Emulator) {
	c.registers[o.in.x] = o.in.nn
}

// op7XNN sets Vx to Vx + nn
type op7XNN struct {
	in *instructionInput
}

func (o op7XNN) run(c *Emulator) {
	c.registers[o.in.x] += o.in.nn
}

// op8XY0 sets Vx to Vy
type op8XY0 struct {
	in *instructionInput
}

func (o op8XY0) run(c *Emulator) {
	c.registers[o.in.x] = c.registers[o.in.y]
}

// op8XY1 sets Vx to Vx OR Vy
type op8XY1 struct {
	in *instructionInput
}

func (o op8XY1) run(c *Emulator) {
	c.registers[o.in.x] |= c.registers[o.in.y]
}

// op8XY2 sets Vx to Vx AND Vy
type op8XY2 struct {
	in *instructionInput
}

func (o op8XY2) run(c *Emulator) {
	c.registers[o.in.x] &= c.registers[o.in.y]
}

// op8XY3 sets Vx to Vx XOR Vy
type op8XY3 struct {
	in *instructionInput
}

func (o op8XY3) run(c *Emulator) {
	c.registers[o.in.x] ^= c.registers[o.in.y]
}

// op8XY4 sets Vx to Vx + Vy, and carry.
type op8XY4 struct {
	in *instructionInput
}

func (o op8XY4) run(c *Emulator) {
	sum := uint16(c.registers[o.in.x]) + uint16(c.registers[o.in.y])
	c.flag(sum > 0xFF)

	c.registers[o.in.x] = uint8(sum)
}

// op8XY5 sets Vx to Vx - Vy, and borrow.
type op8XY5 struct {
	in *instructionInput
}

func (o op8XY5) run(c *Emulator) {
	diff := int8(c.registers[o.in.x]) - int8(c.registers[o.in.y])
	c.registers[o.in.x] = uint8(diff)
	c.flag(diff >= 0)
}

// op8XY6 sets Vx to Vx SHIFT RIGHT.
type op8XY6 struct {
	in *instructionInput
}

func (o op8XY6) run(c *Emulator) {
	c.flag(c.registers[o.in.x]%2 == 1)
	c.registers[o.in.x] = c.registers[o.in.y] / 2
}

// op8XY7 sets Vx to Vy - Vx, and borrow.
type op8XY7 struct {
	in *instructionInput
}

func (o op8XY7) run(c *Emulator) {
	diff := c.registers[o.in.y] - c.registers[o.in.x]
	c.registers[o.in.x] = diff
	c.flag(c.registers[o.in.y] >= c.registers[o.in.x])
}

// op8XYE sets Vx to Vx SHIFT LEFT.
type op8XYE struct {
	in *instructionInput
}

func (o op8XYE) run(c *Emulator) {
	c.flag(c.registers[o.in.x]&0x80 == 0x80)
	c.registers[o.in.x] = c.registers[o.in.y] << 1
}

// op9XY0 skips the next instruction if Vx != Vy.
type op9XY0 struct {
	in *instructionInput
}

func (o op9XY0) run(c *Emulator) {
	if c.registers[o.in.x] != c.registers[o.in.y] {
		c.pcUP()
	}
}

// opANNN sets I to nnn.
type opANNN struct {
	in *instructionInput
}

func (o opANNN) run(c *Emulator) {
	c.index = o.in.nnn
}

// opBNNN jumps to location nnn + V0.
type opBNNN struct {
	in *instructionInput
}

func (o opBNNN) run(c *Emulator) {
	// Jump to location nnn + V0.
	c.pc = uint16(c.registers[0]) + o.in.nnn
}

type opCXNN struct {
	in *instructionInput
}

func (o opCXNN) run(c *Emulator) {
	c.registers[o.in.x] = uint8(rand.Intn(256)) & o.in.nn
}

// opDXYN draws a sprite at position Vx, Vy with n bytes of sprite data starting at memory address I.
type opDXYN struct {
	in *instructionInput
}

func (o opDXYN) run(c *Emulator) {
	c.fbMu.Lock()
	defer c.fbMu.Unlock()

	c.flag(false)

	for i := uint8(0); i < o.in.n; i++ {
		spriteRow := c.memory[c.index+uint16(i)]

		for j := uint8(0); j < 8; j++ {
			pixel := spriteRow & (0x80 >> j)
			if pixel == 0 {
				continue
			}
			col := (c.registers[o.in.x] + j) % display.Width
			row := (c.registers[o.in.y] + i) % display.Height

			fbValue := c.fb[col][row]
			c.fb[col][row] = !fbValue

			if fbValue {
				c.flag(true)
			}
		}
	}
}

// opEX9E skips the next instruction if the key with the value of Vx is pressed.
type opEX9E struct {
	in *instructionInput
}

func (o opEX9E) run(c *Emulator) {
	keys := c.input.GetKeys()
	if keys[c.registers[o.in.x]] {
		c.pcUP()
	}
}

// opEXA1 skips the next instruction if the key with the value of Vx is not pressed.
type opEXA1 struct {
	in *instructionInput
}

func (o opEXA1) run(c *Emulator) {
	keys := c.input.GetKeys()
	if !keys[c.registers[o.in.x]] {
		c.pcUP()
	}
}

// opFX0A waits for a key press, stores the value of the key in Vx.
type opFX0A struct {
	in *instructionInput
}

func (o opFX0A) run(c *Emulator) {
	keys := c.input.GetKeys()

	if c.waitingForKey && !keys[c.keyWaitTarget] {
		c.registers[o.in.x] = c.keyWaitTarget
		c.waitingForKey = false
		c.keyWaitTarget = 0
		return
	}

	for i := range uint8(16) {
		if keys[i] {
			c.waitingForKey = true
			c.keyWaitTarget = i
		}
	}

	c.pcDown()
}

// opFX1E adds the value of Vx to I.
type opFX1E struct {
	in *instructionInput
}

func (o opFX1E) run(c *Emulator) {
	c.index += uint16(c.registers[o.in.x])
}

// opFX07 sets Vx to the value of the delay timer.
type opFX07 struct {
	in *instructionInput
}

func (o opFX07) run(c *Emulator) {
	c.registers[o.in.x] = c.delayTimer
}

// opFX15 sets the delay timer to the value of Vx.
type opFX15 struct {
	in *instructionInput
}

func (o opFX15) run(c *Emulator) {
	c.delayTimer = c.registers[o.in.x]
}

// opFX18 sets the sound timer to the value of Vx.
type opFX18 struct {
	in *instructionInput
}

func (o opFX18) run(c *Emulator) {
	c.soundTimer = c.registers[o.in.x]
}

// opFX29 sets I to the location of the sprite for the character in Vx.
type opFX29 struct {
	in *instructionInput
}

func (o opFX29) run(c *Emulator) {
	c.index = uint16(c.registers[o.in.x]) * 5
}

// opFX33 decodes the decimal value of Vx into three digits and stores them in memory at locations I, I+1, and I+2.
type opFX33 struct {
	in *instructionInput
}

func (o opFX33) run(c *Emulator) {
	c.memory[c.index] = (c.registers[o.in.x] / 100) % 10
	c.memory[c.index+1] = (c.registers[o.in.x] / 10) % 10
	c.memory[c.index+2] = (c.registers[o.in.x]) % 10
}

// opFX55 copies the values of V0 through Vx into memory, starting at the address in I.
type opFX55 struct {
	in *instructionInput
}

func (o opFX55) run(c *Emulator) {
	for i := uint8(0); i <= o.in.x; i++ {
		c.memory[c.index+uint16(i)] = c.registers[i]
	}
	// Super Chip8 behavior: increment I by x+1
	c.index += uint16(o.in.x) + 1
}

// opFX65 copies memory into registers V0 through Vx.
type opFX65 struct {
	in *instructionInput
}

func (o opFX65) run(c *Emulator) {
	for i := uint8(0); i <= o.in.x; i++ {
		c.registers[i] = c.memory[c.index+uint16(i)]
	}
	// Super Chip8 behavior: increment I by x+1
	c.index += uint16(o.in.x) + 1
}
