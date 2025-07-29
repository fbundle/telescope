package log

type Command string

const (
	CommandSetVersion Command = "set_version" // set version of serializer
	CommandType       Command = "type"
	CommandEnter      Command = "enter"
	CommandBackspace  Command = "backspace"
	CommandDelete     Command = "delete"
	CommandUndo       Command = "undo"
	CommandRedo       Command = "redo"
)

type Entry struct {
	Command   Command `json:"command"`
	Version   uint64  `json:"version,omitempty"`
	CursorRow uint64  `json:"cursor_row,omitempty"`
	CursorCol uint64  `json:"cursor_col,omitempty"`
	Rune      rune    `json:"rune,omitempty"`
	Line      []rune  `json:"line,omitempty"` // only for version 2+
}
