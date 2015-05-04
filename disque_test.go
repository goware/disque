package disque_test

import (
	"testing"

	"github.com/goware/disque"
)

func TestPriorityQueue(t *testing.T) {
	jobs, err := disque.Connect("127.0.0.1:7711")
	if err != nil {
		t.Fatal(err)
	}
	defer jobs.Close()

	if jobs.Ping() != nil {
		t.Fatal(err)
	}

	// Enqueue three jobs.
	_, err = jobs.Enqueue("data1", "test:low")
	if err != nil {
		t.Error(err)
	}
	_, err = jobs.Enqueue("data2", "test:urgent")
	if err != nil {
		t.Error(err)
	}
	_, err = jobs.Enqueue("data3", "test:high")
	if err != nil {
		t.Error(err)
	}

	// Dequeue first job.
	job, err := jobs.Dequeue("test:urgent", "test:high", "test:low")
	if err != nil {
		t.Error(err)
	}
	err = jobs.Ack(job)
	if err != nil {
		t.Error(err)
	}
	if e := "test:urgent"; job.Queue != e {
		t.Fatalf("expected %s, got %s", e, job.Queue)
	}
	if e := "data2"; job.Data != e {
		t.Fatalf("expected %s, got %s", e, job.Data)
	}

	// Dequeue second job.
	job, err = jobs.Dequeue("test:urgent", "test:high", "test:low")
	if err != nil {
		t.Error(err)
	}
	err = jobs.Ack(job)
	if err != nil {
		t.Error(err)
	}
	if e := "test:high"; job.Queue != e {
		t.Fatalf("expected %s, got %s", e, job.Queue)
	}
	if e := "data3"; job.Data != e {
		t.Fatalf("expected %s, got %s", e, job.Data)
	}

	// Dequeue third job.
	job, err = jobs.Dequeue("test:urgent", "test:high", "test:low")
	if err != nil {
		t.Error(err)
	}
	err = jobs.Ack(job)
	if err != nil {
		t.Error(err)
	}
	if e := "test:low"; job.Queue != e {
		t.Fatalf("expected %s, got %s", e, job.Queue)
	}
	if e := "data1"; job.Data != e {
		t.Fatalf("expected %s, got %s", e, job.Data)
	}

}
