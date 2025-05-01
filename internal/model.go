package internal

type Model struct {
	mode string
}

func (m *Model) Mode() string {
	if m.mode == "" {
		return "io"
	}
	return m.mode
}

func (m *Model) SetMode(mode string) {
	m.mode = mode
}
