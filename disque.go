package disque

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/garyburd/redigo/redis"
)

// Pool represent Redis connection to a Disque Pool
// with a certain Disque configuration.
type Pool struct {
	redis *redis.Pool
	conf  Config
}

// New creates a new connection to a given Disque Pool.
func New(address string, extra ...string) (*Pool, error) {
	pool := &redis.Pool{
		MaxIdle:     1024,
		MaxActive:   1024,
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

	return &Pool{redis: pool}, nil
}

// Close closes the connection to a Disque Pool.
func (pool *Pool) Close() error {
	return pool.redis.Close()
}

// Ping returns nil if Disque Pool is alive, error otherwise.
func (pool *Pool) Ping() error {
	sess := pool.redis.Get()
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
func (pool *Pool) do(args []interface{}) (interface{}, error) {
	sess := pool.redis.Get()
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
func (pool *Pool) Add(data string, queue string) (*Job, error) {
	args := []interface{}{
		"ADDJOB",
		queue,
		data,
		int(pool.conf.Timeout.Nanoseconds() / 1000000),
	}

	if pool.conf.Replicate > 0 {
		args = append(args, "REPLICATE", pool.conf.Replicate)
	}
	if pool.conf.Delay > 0 {
		delay := int(pool.conf.Delay.Seconds())
		if delay == 0 {
			delay = 1
		}
		args = append(args, "DELAY", delay)
	}
	if pool.conf.RetryAfter > 0 {
		retry := int(pool.conf.RetryAfter.Seconds())
		if retry == 0 {
			retry = 1
		}
		args = append(args, "RETRY", retry)
	}
	if pool.conf.TTL > 0 {
		ttl := int(pool.conf.TTL.Seconds())
		if ttl == 0 {
			ttl = 1
		}
		args = append(args, "TTL", ttl)
	}
	if pool.conf.MaxLen > 0 {
		args = append(args, "MAXLEN", pool.conf.MaxLen)
	}

	reply, err := pool.do(args)
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
func (pool *Pool) Get(queues ...string) (*Job, error) {
	if len(queues) == 0 {
		return nil, errors.New("expected at least one queue")
	}

	args := []interface{}{
		"GETJOB",
		"TIMEOUT",
		int(pool.conf.Timeout.Nanoseconds() / 1000000),
		"FROM",
	}
	for _, arg := range queues {
		args = append(args, arg)
	}

	reply, err := pool.do(args)
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
func (pool *Pool) Ack(job *Job) error {
	sess := pool.redis.Get()
	defer sess.Close()

	if _, err := sess.Do("ACKJOB", job.ID); err != nil {
		return err
	}
	return nil
}

// Nack re-queues a job back into its queue.
// Native NACKJOB discussed upstream at https://github.com/antirez/disque/issues/43.
func (pool *Pool) Nack(job *Job) error {
	sess := pool.redis.Get()
	defer sess.Close()

	if _, err := sess.Do("ENQUEUE", job.ID); err != nil {
		return err
	}
	return nil
}

// Wait waits for a job to finish (blocks until it's ACKed).
// Native WAITJOB discussed upstream at https://github.com/antirez/disque/issues/43.
func (pool *Pool) Wait(job *Job) error {
	sess := pool.redis.Get()
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
func (pool *Pool) Len(queue string) (int, error) {
	sess := pool.redis.Get()
	defer sess.Close()

	length, err := redis.Int(sess.Do("QLEN", queue))
	if err != nil {
		return 0, err
	}

	return length, nil
}
