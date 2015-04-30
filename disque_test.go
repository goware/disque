package pjobs_test

import (
	"testing"

	"github.com/goware/pjobs"
)

func TestPriorityQueue(t *testing.T) {
	jobs, err := pjobs.Connect("127.0.0.1:7711")
	if err != nil {
		t.Fatal(err)
	}
	defer jobs.Close()

	if jobs.Ping() != nil {
		t.Fatal(err)
	}

	// Enqueue two jobs.
	_, err = jobs.Enqueue(`{"request":"data"}`, "test:low")
	if err != nil {
		t.Fatal(err)
	}
	_, err = jobs.Enqueue(`{"request":"data"}`, "test:high")
	if err != nil {
		t.Fatal(err)
	}

	// Dequeue first job. Must be from high priority queue.
	job1, err := jobs.Dequeue("test:high", "test:low")
	if err != nil {
		t.Fatal(err)
	}
	if job1.Queue != "test:high" {
		t.Fatal("unexpected priority")
	}
	err = jobs.Ack(job1)
	if err != nil {
		t.Fatal(err)
	}

	// Dequeue second job. Must be from low priority queue.
	job2, err := jobs.Dequeue("test:high", "test:low")
	if err != nil {
		t.Fatal(err)
	}
	if job2.Queue != "test:low" {
		t.Fatal("unexpected priority")
	}
	err = jobs.Ack(job2)
	if err != nil {
		t.Fatal(err)
	}
}
