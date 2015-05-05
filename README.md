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

// Enqueue some jobs.
job1, _ := jobs.Add(data1, "low")
job2, _ := jobs.Add(data2, "urgent")
job3, _ := jobs.Add(data3, "high")
```

## Consumer (worker)

```go
jobs, _ := disque.Connect("127.0.0.1:7711")

for {
    // Dequeue a job (from higher priority queues first).
    job, _ := jobs.Get("urgent", "high", "low")

    // Do some hard work with the job data.
    err := Process(job.Data)
    if err != nil {
        // Re-queue job.
        jobs.Nack(job)
    }

    // Acknowledge that we processed the job successfully.
    jobs.Ack(job)
}
```

## License
Disque is licensed under the [MIT License](./LICENSE).
