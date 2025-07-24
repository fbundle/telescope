package journal

import "context"

type Command string

const (
	CommandType      Command = "type"
	CommandEnter     Command = "enter"
	CommandBackspace Command = "backspace"
	CommandDelete    Command = "delete"
)

type Entry struct {
	Command   Command `json:"command"`
	Rune      rune    `json:"rune"`
	CursorRow int     `json:"cursor_row"`
	CursorCol int     `json:"cursor_col"`
}

type Writer interface {
	Write(e Entry) Writer
}

var NewJournalWriter = func(ctx context.Context, filename string) (Writer, error) {
	return NewWriter[Entry](ctx, filename)
}
