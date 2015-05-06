# disque

[Golang](http://golang.org/) client for [Disque](https://github.com/antirez/disque), the Persistent Distributed Job Priority Queue.

- **Persistent** - Jobs can be either in-memory or persisted on disk<sup>[[1]](https://github.com/antirez/disque#disque-and-disk-persistence)</sup>.
- **Distributed** - Disque pool. Multiple producers, multiple consumers.
- **Job Priority Queue** - Multiple queues. Consumers `Get()` from higher priority queues first.
- **Fault tolerant** - Jobs must be replicated to N nodes before `Add()` returns. Consumer must `Ack()` the job within a specified `RetryAfter` timeout or the job will be re-queued automatically.

[![GoDoc](https://godoc.org/github.com/goware/disque?status.png)](https://godoc.org/github.com/goware/disque)
[![Travis](https://travis-ci.org/goware/disque.svg?branch=master)](https://travis-ci.org/goware/disque)

*Note: The examples below ignore error handling for readability.*

## Producer

```go
import (
    "github.com/goware/disque"
)

func main() {
    // Connect to Disque pool.
    jobs, _ := disque.New("127.0.0.1:7711") // Accepts more arguments.

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
    jobs, _ := disque.New("127.0.0.1:7711") // Accepts more arguments.

    for {
        // Get job from highest priority queue possible. Blocks by default.
        job, _ := jobs.Get("urgent", "high", "low") // Left-to-right priority.

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

## Default configuration

| Config option | Default value | Description  |
| ------------- |:-------------:| ------------ |
| Timeout       | 0             | Block on each operation until it returns. |
| Replicate     | 0             | Job doesn't need to be replicated before Add() returns. |
| Delay         | 0             | Job is enqueued immediately. |
| RetryAfter    | 0             | Don't re-queue job automatically. |
| TTL           | 0             | Job lives until it's ACKed. |
| MaxLen        | 0             | Unlimited queue. |

## Custom configuration

```go
jobs, _ := disque.New("127.0.0.1:7711")

config := disque.Config{
    Timeout:    time.Second,    // Each operation fails after 1s timeout elapses.
    Replicate:  2,              // Replicates job to 2+ nodes before Add() returns.
    Delay:      time.Hour,      // Schedules the job (enqueues after 1h).
    RetryAfter: time.Minute,    // Re-queues the job after 1min of not being ACKed.
    TTL:        24 * time.Hour, // Removes the job from the queue after one day.
    MaxLen:     1000,           // Fails if there are 1000+ jobs in the queue.
}

// Apply globally.
jobs.Use(config)

// Apply to a single operation.
jobs.With(config).Add(data, "queue")

// Apply single option to a single operation.
jobs.Timeout(time.Second).Get("queue", "queue2")
jobs.MaxLen(1000).RetryAfter(time.Minute).Add(data, "queue")
jobs.TTL(24 * time.Hour).Add(data, "queue")
```

## License
Disque is licensed under the [MIT License](./LICENSE).
