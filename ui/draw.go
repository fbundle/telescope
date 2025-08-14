package ui

import "github.com/gdamore/tcell/v2"

type drawFunc = func(x int, y int, primary rune, combining []rune, style tcell.Style)
type drawCtx = func(func(width int, height int, draw drawFunc))

// makeDrawContext - draw into a rectangle (offsetX int, offsetY int, width int, height int)
func makeDrawContext(s tcell.Screen, offsetX int, offsetY int, width int, height int) drawCtx {
	return func(f func(width int, height int, draw drawFunc)) {
		f(width, height, func(x int, y int, primary rune, combining []rune, style tcell.Style) {
			if 0 <= x && x < width && 0 <= y && y < height {
				s.SetContent(offsetX+x, offsetY+y, primary, combining, style)
			}
		})
	}
}
