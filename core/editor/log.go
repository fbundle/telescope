package editor

type Command string

const (
	CommandSetVersion Command = "set_version" // set version of serializer
	CommandType       Command = "type"
	CommandEnter      Command = "enter"
	CommandBackspace  Command = "backspace"
	CommandDelete     Command = "delete"
	CommandUndo       Command = "undo"
	CommandRedo       Command = "redo"
	CommandInsertLine Command = "insert_line"
	CommandDeleteLine Command = "delete_line"
)

type LogEntry struct {
	Command Command  `json:"command"`
	Version uint64   `json:"version,omitempty"`
	Row     uint64   `json:"row,omitempty"`
	Col     uint64   `json:"col,omitempty"`
	Rune    rune     `json:"rune,omitempty"`
	Text    [][]rune `json:"text,omitempty"`
	Count   uint64   `json:"count,omitempty"`
	Beg     uint64   `json:"beg,omitempty"`
	End     uint64   `json:"end,omitempty"`
}
