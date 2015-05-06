package disque

// Job represents job/message returned from a Disque server.
type Job struct {
	ID    string
	Data  string
	Queue string
}
