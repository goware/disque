# disque

[Golang](http://golang.org/) client for [Disque](https://github.com/antirez/disque), the Persistent Distributed Job Priority Queue.

- **Persistent** - Jobs can be either in-memory or persisted on disk<sup>[[1]](https://github.com/antirez/disque#disque-and-disk-persistence)</sup>.
- **Distributed** - Multiple producers, multiple consumers.
- **Job Priority Queue** - Multiple queues. Consumers Dequeue() from higher priority queues first.
- **Fault tolerant** - Jobs must be replicated to N nodes before Enqueue() returns. Jobs must be ACKed or they'll be re-queued automatically within a specified Retry Timeout.

[![GoDoc](https://godoc.org/github.com/goware/disque?status.png)](https://godoc.org/github.com/goware/disque)
[![Travis](https://travis-ci.org/goware/disque.svg?branch=master)](https://travis-ci.org/goware/disque)

**This project is in early development stage. You can expect changes to both functionality and the API. Feedback welcome!**

## Producer

```go
jobs, _ := disque.Connect("127.0.0.1:7711")

// Enqueue "high" priority job.
job1, _ := jobs.Add(data1, "high")

// Enqueue "low" priority jobs.
job2, _ := jobs.TTL(24 * time.Hour).Add(data2, "low")

// Enqueue "urgent" priority job. Re-queue if not ACKed within one minute.
job3, err := jobs.RetryAfter(time.Minute).Add(data3, "urgent")
if err != nil {
    jobs.Wait(job3)
}
```

## Consumer (worker)

```go
jobs, _ := disque.Connect("127.0.0.1:7711")

for {
    // Dequeue a job from highest priority queue (priority left to right).
    job, _ := jobs.Get("urgent", "high", "low")

    // Do some hard work with the job data.
    err := Process(job.Data)
    if err != nil {
        // Re-queue the job. This may be triggered by Panic/Recover.
        jobs.Nack(job)
    }

    // Acknowledge that we processed the job successfully.
    jobs.Ack(job)
}
```

## License
Disque is licensed under the [MIT License](./LICENSE).
