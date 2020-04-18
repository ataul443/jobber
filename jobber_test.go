package jobber

import (
	"testing"
)

type myJob struct {
	category string
}

func (mj *myJob) Category() string {
	return mj.category
}

func (my *myJob) Expired() bool {
	return false
}

func TestPublishWithNoSubscriber(t *testing.T) {
	jbManager := New(5)

	err := jbManager.Publish(&myJob{"IMAGE"})
	if err.Error() != errNoSubs {
		t.Fatalf("err want %v, got %v", errNoSubs, err)
	}

	jbManager.Close()
}

func TestPublishWithSubscriber(t *testing.T) {
	jbManager := New(5)

	_ = jbManager.Subscribe("IMAGE")

	err := jbManager.Publish(&myJob{"IMAGE"})
	if err != nil {
		t.Fatalf("err want %v, got %v", nil, err)
	}

	jbManager.Close()
}

func TestJobShouldFailed(t *testing.T) {
	jbManager := New(5)

	s := jbManager.Subscribe("IMAGE")

	err := jbManager.Publish(&myJob{"IMAGE"})
	if err != nil {
		t.Fatalf("err want %v, got %v", nil, err)
	}

	go func() {
		k := <-s
		_ = MarkJobFailed(k)
	}()

	status := jbManager.Status()
	select {
	case jb := <-status:
		f, _ := JobFailed(jb)
		if !f {
			t.Fatalf("job should failed!")
		}

	}

	jbManager.Close()
}

func TestJobShouldNotFailed(t *testing.T) {
	jbManager := New(5)

	s := jbManager.Subscribe("IMAGE")

	err := jbManager.Publish(&myJob{"IMAGE"})
	if err != nil {
		t.Fatalf("err want %v, got %v", nil, err)
	}

	go func() {
		k := <-s
		MarkJobDone(k)
	}()

	status := jbManager.Status()
	select {
	case jb := <-status:
		f, _ := JobFailed(jb)
		if f {
			t.Fatalf("job should not failed!")
		}

	}

	jbManager.Close()
}

func TestJobCapacityFull(t *testing.T) {
	jbManager := New(5)

	_ = jbManager.Subscribe("IMAGE")

	for i := 0; i < 6; i++ {
		err := jbManager.Publish(&myJob{"IMAGE"})
		if err != nil && i < 5 {
			t.Fatalf("err want %v, got %v", nil, err)
		}
	}

	jbManager.Close()

}
