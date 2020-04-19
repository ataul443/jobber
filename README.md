# Jobber

Jobber is a simple library for dispatching _jobs_ to
_workers_ for processing. It is based on Pub/Sub pattern where workers can
subscribed to a channel to take on a particular of their interest.

## Job

Jobber can publish any implementation of job to their subscribers as long
as the underlying type of a job implements below interface.

```go
type Job interface {
    // Type returns the category of the job.
    Category() string

    // Expired returns true if job time out, otherwise returns false.
    Expired() bool
}
```

## Examples

Basic example can be found [here](https://github.com/ataul443/jobber/tree/master/examples/basic).


