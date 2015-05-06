package disque

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/garyburd/redigo/redis"
)

// Conn represent connection to a Disque Pool.
type Conn struct {
	pool *redis.Pool
	conf Config
}

// Connect creates new connection to a given Disque Pool.
func Connect(address string, extra ...string) (*Conn, error) {
	pool := &redis.Pool{
		MaxIdle:     64,
		MaxActive:   64,
		IdleTimeout: 300 * time.Second,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", address)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	return &Conn{pool: pool}, nil
}

// Close closes the connection to a Disque Pool.
func (conn *Conn) Close() {
	conn.pool.Close()
}

// Ping returns nil if Disque Pool is alive, error otherwise.
func (conn *Conn) Ping() error {
	sess := conn.pool.Get()
	defer sess.Close()

	if _, err := sess.Do("PING"); err != nil {
		return err
	}
	return nil
}

// do is a helper function that workarounds redigo/redis API
// flaws with a magic function Call() from the reflect pkg.
//
// None of the following builds or works successfully:
//
// reply, err := sess.Do("GETJOB", "FROM", queues, redis.Args{})
// reply, err := sess.Do("GETJOB", "FROM", queues, redis.Args{}...)
// reply, err := sess.Do("GETJOB", "FROM", queues)
// reply, err := sess.Do("GETJOB", "FROM", queues...)
//
// > Build error: "too many arguments in call to sess.Do"
// > Runtime error: "ERR wrong number of arguments for '...' command"
//
func (conn *Conn) do(args []interface{}) (interface{}, error) {
	sess := conn.pool.Get()
	defer sess.Close()

	fn := reflect.ValueOf(sess.Do)
	reflectArgs := make([]reflect.Value, len(args))
	for i, arg := range args {
		reflectArgs[i] = reflect.ValueOf(arg)
	}
	ret := fn.Call(reflectArgs)
	if len(ret) != 2 {
		return nil, errors.New("expected two return values")
	}
	if !ret[1].IsNil() {
		err, ok := ret[1].Interface().(error)
		if !ok {
			return nil, fmt.Errorf("expected error type, got: %T %#v", ret[1], ret[1])
		}
		return nil, err
	}
	if ret[0].IsNil() {
		return nil, fmt.Errorf("no data available")
	}
	reply, ok := ret[0].Interface().(interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected interface{} error type, got: %T %#v", ret[0], ret[0])
	}
	return reply, nil
}

// Add enqueues new job with a specified data to a given queue.
func (conn *Conn) Add(data string, queue string) (*Job, error) {
	args := []interface{}{
		"ADDJOB",
		queue,
		data,
		int(conn.conf.Timeout.Nanoseconds() / 1000000),
	}

	if conn.conf.Replicate > 0 {
		args = append(args, "REPLICATE", conn.conf.Replicate)
	}
	if conn.conf.Delay > 0 {
		delay := int(conn.conf.Delay.Seconds())
		if delay == 0 {
			delay = 1
		}
		args = append(args, "DELAY", delay)
	}
	if conn.conf.RetryAfter > 0 {
		retry := int(conn.conf.RetryAfter.Seconds())
		if retry == 0 {
			retry = 1
		}
		args = append(args, "RETRY", retry)
	}
	if conn.conf.TTL > 0 {
		ttl := int(conn.conf.TTL.Seconds())
		if ttl == 0 {
			ttl = 1
		}
		args = append(args, "TTL", ttl)
	}
	if conn.conf.MaxLen > 0 {
		args = append(args, "MAXLEN", conn.conf.MaxLen)
	}

	reply, err := conn.do(args)
	if err != nil {
		return nil, err
	}

	id, ok := reply.(string)
	if !ok {
		return nil, errors.New("unexpected reply: id")
	}

	return &Job{
		ID:    id,
		Data:  data,
		Queue: queue,
	}, nil
}

// Get returns first available job from a highest priority
// queue possible (left-to-right priority).
func (conn *Conn) Get(queues ...string) (*Job, error) {
	if len(queues) == 0 {
		return nil, errors.New("expected at least one queue")
	}

	args := []interface{}{
		"GETJOB",
		"TIMEOUT",
		int(conn.conf.Timeout.Nanoseconds() / 1000000),
		"FROM",
	}
	for _, arg := range queues {
		args = append(args, arg)
	}

	reply, err := conn.do(args)
	if err != nil {
		return nil, err
	}

	replyArr, ok := reply.([]interface{})
	if !ok || len(replyArr) != 1 {
		return nil, errors.New("unexpected reply #1")
	}
	arr, ok := replyArr[0].([]interface{})
	if !ok || len(arr) != 3 {
		return nil, errors.New("unexpected reply #2")
	}

	que, ok := arr[0].([]byte)
	if !ok {
		return nil, errors.New("unexpected reply: queue")
	}

	id, ok := arr[1].([]byte)
	if !ok {
		return nil, errors.New("unexpected reply: id")
	}

	data, ok := arr[2].([]byte)
	if !ok {
		return nil, errors.New("unexpected reply: data")
	}

	return &Job{
		ID:    string(id),
		Data:  string(data),
		Queue: string(que),
	}, nil
}

// Ack acknowledges (dequeues/removes) a job from its queue.
func (conn *Conn) Ack(job *Job) error {
	sess := conn.pool.Get()
	defer sess.Close()

	if _, err := sess.Do("ACKJOB", job.ID); err != nil {
		return err
	}
	return nil
}

// Nack re-queues a job back into its queue.
// Native NACKJOB discussed upstream at https://github.com/antirez/disque/issues/43.
func (conn *Conn) Nack(job *Job) error {
	sess := conn.pool.Get()
	defer sess.Close()

	if _, err := sess.Do("ENQUEUE", job.ID); err != nil {
		return err
	}
	return nil
}

// Wait waits for a job to finish (blocks until it's ACKed).
// Native WAITJOB discussed upstream at https://github.com/antirez/disque/issues/43.
func (conn *Conn) Wait(job *Job) error {
	sess := conn.pool.Get()
	defer sess.Close()

	for {
		reply, err := sess.Do("SHOW", job.ID)
		if err != nil {
			return err
		}
		if reply == nil {
			break
		}

		time.Sleep(10 * time.Millisecond)
	}

	return nil
}

// Len returns length of a given queue.
func (conn *Conn) Len(queue string) (int, error) {
	sess := conn.pool.Get()
	defer sess.Close()

	length, err := redis.Int(sess.Do("QLEN", queue))
	if err != nil {
		return 0, err
	}

	return length, nil
}
