package disque

import "time"

type Job struct {
	ID    string
	Data  string
	Queue string
}

type JobConfig struct {
	Timeout    time.Duration `redis:"TIMEOUT"`
	Replicate  int           `redis:"REPLICATE"`
	Delay      time.Duration `redis:"DELAY"`
	RetryAfter time.Duration `redis:"RETRY"`
	TTL        time.Duration `redis:"TTL"`
	MaxLen     int           `redis:"MAXLEN"`
}

var defaultJobConfig = JobConfig{
	Timeout: 500 * time.Millisecond,
}
