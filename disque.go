package disque

import (
	"errors"
	"reflect"
	"time"

	"github.com/garyburd/redigo/redis"
)

type Conn struct {
	pool *redis.Pool
}

func Connect(address string, extra ...string) (Conn, error) {
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
	return Conn{pool: pool}, nil
}

func (conn Conn) Close() {
	conn.pool.Close()
}

func (conn Conn) Ping() error {
	sess := conn.pool.Get()
	defer sess.Close()

	if _, err := sess.Do("PING"); err != nil {
		return err
	}
	return nil
}

func (conn Conn) Enqueue(data string, queue string) (*Job, error) {
	sess := conn.pool.Get()
	defer sess.Close()

	timeout := "1000"
	reply, err := sess.Do("ADDJOB", queue, data, timeout)
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

func (conn Conn) Dequeue(queue string, extra ...string) (*Job, error) {
	sess := conn.pool.Get()
	defer sess.Close()

	// Eh, the following doesn't build:
	//   reply, err := sess.Do("GETJOB", "FROM", queue, extra...)
	//   if err != nil {
	//		return nil, err
	//   }
	// I get "too many arguments in call to sess.Do" build error.
	// So I have to build the function arguments using reflect pkg.
	// <HACK>
	fn := reflect.ValueOf(sess.Do)
	args := []reflect.Value{reflect.ValueOf("GETJOB"), reflect.ValueOf("FROM"), reflect.ValueOf(queue)}
	for _, arg := range extra {
		args = append(args, reflect.ValueOf(arg))
	}
	ret := fn.Call(args)
	if len(ret) != 2 {
		return nil, errors.New("expected return value #1")
	}
	reply, ok := ret[0].Interface().(interface{})
	if !ok {
		return nil, errors.New("unexpected return value #2")
	}
	if !ret[1].IsNil() {
		err, ok := ret[1].Interface().(error)
		if !ok {
			return nil, errors.New("unexpected return value #3")
		}
		return nil, err
	}
	// </HACK>

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

func (conn Conn) Ack(job *Job) error {
	sess := conn.pool.Get()
	defer sess.Close()

	if _, err := sess.Do("ACKJOB", job.ID); err != nil {
		return err
	}
	return nil
}
