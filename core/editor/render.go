package editor

import (
	"telescope/core/text"
)

func (e *editor) renderWithoutLock() {
	e.renderCh <- e.makeView()
}

func (e *editor) makeView() View {
	getLineForView := func(t text.Text, row int, col int) []rune {
		if row < t.Len() {
			line := t.Get(row)
			if col >= len(line) {
				return nil
			} else {
				return line[col:] // shifted by col
			}
		} else {
			return []rune{'~'}
		}
	}
	render := func() View {
		c := e.cursor
		t := e.text.Get()
		data := make([][]rune, e.window.Height)
		for row := 0; row < e.window.Height; row++ {
			data[row] = getLineForView(t, row+e.window.TopLeftRow, e.window.TopLeftCol)
		}
		e.window.Data = data

		return View{
			Window: e.window,
			Status: e.status,
			Cursor: c,
			Text:   t,
		}

	}
	return render()
}

func (e *editor) Render() (view View) {
	e.lock(func() {
		view = e.makeView()
	})
	return view
}

func (e *editor) Update() <-chan View {
	return e.renderCh
}
