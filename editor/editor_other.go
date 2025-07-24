package editor

func (e *editor) Resize(height int, width int) {
	e.lockUpdateRender(func() {
		if e.window.height == height && e.window.width == width {
			return
		}
		e.window.height, e.window.width = height, width
		e.moveRelativeAndFixWithoutLock(0, 0)
		e.setStatusWithoutLock("resize to %dx%d", height, width)
	})
}

func (e *editor) Escape() {
	// do nothing
}
func (e *editor) Tabular() {
	// tab is two spaces
	e.Type(' ')
	e.Type(' ')
}
