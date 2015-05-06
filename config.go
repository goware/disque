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
func (pool *Pool) Use(conf Config) *Pool {
	pool.conf = conf
	return pool
}

// With applies given config to a single operation.
func (pool *Pool) With(conf Config) *Pool {
	return &Pool{redis: pool.redis, conf: conf}
}

// Timeout option applied to a single operation.
func (pool *Pool) Timeout(timeout time.Duration) *Pool {
	pool.conf.Timeout = timeout
	return &Pool{redis: pool.redis, conf: pool.conf}
}

// Replicate option applied to a single operation.
func (pool *Pool) Replicate(replicate int) *Pool {
	pool.conf.Replicate = replicate
	return &Pool{redis: pool.redis, conf: pool.conf}
}

// Delay option applied to a single operation.
func (pool *Pool) Delay(delay time.Duration) *Pool {
	pool.conf.Delay = delay
	return &Pool{redis: pool.redis, conf: pool.conf}
}

// RetryAfter option applied to a single operation.
func (pool *Pool) RetryAfter(after time.Duration) *Pool {
	pool.conf.RetryAfter = after
	return &Pool{redis: pool.redis, conf: pool.conf}
}

// TTL option applied to a single operation.
func (pool *Pool) TTL(ttl time.Duration) *Pool {
	pool.conf.TTL = ttl
	return &Pool{redis: pool.redis, conf: pool.conf}
}

// MaxLen option applied to a single operation.
func (pool *Pool) MaxLen(maxlen int) *Pool {
	pool.conf.MaxLen = maxlen
	return &Pool{redis: pool.redis, conf: pool.conf}
}
