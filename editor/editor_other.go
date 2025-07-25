package editor

func (e *editor) Resize(height int, width int) {
	e.lockUpdateRender(func() {
		if e.view.height == height && e.view.width == width {
			return
		}
		e.view.height, e.view.width = height, width
		e.moveRelativeAndFixWithoutLock(0, 0)
		e.setMessageWithoutLock("resize to %dx%d", height, width)
	})
}

func (e *editor) Tabular() {
	// tab is two spaces
	e.Type(' ')
	e.Type(' ')
}
