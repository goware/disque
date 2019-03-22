package disque

import "time"

// Job represents job/message returned from a Disque server.
type Job struct {
	ID                   string
	Data                 string
	Queue                string
	State                string
	TTL                  time.Duration
	Delay                time.Duration
	Retry                time.Duration
	CreatedAt            time.Time
	Replication          int64
	Nacks                int64
	AdditionalDeliveries int64
}
