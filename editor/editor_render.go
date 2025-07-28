package editor

import "telescope/text"

func (e *editor) updateWithoutLock() {
	e.renderCh <- e.renderWithoutLock()
}

func (e *editor) renderWithoutLock() View {
	getRowForView := func(t text.Text, row int) []rune {
		if row < t.Len() {
			return t.Get(row)
		} else {
			return []rune{'~'}
		}
	}
	render := func() View {
		view := View{
			WinData: nil,
			WinCursor: Cursor{
				Row: e.cursor.Row - e.view.tlRow,
				Col: e.cursor.Col - e.view.tlCol,
			},
			TextCursor: e.cursor,

			Text: e.text.Get(),

			Header:     e.view.header,
			Command:    e.view.command,
			Background: e.view.background,
			Message:    e.view.message,
		}

		// data
		view.WinData = make([][]rune, e.view.height)
		for row := 0; row < e.view.height; row++ {
			view.WinData[row] = getRowForView(e.text.Get(), row+e.view.tlRow)
		}
		return view
	}
	return render()
}

func (e *editor) Render() (view View) {
	e.lockUpdate(func() {
		view = e.renderWithoutLock()
	})
	return view
}
