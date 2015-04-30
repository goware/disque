# pjobs

Persistent Distributed Job Priority Queue for [golang](http://golang.org/) powered by [Disque](https://github.com/antirez/disque).

- **Persistent** - Jobs can be persisted on disk.
- **Distributed** - Multiple producers, multiple consumers.
- **Job Priority Queue** - Consumers dequeue "high" priority jobs first.
- **Tolerant to consumer failures** - Jobs are requeued automatically if not ACKed within the Retry Timeout.

[![GoDoc](https://godoc.org/github.com/goware/pjobs?status.png)](https://godoc.org/github.com/goware/pjobs)
[![Travis](https://travis-ci.org/goware/pjobs.svg?branch=master)](https://travis-ci.org/goware/pjobs)

**This project is in early development stage. You can expect changes to both functionality and the API. Feedback welcome!**

## Disque

Install & run [Disque](https://github.com/antirez/disque) server.

*TODO: Explain how to enable disk persistence.*

## Producers

```go
// Connect to Disque server.
jobs, _ := pjobs.Connect("127.0.0.1:7711")

// Enqueue job (data + priority).
job, _ := jobs.Enqueue("data", "low")
job, _ := jobs.Enqueue("data", "high")
```

## Consumers

```go
// Connect to Disque server.
jobs, _ := pjobs.Connect("127.0.0.1:7711")

// Dequeue job ("high" priority jobs first).
job, _ := jobs.Dequeue("high", "low")

// Do some hard work with job.Data.

// Acknowledge that job was processed successfully.
jobs.Ack(job)
```

## License
Pjobs is licensed under the [MIT License](./LICENSE).
