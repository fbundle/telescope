package log

type Command string

const (
	CommandSetVersion Command = "set_version" // set version of serializer
	CommandType       Command = "type"
	CommandEnter      Command = "enter"
	CommandBackspace  Command = "backspace"
	CommandDelete     Command = "delete"
)

type Entry struct {
	Command   Command `json:"command"`
	Rune      rune    `json:"rune,omitempty"`
	CursorRow int     `json:"cursor_row,omitempty"`
	CursorCol int     `json:"cursor_col,omitempty"`
	Version   int     `json:"version,omitempty"`
}
