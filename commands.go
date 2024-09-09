package tinykafka

/*
// Comandos
command|body size|msg pack body

// Comandos relacionados a mensajes
command|credentials size|credentials|MIME type size|MIME type|body size|body
*/

type Command uint8

const (
	EventCommandType = Command(iota)
	ErrorCommandType
)

type EventCommand struct {
	Name string
	Data []byte
}
