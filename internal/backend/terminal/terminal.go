package terminal

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	_ "github.com/gdamore/tcell/v2/encoding"
	"github.com/ruggi/c8/internal/display"
	"github.com/ruggi/c8/internal/input"
)

const simulatedKeyUpMillis = 250 // simulating key up since the terminal doesn't know about those

type terminal struct {
	mu     sync.RWMutex
	keys   [16]int64
	s      tcell.Screen
	stopCh chan struct{}
	keyCh  chan *tcell.EventKey
}

func New() (*terminal, error) {
	s, err := tcell.NewScreen()
	if err != nil {
		return nil, fmt.Errorf("new screen: %w", err)
	}

	err = s.Init()
	if err != nil {
		return nil, fmt.Errorf("init: %w", err)
	}

	s.SetStyle(tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.ColorDefault))
	s.Clear()

	stopCh := make(chan struct{})
	keyCh := make(chan *tcell.EventKey)
	go func() {
		for {
			select {
			case <-stopCh:
				return
			default:
				ev := s.PollEvent()
				switch ev := ev.(type) {
				case *tcell.EventKey:
					keyCh <- ev
				}
			}
		}
	}()

	return &terminal{
		s:      s,
		stopCh: stopCh,
		keyCh:  keyCh,
	}, nil
}

func (t *terminal) Name() string {
	return "terminal"
}

func (t *terminal) Update() {
	select {
	case ev := <-t.keyCh:
		t.handleKeyEvent(ev)
	default:
		t.handleKeyEvent(nil)
	}
}

func (t *terminal) Render(fb display.Framebuffer) error {
	t.s.Clear()

	// draw a rectangle around the screen
	// tl
	t.s.SetCell(0, 0, tcell.StyleDefault, '┌')
	// tr
	t.s.SetCell(display.Width+1, 0, tcell.StyleDefault, '┐')
	// top
	for x := 1; x <= display.Width; x++ {
		t.s.SetCell(x, 0, tcell.StyleDefault, '─')
		t.s.SetCell(x, display.Height+1, tcell.StyleDefault, '─')
	}
	// bl
	t.s.SetCell(0, display.Height+1, tcell.StyleDefault, '└')
	// br
	t.s.SetCell(display.Width+1, display.Height+1, tcell.StyleDefault, '┘')
	// left
	for y := 1; y <= display.Height; y++ {
		t.s.SetCell(0, y, tcell.StyleDefault, '│')
		t.s.SetCell(display.Width+1, y, tcell.StyleDefault, '│')
	}

	for x := range display.Width {
		for y := range display.Height {
			if fb[x][y] {
				t.s.SetCell(x+1, y+1, tcell.StyleDefault, '█')
			}
		}
	}

	t.s.Show()
	return nil
}

func (t *terminal) GetKeys() input.KeysMap {
	t.mu.RLock()
	defer t.mu.RUnlock()

	now := time.Now().UnixMilli()

	keys := input.KeysMap{}
	for i, k := range t.keys {
		keys[i] = now-k < simulatedKeyUpMillis
	}
	return keys
}

func (t *terminal) Close() {
	close(t.stopCh)
	close(t.keyCh)
	t.s.Fini()
}

// handleKeyEvent processes keyboard input and maps it to CHIP-8 keys
func (t *terminal) handleKeyEvent(ev *tcell.EventKey) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Reset all keys
	if ev == nil {
		return
	}

	now := time.Now().UnixMilli()
	switch ev.Rune() {
	case '1':
		t.keys[0x1] = now
	case '2':
		t.keys[0x2] = now
	case '3':
		t.keys[0x3] = now
	case '4':
		t.keys[0x4] = now
	case '5':
		t.keys[0x5] = now
	case '6':
		t.keys[0x6] = now
	case '7':
		t.keys[0x7] = now
	case '8':
		t.keys[0x8] = now
	case '9':
		t.keys[0x9] = now
	case '0':
		t.keys[0x0] = now
	case 'a':
		t.keys[0xA] = now
	case 'b':
		t.keys[0xB] = now
	case 'c':
		t.keys[0xC] = now
	case 'd':
		t.keys[0xD] = now
	case 'e':
		t.keys[0xE] = now
	case 'f':
		t.keys[0xF] = now
	}

	// Handle special keys
	switch ev.Key() {
	case tcell.KeyEsc:
		t.s.Fini()
		os.Exit(0)
	}
}

func (t *terminal) Buzz() error {
	fmt.Print("\a")
	return nil
}
