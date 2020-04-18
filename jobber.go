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
	Subscribe(jbCat string) <-chan Job

	// CancelSubs cancels all the subscription on this job category
	CancelSubs(jbCat string)

	// Close cancels all subscriptions of all category of job
	Close()

	// Status returns a channel which emits all the processed jobs
	Status() <-chan Job
}

// New returns a jobber object with per ctegory job queue size
// equal to catQueueSize.
func New(catQueueSize int) Jobber {
	return &jobber{sync.RWMutex{},
		make(map[string]chan Job),
		make(chan Job),
		catQueueSize,
	}
}

// JobFailed returns true if job fails.
func JobFailed(jb Job) (bool, error) {
	switch jb.(type) {
	case *job:
		return jb.(*job).status(), nil
	default:
		return false, fmt.Errorf(errJobNotPublished)
	}
}

// MarkJobFailed marks a job as failed.
func MarkJobFailed(jb interface{}) error {
	switch jb.(type) {
	case *job:
		return jb.(*job).markFailed()

	default:
		return fmt.Errorf(errJobNotPublished)
	}
}

// MarkJobDone marks a job as done.
func MarkJobDone(jb interface{}) error {
	switch jb.(type) {
	case *job:
		return jb.(*job).markDone()
	default:
		return fmt.Errorf(errJobNotPublished)
	}
}

type job struct {
	Job
	failed bool
	done   bool
	notify chan<- Job
}

func (jb job) status() bool {
	return jb.failed
}

func (jb *job) markFailed() error {
	if jb.done {
		return fmt.Errorf("%v", errJobAlreadyDone)
	}
	jb.failed = true
	jb.notify <- jb
	return nil
}

func (jb *job) markDone() error {
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
	subs          map[string]chan Job
	status        chan Job
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
	subscriber <- &job{jb, false, false, j.status}

	return nil

}

func (j *jobber) Subscribe(jbCat string) <-chan Job {
	j.mu.Lock()
	defer j.mu.Unlock()

	subscriber := make(chan Job, j.cateQueueSize)
	j.subs[jbCat] = subscriber
	return subscriber
}

func (j *jobber) CancelSubs(jbCat string) {
	j.mu.RLock()
	defer j.mu.Unlock()

	subscriber, ok := j.subs[jbCat]
	if !ok {
		return
	}

	close(subscriber)

}

func (j *jobber) Close() {
	j.mu.Lock()
	defer j.mu.Unlock()

	for _, subscriber := range j.subs {
		close(subscriber)
	}

	close(j.status)

}

func (j *jobber) Status() <-chan Job {
	return j.status
}
