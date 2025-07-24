package editor

import "telescope/text"

func (e *editor) renderWithoutLock() {
	getRowForView := func(t text.Text, row int) []rune {
		if row < t.Len() {
			return t.Get(row)
		} else {
			return []rune{'~'}
		}
	}
	render := func() View {

		view := View{
			WinName: e.view.winName,
			WinData: nil,
			WinCursor: Cursor{
				Row: e.textCursor.Row - e.window.tlRow,
				Col: e.textCursor.Col - e.window.tlCol,
			},
			TextCursor: e.textCursor,
			Background: e.view.background,
			Message:    e.view.message,
		}

		// data
		view.WinData = make([][]rune, e.window.height)
		for row := 0; row < e.window.height; row++ {
			view.WinData[row] = getRowForView(e.text, row+e.window.tlRow)
		}
		return view
	}

	e.renderCh <- render()
}
