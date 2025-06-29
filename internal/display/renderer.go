package display

const (
	Width  = 64 // px
	Height = 32 // px
)

type Framebuffer [Width][Height]bool

type Manager interface {
	Render(fb Framebuffer) error
	Close()
}
