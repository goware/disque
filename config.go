package disque

import "time"

// Config represents Disque configuration for certain operations.
type Config struct {
	Timeout    time.Duration // Each operation fails after a specified timeout elapses. Blocks by default.
	Replicate  int           // Replicate job to at least N nodes before Add() returns.
	Delay      time.Duration // Schedule the job on Add() - enqueue after a specified time.
	RetryAfter time.Duration // Re-queue job after a specified time elapses between Get() and Ack().
	TTL        time.Duration // Remove the job from the queue after a specified time.
	MaxLen     int           // Fail on Add() if there are more than N jobs in the queue.
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

// Timeout option applied to a single operation.
func (conn *Conn) Timeout(timeout time.Duration) *Conn {
	conn.conf.Timeout = timeout
	return &Conn{pool: conn.pool, conf: conn.conf}
}

// Replicate option applied to a single operation.
func (conn *Conn) Replicate(replicate int) *Conn {
	conn.conf.Replicate = replicate
	return &Conn{pool: conn.pool, conf: conn.conf}
}

// Delay option applied to a single operation.
func (conn *Conn) Delay(delay time.Duration) *Conn {
	conn.conf.Delay = delay
	return &Conn{pool: conn.pool, conf: conn.conf}
}

// RetryAfter option applied to a single operation.
func (conn *Conn) RetryAfter(after time.Duration) *Conn {
	conn.conf.RetryAfter = after
	return &Conn{pool: conn.pool, conf: conn.conf}
}

// TTL option applied to a single operation.
func (conn *Conn) TTL(ttl time.Duration) *Conn {
	conn.conf.TTL = ttl
	return &Conn{pool: conn.pool, conf: conn.conf}
}

// MaxLen option applied to a single operation.
func (conn *Conn) MaxLen(maxlen int) *Conn {
	conn.conf.MaxLen = maxlen
	return &Conn{pool: conn.pool, conf: conn.conf}
}
