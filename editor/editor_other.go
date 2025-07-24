package editor

import (
	"os"
	"telescope/feature"
	"telescope/text"
	"time"
)

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

func (e *editor) Save() {
	// saving is a synchronous task - can be made async but not needed
	var m text.Text
	e.lockUpdateRender(func() {
		if len(e.filenameTextOut) == 0 {
			e.setStatusWithoutLock("read only mode, cannot save")
			return
		}
		if !e.loaded {
			e.setStatusWithoutLock("cannot save, still loading")
			return
		}
		m = e.text
		file, err := os.Create(e.filenameTextOut)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		for _, line := range m.Iter {
			if feature.Debug() {
				time.Sleep(feature.DEBUG_IO_INTERVAL_MS * time.Millisecond)
			}
			_, err = file.WriteString(string(line) + "\n")
			if err != nil {
				panic(err)
			}
		}
		e.setStatusWithoutLock("saved")
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
