package sdl

import (
	"fmt"
	"math"
	"os"
	"sync"
	"time"

	"github.com/ruggi/c8/internal/display"
	"github.com/ruggi/c8/internal/input"
	"github.com/veandco/go-sdl2/sdl"
)

const scale = 10

type sdlBackend struct {
	// display
	window   *sdl.Window
	renderer *sdl.Renderer
	texture  *sdl.Texture
	pixels   []byte

	// input
	keys input.KeysMap

	// buzzer
	audioDevice  sdl.AudioDeviceID
	audioBytes   []byte
	isBuzzing    bool
	buzzDuration float64
	rmu          sync.RWMutex
}

func (b *sdlBackend) Name() string {
	return "SDL"
}

func New(title string) (*sdlBackend, error) {
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		return nil, fmt.Errorf("init sdl: %w", err)
	}

	window, err := sdl.CreateWindow(
		title,
		sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		display.Width*scale, display.Height*scale,
		sdl.WINDOW_SHOWN,
	)
	if err != nil {
		return nil, fmt.Errorf("create window: %w", err)
	}

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		return nil, fmt.Errorf("create framebuffer: %w", err)
	}

	texture, err := renderer.CreateTexture(
		sdl.PIXELFORMAT_RGBA8888,
		sdl.TEXTUREACCESS_STREAMING,
		display.Width, display.Height,
	)
	if err != nil {
		return nil, fmt.Errorf("create texture: %w", err)
	}

	audioDevice, audioBytes, err := setupSDLBuzzer(44100, 0.1, 440)
	if err != nil {
		return nil, fmt.Errorf("setup buzzer: %w", err)
	}

	backend := &sdlBackend{
		window:   window,
		renderer: renderer,
		texture:  texture,
		pixels:   make([]byte, display.Width*display.Height*4), // RGBA format

		buzzDuration: 0.1,
		audioDevice:  audioDevice,
		audioBytes:   audioBytes,
	}

	return backend, nil
}

func (b *sdlBackend) Render(fb display.Framebuffer) error {
	b.renderer.SetDrawColor(0, 0, 0, 0xFF)
	b.renderer.Clear()

	b.renderer.SetDrawColor(0xFF, 0xFF, 0xFF, 0xFF)
	for x := range fb {
		for y := range fb[x] {
			if !fb[x][y] {
				continue
			}
			rect := sdl.Rect{
				X: int32(x * scale),
				Y: int32(y * scale),
				W: scale,
				H: scale,
			}
			b.renderer.FillRect(&rect)
		}
	}

	b.renderer.Present()

	return nil
}

func (b *sdlBackend) Close() {
	if b.audioDevice != 0 {
		sdl.CloseAudioDevice(b.audioDevice)
	}
	_ = b.texture.Destroy()
	_ = b.window.Destroy()
	sdl.Quit()
}

func (b *sdlBackend) Update() {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch event := event.(type) {
		case *sdl.QuitEvent:
			os.Exit(0)
		case *sdl.KeyboardEvent:
			ke := event
			pressed := ke.Type == sdl.KEYDOWN
			switch ke.Keysym.Scancode {
			case sdl.SCANCODE_1, sdl.SCANCODE_KP_1:
				b.keys[0x1] = pressed
			case sdl.SCANCODE_2, sdl.SCANCODE_KP_2:
				b.keys[0x2] = pressed
			case sdl.SCANCODE_3, sdl.SCANCODE_KP_3:
				b.keys[0x3] = pressed
			case sdl.SCANCODE_4, sdl.SCANCODE_KP_4:
				b.keys[0x4] = pressed
			case sdl.SCANCODE_5, sdl.SCANCODE_KP_5:
				b.keys[0x5] = pressed
			case sdl.SCANCODE_6, sdl.SCANCODE_KP_6:
				b.keys[0x6] = pressed
			case sdl.SCANCODE_7, sdl.SCANCODE_KP_7:
				b.keys[0x7] = pressed
			case sdl.SCANCODE_8, sdl.SCANCODE_KP_8:
				b.keys[0x8] = pressed
			case sdl.SCANCODE_9, sdl.SCANCODE_KP_9:
				b.keys[0x9] = pressed
			case sdl.SCANCODE_0, sdl.SCANCODE_KP_0:
				b.keys[0x0] = pressed
			case sdl.SCANCODE_A:
				b.keys[0xA] = pressed
			case sdl.SCANCODE_B:
				b.keys[0xB] = pressed
			case sdl.SCANCODE_C:
				b.keys[0xC] = pressed
			case sdl.SCANCODE_D:
				b.keys[0xD] = pressed
			case sdl.SCANCODE_E:
				b.keys[0xE] = pressed
			case sdl.SCANCODE_F:
				b.keys[0xF] = pressed
			}
		}
	}
}

func (b *sdlBackend) GetKeys() input.KeysMap {
	return b.keys
}

func (b *sdlBackend) Buzz() error {
	b.rmu.RLock()
	isBuzzing := b.isBuzzing
	b.rmu.RUnlock()

	if isBuzzing {
		return nil
	}

	b.rmu.Lock()
	b.isBuzzing = true
	b.rmu.Unlock()

	err := sdl.QueueAudio(b.audioDevice, b.audioBytes)
	if err != nil {
		return fmt.Errorf("queue audio: %w", err)
	}

	sdl.PauseAudioDevice(b.audioDevice, false)

	// stop the audio after the buzz duration
	go func() {
		time.Sleep(time.Duration(b.buzzDuration * float64(time.Second)))
		sdl.PauseAudioDevice(b.audioDevice, true)
		b.rmu.Lock()
		b.isBuzzing = false
		b.rmu.Unlock()
	}()

	return nil
}

func setupSDLBuzzer(sampleRate int, duration, frequency float64) (sdl.AudioDeviceID, []byte, error) {
	samples := int(float64(sampleRate) * duration)
	audioBuffer := make([]int16, samples)

	maxInt := int16(0x7FFF)
	volume := 0.3
	phase := 0.0
	for i := range samples {
		sample := int16(float64(maxInt) * volume * math.Sin(phase))
		audioBuffer[i] = sample
		phase += 2.0 * math.Pi * frequency / float64(sampleRate)
		phase = math.Mod(phase, 2*math.Pi)
	}

	audioSpec := sdl.AudioSpec{
		Freq:     int32(sampleRate),
		Format:   sdl.AUDIO_S16,
		Channels: 1,
		Samples:  uint16(samples),
	}

	audioBytes := make([]byte, len(audioBuffer)*2)
	for i, sample := range audioBuffer {
		audioBytes[i*2] = byte(sample & 0xFF)
		audioBytes[i*2+1] = byte((sample >> 8) & 0xFF)
	}

	device, err := sdl.OpenAudioDevice("", false, &audioSpec, nil, 0)
	if err != nil {
		return 0, nil, fmt.Errorf("open audio device: %w", err)
	}

	return device, audioBytes, nil
}
