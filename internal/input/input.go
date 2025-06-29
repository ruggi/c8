package input

type KeysMap [16]bool

type Manager interface {
	GetKeys() KeysMap
}
