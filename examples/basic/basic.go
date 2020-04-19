package main

import (
	"context"
	"fmt"
	"github.com/ataul443/jobber"
	"math/rand"
	"sync"
)

type BaseJob struct {
	id       int
	category string
	ctx      context.Context
}

func (bj BaseJob) ID() int {
	return bj.id
}

func (bj BaseJob) Category() string {
	return bj.category
}

func (bj BaseJob) Expired() bool {
	select {
	case <-bj.ctx.Done():
		// Job expired
		return true
	default:
		return false
	}
}

type numGenJob struct {
	BaseJob
	count int
}

func (ng numGenJob) Count() int {
	return ng.count
}

type greetJob struct {
	BaseJob
	name string
}

func (gj greetJob) Name() string {
	return gj.name
}

// NumberGeneratorWorker generates and print exactly
// count number of random numbers every 2 seonds
func NumberGeneratorWorker(pjb *jobber.PendingJob) {
	numJob := pjb.Payload().(numGenJob)

	count := numJob.Count()

	for i := 0; i < count; i++ {
		fmt.Printf("Random number: %d\n", rand.Int())
	}
}

// GreetJobWorker greetes a person
func GreetJobWorker(pjb *jobber.PendingJob) {
	gj := pjb.Payload().(greetJob)
	fmt.Printf("Hello, %s\n", gj.Name())
}

func main() {
	jobWaitSize := 2
	jbManager := jobber.New(jobWaitSize)

	// Subscribe for jobs

	numGenJobType := "NUM_GEN_JOB"
	greetJobType := "GREET_JOB"

	numGenJobCh := jbManager.Subscribe(numGenJobType)
	greetJobCh := jbManager.Subscribe(greetJobType)

	var wg sync.WaitGroup
	wg.Add(2)

	// Number Genrator worker
	go func(jobCh <-chan *jobber.PendingJob) {
		defer wg.Done()
		for jb := range jobCh {
			NumberGeneratorWorker(jb)
		}
	}(numGenJobCh)

	// Greet Job Worker
	go func(jobCh <-chan *jobber.PendingJob) {
		defer wg.Done()
		for jb := range jobCh {
			GreetJobWorker(jb)
		}
	}(greetJobCh)



	ctx := context.Background()
	err := jbManager.Publish(numGenJob{BaseJob{1, numGenJobType, ctx}, 5})
	if err != nil {
		fmt.Printf("err, %v", err.Error())
	}


	err = jbManager.Publish(greetJob{BaseJob{2, greetJobType, ctx}, "Ataul"})
	if err != nil {
		fmt.Printf("err, %v", err.Error())
	}
	jbManager.Close()
	wg.Wait()


}
