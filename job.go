package disque

// Job represents job (and its data) belonging to a queue.
type Job struct {
	ID    string
	Data  string
	Queue string
}
