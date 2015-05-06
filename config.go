package disque

import "time"

type Config struct {
	Timeout    time.Duration
	Replicate  int
	Delay      time.Duration
	RetryAfter time.Duration
	TTL        time.Duration
	MaxLen     int
}

func (conn *Conn) Use(conf Config) *Conn {
	conn.conf = conf
	return conn
}

func (conn *Conn) With(conf Config) *Conn {
	return &Conn{pool: conn.pool, conf: conf}
}

func (conn *Conn) Timeout(timeout time.Duration) *Conn {
	conn.conf.Timeout = timeout
	return &Conn{pool: conn.pool, conf: conn.conf}
}

func (conn *Conn) Replicate(replicate int) *Conn {
	conn.conf.Replicate = replicate
	return &Conn{pool: conn.pool, conf: conn.conf}
}

func (conn *Conn) Delay(delay time.Duration) *Conn {
	conn.conf.Delay = delay
	return &Conn{pool: conn.pool, conf: conn.conf}
}

func (conn *Conn) RetryAfter(after time.Duration) *Conn {
	conn.conf.RetryAfter = after
	return &Conn{pool: conn.pool, conf: conn.conf}
}

func (conn *Conn) TTL(ttl time.Duration) *Conn {
	conn.conf.TTL = ttl
	return &Conn{pool: conn.pool, conf: conn.conf}
}

func (conn *Conn) MaxLen(maxlen int) *Conn {
	conn.conf.MaxLen = maxlen
	return &Conn{pool: conn.pool, conf: conn.conf}
}
