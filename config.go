package disque

import "time"

// Config represents Disque configuration for certain operations.
type Config struct {
	Timeout    time.Duration // Each operation will fail after a specified timeout. Blocks by default.
	Replicate  int           // Add(): Replicate job to at least N nodes before return.
	Delay      time.Duration // Add(): Schedule the job - enqueue after a specified time.
	RetryAfter time.Duration // Add(): Re-queue job after a specified time (between Get() and Ack()).
	TTL        time.Duration // Add(): Remove the job from queue after a specified time.
	MaxLen     int           // Add(): Fail if there are more than N jobs in the queue.
}

// Use applies given config to every subsequent operation of this connection.
func (conn *Conn) Use(conf Config) *Conn {
	conn.conf = conf
	return conn
}

// With applies given config to a single operation.
func (conn *Conn) With(conf Config) *Conn {
	return &Conn{pool: conn.pool, conf: conf}
}

// With applies Timeout to a single operation.
func (conn *Conn) Timeout(timeout time.Duration) *Conn {
	conn.conf.Timeout = timeout
	return &Conn{pool: conn.pool, conf: conn.conf}
}

// With applies Replicate to a single operation.
func (conn *Conn) Replicate(replicate int) *Conn {
	conn.conf.Replicate = replicate
	return &Conn{pool: conn.pool, conf: conn.conf}
}

// With applies Delay to a single operation.
func (conn *Conn) Delay(delay time.Duration) *Conn {
	conn.conf.Delay = delay
	return &Conn{pool: conn.pool, conf: conn.conf}
}

// With applies RetryAfter to a single operation.
func (conn *Conn) RetryAfter(after time.Duration) *Conn {
	conn.conf.RetryAfter = after
	return &Conn{pool: conn.pool, conf: conn.conf}
}

// With applies TTL to a single operation.
func (conn *Conn) TTL(ttl time.Duration) *Conn {
	conn.conf.TTL = ttl
	return &Conn{pool: conn.pool, conf: conn.conf}
}

// With applies MaxLen to a single operation.
func (conn *Conn) MaxLen(maxlen int) *Conn {
	conn.conf.MaxLen = maxlen
	return &Conn{pool: conn.pool, conf: conn.conf}
}
