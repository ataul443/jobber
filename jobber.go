// Jobber is a basic job manager aims to deliver specific jobs
// to specific subscribers waiting for those jobs
package jobber

import (
	"fmt"
	"sync"
)

const (
	errJobExpired = "job timed out"
	errNoSubs     = "no subscriber found"

	errJobNotPublished = "job not published"
	errJobAlreadyDone  = "job already done"
)

type Job interface {
	// Type returns the category of the job.
	Category() string

	// Expired returns true if job time out, otherwise returns false.
	Expired() bool
}

type Jobber interface {
	// Publish dispatches a job to all the subcribers
	// of this job category.
	// It will return error if no subscribers found.
	Publish(jb Job) error

	// Subscribe returns a channel on which a subscriber
	// listen for a particular category of job
	Subscribe(jbCat string) <-chan *PendingJob

	// CancelSubs cancels all the subscription on this job category
	CancelSubs(jbCat string)

	// Close cancels all subscriptions of all category of job
	Close()

	// Status returns a channel which emits all the processed jobs
	Status() <-chan *PendingJob
}

// New returns a jobber object with per ctegory job queue size
// equal to catQueueSize.
func New(catQueueSize int) Jobber {
	return &jobber{sync.RWMutex{},
		make(map[string]chan *PendingJob),
		make(chan *PendingJob, catQueueSize),
		catQueueSize,
	}
}

// JobFailed returns true if job fails.
func (jb *PendingJob) JobFailed() (bool, error) {
	return jb.status(), nil
}

// MarkJobFailed marks a job as failed.
func (jb *PendingJob) MarkJobFailed() error {
	return jb.markFailed()
}

// MarkJobDone marks a job as done.
func (jb *PendingJob) MarkJobDone() error {
	return jb.markDone()
}

// Payload returns underlying job payload
func (jb *PendingJob) Payload() Job {
	return jb.Job
}

type PendingJob struct {
	Job
	failed bool
	done   bool
	notify chan<- *PendingJob
}

func (jb *PendingJob) status() bool {
	return jb.failed
}

func (jb *PendingJob) markFailed() error {
	if jb.done {
		return fmt.Errorf("%v", errJobAlreadyDone)
	}
	jb.failed = true
	jb.notify <- jb
	return nil
}

func (jb *PendingJob) markDone() error {
	if jb.done {
		return fmt.Errorf("%v", errJobAlreadyDone)
	}
	jb.failed = false
	jb.done = true
	jb.notify <- jb
	return nil
}

type jobber struct {
	mu            sync.RWMutex
	subs          map[string]chan *PendingJob
	status        chan *PendingJob
	cateQueueSize int
}

func (j *jobber) Publish(jb Job) error {
	j.mu.RLock()
	defer j.mu.RUnlock()

	jbCat := jb.Category()

	subscriber, ok := j.subs[jbCat]
	if !ok {
		// No subscriber is present yet for this job
		return fmt.Errorf("%v", errNoSubs)
	}

	if jb.Expired() {
		return fmt.Errorf("%v", errJobExpired)
	}

	if len(subscriber) == j.cateQueueSize {
		return fmt.Errorf("job capacity for this category is full")
	}

	// The subscriber channel should be buffered
	// else it will get block if the worker is not
	// yet ready for the job
	subscriber <- &PendingJob{jb, false, false, j.status}

	return nil

}

func (j *jobber) Subscribe(jbCat string) <-chan *PendingJob {
	j.mu.Lock()
	defer j.mu.Unlock()

	subscriber := make(chan *PendingJob, j.cateQueueSize)
	j.subs[jbCat] = subscriber
	return subscriber
}

func (j *jobber) CancelSubs(jbCat string) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	subscriber, ok := j.subs[jbCat]
	if !ok {
		return
	}

	close(subscriber)
	delete(j.subs, jbCat)

}

func (j *jobber) Close() {
	j.mu.Lock()
	defer j.mu.Unlock()

	for _, subscriber := range j.subs {
		close(subscriber)
	}

	close(j.status)

}

func (j *jobber) Status() <-chan *PendingJob {
	return j.status
}
