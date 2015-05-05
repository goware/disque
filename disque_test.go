package disque_test

import (
	"testing"
	"time"

	"github.com/goware/disque"
)

func TestPing(t *testing.T) {
	// Connect to Disque.
	jobs, err := disque.Connect("127.0.0.1:7711")
	if err != nil {
		t.Fatal(err)
	}
	defer jobs.Close()

	// Ping.
	if jobs.Ping() != nil {
		t.Fatal(err)
	}
}

func TestDelay(t *testing.T) {
	// Connect to Disque.
	jobs, err := disque.Connect("127.0.0.1:7711")
	if err != nil {
		t.Fatal(err)
	}
	defer jobs.Close()

	// Enqueue job after one second.
	_, err = jobs.Delay(time.Second).Add("data1", "test:delay")
	if err != nil {
		t.Error(err)
	}

	// The job should not exist yet.
	_, err = jobs.Get("test:delay")
	if err == nil {
		t.Fatal("expected error")
	}

	time.Sleep(time.Second)

	// The job should exist now.
	job, err := jobs.Timeout(500 * time.Millisecond).Get("test:delay")
	if err != nil {
		t.Fatal(err)
	}

	jobs.Ack(job)
}

func TestTTL(t *testing.T) {
	// Connect to Disque.
	jobs, err := disque.Connect("127.0.0.1:7711")
	if err != nil {
		t.Fatal(err)
	}
	defer jobs.Close()

	// Enqueue job with TTL one second.
	_, err = jobs.TTL(time.Second).Add("data1", "test:ttl")
	if err != nil {
		t.Error(err)
	}

	time.Sleep(1500 * time.Millisecond)

	// The job should no longer exist.
	_, err = jobs.Get("test:ttl")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestTimeoutRetryAfter(t *testing.T) {
	// Connect to Disque.
	jobs, err := disque.Connect("127.0.0.1:7711")
	if err != nil {
		t.Fatal(err)
	}
	defer jobs.Close()

	// Enqueue job with retry after one second.
	_, err = jobs.RetryAfter(time.Second).Add("data1", "test:retry")
	if err != nil {
		t.Error(err)
	}

	// Dequeue job.
	_, err = jobs.Get("test:retry")
	if err != nil {
		t.Fatal(err)
	}

	// Don't Ack() to pretend consumer failure.

	// Try to dequeue job again..
	// We should hit time-out for the first time..
	_, err = jobs.Timeout(250 * time.Millisecond).Get("test:retry")
	if err == nil {
		t.Fatal("expected error")
	}
	// and we should be successful for the second time..
	job, err := jobs.Timeout(time.Second).Get("test:retry")
	if err != nil {
		t.Fatal(err)
	}

	// Ack the job.
	err = jobs.Ack(job)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPriorityQueue(t *testing.T) {
	// Connect to Disque.
	jobs, err := disque.Connect("127.0.0.1:7711")
	if err != nil {
		t.Fatal(err)
	}
	defer jobs.Close()

	// Enqueue three jobs.
	_, err = jobs.Add("data1", "test:low")
	if err != nil {
		t.Error(err)
	}
	_, err = jobs.Add("data2", "test:urgent")
	if err != nil {
		t.Error(err)
	}
	_, err = jobs.Add("data3", "test:high")
	if err != nil {
		t.Error(err)
	}

	// Dequeue first job.
	job, err := jobs.Get("test:urgent", "test:high", "test:low")
	if err != nil {
		t.Fatal(err)
	}
	err = jobs.Ack(job)
	if err != nil {
		t.Fatal(err)
	}
	if e := "test:urgent"; job.Queue != e {
		t.Fatalf("expected %s, got %s", e, job.Queue)
	}
	if e := "data2"; job.Data != e {
		t.Fatalf("expected %s, got %s", e, job.Data)
	}

	// Dequeue second job.
	job, err = jobs.Get("test:urgent", "test:high", "test:low")
	if err != nil {
		t.Fatal(err)
	}
	err = jobs.Ack(job)
	if err != nil {
		t.Fatal(err)
	}
	if e := "test:high"; job.Queue != e {
		t.Fatalf("expected %s, got %s", e, job.Queue)
	}
	if e := "data3"; job.Data != e {
		t.Fatalf("expected %s, got %s", e, job.Data)
	}

	// Dequeue third job and re-queue it again.
	job, err = jobs.Get("test:urgent", "test:high", "test:low")
	if err != nil {
		t.Fatal(err)
	}
	err = jobs.Nack(job)
	if err != nil {
		t.Fatal(err)
	}
	if e := "test:low"; job.Queue != e {
		t.Fatalf("expected %s, got %s", e, job.Queue)
	}
	if e := "data1"; job.Data != e {
		t.Fatalf("expected %s, got %s", e, job.Data)
	}

	// Dequeue third job again.
	job, err = jobs.Get("test:urgent", "test:high", "test:low")
	if err != nil {
		t.Fatal(err)
	}
	err = jobs.Ack(job)
	if err != nil {
		t.Fatal(err)
	}
	if e := "test:low"; job.Queue != e {
		t.Fatalf("expected %s, got %s", e, job.Queue)
	}
	if e := "data1"; job.Data != e {
		t.Fatalf("expected %s, got %s", e, job.Data)
	}
}

func TestWait(t *testing.T) {
	// Connect to Disque.
	jobs, err := disque.Connect("127.0.0.1:7711")
	if err != nil {
		t.Fatal(err)
	}
	defer jobs.Close()

	// Enqueue job.
	start := time.Now()
	job, err := jobs.Add("data1", "test:wait")
	if err != nil {
		t.Error(err)
	}

	go func() {
		// Dequeue the job.
		job, err := jobs.Get("test:wait")
		if err != nil {
			t.Fatal(err)
		}

		// Sleep for 1 second before ACK.
		time.Sleep(time.Second)
		jobs.Ack(job)

	}()

	// Wait for the job to finish. Should take more than 1 second.
	jobs.Wait(job)
	duration := time.Since(start)
	if duration < time.Second || duration > 1500*time.Millisecond {
		t.Fatalf("expected 1.0s - 1.5s, got %v", time.Since(start))
	}
}
