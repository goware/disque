# disque

[Golang](http://golang.org/) client for [Disque](https://github.com/antirez/disque), the Persistent Distributed Job Priority Queue.

- **Persistent** - Jobs can be either in-memory or persisted on disk<sup>[[1]](https://github.com/antirez/disque#disque-and-disk-persistence)</sup>.
- **Distributed** - Disque pool. Multiple producers, multiple consumers.
- **Job Priority Queue** - Multiple queues. Consumers `Get()` from higher priority queues first.
- **Fault tolerant** - Jobs must be replicated to N nodes before `Add()` returns. Jobs must be `ACK()`ed after `Get()` or they'll be re-queued automatically within a specified `RetryAfter` timeout.

[![GoDoc](https://godoc.org/github.com/goware/disque?status.png)](https://godoc.org/github.com/goware/disque)
[![Travis](https://travis-ci.org/goware/disque.svg?branch=master)](https://travis-ci.org/goware/disque)

## Producer

```go
import (
    "github.com/goware/disque"
)

func main() {
    // Connect to Disque pool.
    jobs, _ := disque.Connect("127.0.0.1:7711") // Accepts more arguments.

    // Enqueue three jobs with different priorities.
    job1, _ := jobs.Add(data1, "high")
    job2, _ := jobs.Add(data2, "low")
    job3, _ := jobs.Add(data3, "urgent")

    // Block until job3 is done.
    jobs.Wait(job3)
}
```

## Consumer (worker)

```go
import (
    "github.com/goware/disque"
)

func main() {
    // Connect to Disque pool.
    jobs, _ := disque.Connect("127.0.0.1:7711") // Accepts more arguments.

    for {
        // Get job from highest priority queue possible. Blocks by default.
        job, _ := jobs.Get("urgent", "high", "low") // Left-right priority.

        // Do some hard work with the job data.
        if err := Process(job.Data); err != nil {
            // Failed. Re-queue the job.
            jobs.Nack(job)
        }

        // Acknowledge (dequeue) the job.
        jobs.Ack(job)
    }
}
```

## Config (Timeout, Replicate, Delay, Retry, TTL, MaxLen)

```go
jobs, _ := disque.Connect("127.0.0.1:7711")

config := disque.Config{
    Timeout:    time.Second,    // Each operation will fail after 1s. It blocks by default.
    Replicate:  2,              // Add(): Replicate job to at least two nodes before return.
    Delay:      time.Hour,      // Add(): Schedule the job - enqueue after one hour.
    RetryAfter: time.Minute,    // Add(): Re-queue job after 1min (time between Get() and Ack()).
    TTL:        24 * time.Hour, // Add(): Remove the job from queue after one day.
    MaxLen:     1000,           // Add(): Fail if there are more than 1000 jobs in the queue.
}

// Apply globally.
jobs.Use(config)

// Apply to a single operation.
jobs.With(config).Add(data, "queue")

// Apply single option to a single operation.
jobs.Timeout(time.Second).Get("queue", "queue2")
jobs.MaxLen(1000).RetryAfter(time.Minute).Add(data, "queue")
jobs.Timeout(time.Second).Add(data, "queue")
```

## License
Disque is licensed under the [MIT License](./LICENSE).
