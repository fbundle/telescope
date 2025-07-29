package editor

import "telescope/text"

func (e *editor) postWithoutLock() {
	e.renderCh <- e.renderWithoutLock()
}

func (e *editor) renderWithoutLock() View {
	getRowForView := func(t text.Text, row int, col int) []rune {
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
		data := make([][]rune, e.windowInfo.height)
		for row := 0; row < e.windowInfo.height; row++ {
			data[row] = getRowForView(t, row+e.windowInfo.tlRow, e.windowInfo.tlCol)
		}

		return View{
			Window: Window{
				Data: data,
				Cursor: Cursor{
					Row: e.cursor.Row - e.windowInfo.tlRow,
					Col: e.cursor.Col - e.windowInfo.tlCol,
				},
			},
			Status: e.status,
			Cursor: c,
			Text:   t,
		}

	}
	return render()
}

func (e *editor) Render() (view View) {
	e.lockUpdate(func() {
		view = e.renderWithoutLock()
	})
	return view
}
func (e *editor) Update() <-chan View {
	return e.renderCh
}
