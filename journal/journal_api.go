package journal

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
