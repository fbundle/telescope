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
		view := View{
			Cursor: e.cursor,
			Window: e.window,
			Status: e.status,
		}

		t := e.text.Get()
		data := make([][]rune, e.window.Dimension.Row)
		for row := 0; row < e.window.Dimension.Row; row++ {
			data[row] = getLineForView(t, row+e.window.TopLeft.Row, e.window.TopLeft.Col)
		}

		view.Text = t
		view.Window.Data = data
		return view

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
