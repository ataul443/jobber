## Basic Example

Concrete type which satisfies the jobber.Job interface.
We will publish these jobs to workers

````go
type BaseJob struct {
	id       int
	category string
	ctx      context.Context
}

type numGenJob struct {
	BaseJob
	count int
}

type greetJob struct {
	BaseJob
	name string
}
````

We will create our job manager

````go
jobWaitSize := 2
jbManager := jobber.New(jobWaitSize)
````

Here are our workers ready to take a pending job and process it.
````go
//NumberGeneratorWorker generates and print exactly
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
````

Lets create job portals for wrokers, jbManager.Subscribe() will
return a channel of type *jobber.PendingJob
````go
numGenJobCh := jbManager.Subscribe(numGenJobType)
greetJobCh := jbManager.Subscribe(greetJobType)
````

Spin up the workers to get some processing done
````go
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
````

Now its time to publish some jobs into the job portals
````go
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
````

Note: While testing the code, I encounterd some race bugs in Jobber.Close() which still needs to be addressed.