package disque_test

import (
	"testing"
	"time"

	"disque"
)

func TestPing(t *testing.T) {
	// Connect to Disque.
	jobs, err := disque.New("127.0.0.1:7711")
	if err != nil {
		t.Fatal(err)
	}
	defer jobs.Close()

	// Ping.
	if jobs.Ping() != nil {
		t.Fatal(err)
	}
}

func TestFetch(t *testing.T) {
	// Connect to Disque.
	jobs, err := disque.New("127.0.0.1:7711")
	if err != nil {
		t.Fatal(err)
	}
	defer jobs.Close()

	// Add.
	job, err := jobs.Add("data0x5f3759df", "test:fetch")
	if err != nil {
		t.Error(err)
	}

	// Fetch.
	_job, err := jobs.Fetch(job.ID)
	if err != nil {
		t.Error(err)
	}

	// Job should have been fetched and content of two jobs should be equal
	if _job == nil {
		t.Fatal("expected job to be fetched")
	}
	if job.Data != _job.Data {
		t.Error("expected data of the jobs to be equal")
	}

	jobs.Ack(_job)
}

func TestDelay(t *testing.T) {
	// Connect to Disque.
	jobs, err := disque.New("127.0.0.1:7711")
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
	_, err = jobs.Timeout(time.Millisecond).Get("test:delay")
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
	jobs, err := disque.New("127.0.0.1:7711")
	if err != nil {
		t.Fatal(err)
	}
	defer jobs.Close()

	// Enqueue job with TTL one second.
	_, err = jobs.Timeout(time.Millisecond).TTL(time.Second).Add("data1", "test:ttl")
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
	jobs, err := disque.New("127.0.0.1:7711")
	if err != nil {
		t.Fatal(err)
	}
	defer jobs.Close()

	// Enqueue job with retry after one second.
	_, err = jobs.RetryAfter(time.Second).Add("data1", "test:retry")
	if err != nil {
		t.Error(err)
	}

	// Get the job.
	_, err = jobs.Get("test:retry")
	if err != nil {
		t.Fatal(err)
	}

	// Don't Ack() to pretend consumer failure.

	// Try to get the job again..
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
	jobs, err := disque.New("127.0.0.1:7711")
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

	// Get first job.
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

	// Get second job.
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

	// Get third job and re-queue it again.
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

	// Get third job again.
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
	jobs, err := disque.New("127.0.0.1:7711")
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
		// Get the job.
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

func TestConfig(t *testing.T) {
	// Connect to Disque.
	jobs, err := disque.New("127.0.0.1:7711")
	if err != nil {
		t.Fatal(err)
	}
	defer jobs.Close()

	config := disque.Config{
		Timeout: time.Millisecond,
	}

	// Should fail on timeout.
	_, err = jobs.With(config).Get("test:non-existant-queue")
	if err == nil {
		t.Fatal("expected error")
	}

	// Should fail on timeout.
	jobs.Use(config)
	_, err = jobs.Get("test:non-existant-queue")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestQueueLength(t *testing.T) {
	// Connect to Disque.
	jobs, err := disque.New("127.0.0.1:7711")
	if err != nil {
		t.Fatal(err)
	}
	defer jobs.Close()

	var length int

	// Make sure 0 jobs are enqueued.
	length, err = jobs.Len("test:len")
	if err != nil {
		t.Error(err)
	}
	if length != 0 {
		t.Fatalf("unexpected length %v", length)
	}

	// Enqueue hundred jobs.
	for i := 0; i < 100; i++ {
		_, err = jobs.Add("data1", "test:len")
		if err != nil {
			t.Error(err)
		}
	}

	// Make sure 100 jobs are enqueued.
	length, err = jobs.Len("test:len")
	if err != nil {
		t.Error(err)
	}
	if length != 100 {
		t.Error("unexpected length %v", length)
	}

	// Dequeue hundred jobs.
	for i := 0; i < 100; i++ {
		job, err := jobs.Get("test:len")
		if err != nil {
			t.Error(err)
		}
		jobs.Ack(job)
	}

	// Make sure 0 jobs are enqueued.
	length, err = jobs.Len("test:len")
	if err != nil {
		t.Error(err)
	}
	if length != 0 {
		t.Error("unexpected length %v", length)
	}
}
