package log

type Command string

const (
	CommandType       Command = "type"
	CommandEnter      Command = "enter"
	CommandBackspace  Command = "backspace"
	CommandDelete     Command = "delete"
	CommandSetVersion Command = "set_version" // set version of serializer
)

type Entry struct {
	Command   Command `json:"command"`
	Rune      rune    `json:"rune,omitempty"`
	CursorRow int     `json:"cursor_row,omitempty"`
	CursorCol int     `json:"cursor_col,omitempty"`
	Version   int     `json:"version,omitempty"`
}
