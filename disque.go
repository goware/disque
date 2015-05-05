package disque

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/garyburd/redigo/redis"
)

type Conn struct {
	pool *redis.Pool
	conf JobConfig
}

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

	return &Conn{pool: pool, conf: defaultJobConfig}, nil
}

func (conn *Conn) Close() {
	conn.pool.Close()
}

func (conn *Conn) Use(conf JobConfig) *Conn {
	conn.conf = conf
	return conn
}

func (conn *Conn) With(conf JobConfig) *Conn {
	return &Conn{pool: conn.pool, conf: conf}
}

func (conn *Conn) RetryAfter(after time.Duration) *Conn {
	conn.conf.RetryAfter = after
	return &Conn{pool: conn.pool, conf: conn.conf}
}

func (conn *Conn) Timeout(timeout time.Duration) *Conn {
	conn.conf.Timeout = timeout
	return &Conn{pool: conn.pool, conf: conn.conf}
}

func (conn *Conn) Ping() error {
	sess := conn.pool.Get()
	defer sess.Close()

	if _, err := sess.Do("PING"); err != nil {
		return err
	}
	return nil
}

// Eh, none of the following builds successfully:
//
//   reply, err := sess.Do("GETJOB", "FROM", queue, redis.Args{})
//   reply, err := sess.Do("GETJOB", "FROM", queue, redis.Args{}...)
//   reply, err := sess.Do("GETJOB", "FROM", queue, extraQueues)
//   reply, err := sess.Do("GETJOB", "FROM", queue, extraQueues...)
//
//   > build error:
//   > too many arguments in call to sess.Do
//
// So.. let's work around this with reflect pkg.
func (conn *Conn) do(args []interface{}) (interface{}, error) {
	sess := conn.pool.Get()
	defer sess.Close()

	fn := reflect.ValueOf(sess.Do)
	reflectArgs := []reflect.Value{}
	for _, arg := range args {
		reflectArgs = append(reflectArgs, reflect.ValueOf(arg))
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

func (conn *Conn) Add(data string, queue string) (*Job, error) {

	args := []interface{}{
		"ADDJOB", queue, data, conn.conf.Timeout.Nanoseconds() / 1000000,
	}

	if conn.conf.RetryAfter > 0 {
		args = append(args, "RETRY", conn.conf.RetryAfter.Seconds())
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

func (conn *Conn) Get(queue string, extra ...string) (*Job, error) {
	args := []interface{}{
		"GETJOB",
		"TIMEOUT",
		int(conn.conf.Timeout.Nanoseconds() / 1000000),
		"FROM",
		queue,
	}
	for _, arg := range extra {
		args = append(args, reflect.ValueOf(arg))
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

func (conn *Conn) Ack(job *Job) error {
	sess := conn.pool.Get()
	defer sess.Close()

	if _, err := sess.Do("ACKJOB", job.ID); err != nil {
		return err
	}
	return nil
}

func (conn *Conn) Nack(job *Job) error {
	sess := conn.pool.Get()
	defer sess.Close()

	if _, err := sess.Do("ENQUEUE", job.ID); err != nil {
		return err
	}
	return nil
}
