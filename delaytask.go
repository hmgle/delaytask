package delaytask

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type Job struct {
	id    string
	fn    func()
	delay time.Duration
}

// Task defines a task.
type Task struct {
	tm          sync.Map
	waitCh      chan string
	quitCh      chan bool
	ctx         context.Context
	cancel      context.CancelFunc
	atomWaitCnt int64
}

// New will create a task.
func New() *Task {
	ctx, cancel := context.WithCancel(context.Background())
	t := &Task{
		tm:     sync.Map{},
		waitCh: make(chan string),
		quitCh: make(chan bool),
		ctx:    ctx,
		cancel: cancel,
	}
	go func() {
		wg := sync.WaitGroup{}
		for id := range t.waitCh {
			v, ok := t.tm.Load(id)
			if ok {
				t.tm.Delete(id)
				j := v.(*Job)
				wg.Add(1)
				go func() {
					j.fn()
					wg.Done()
				}()
			}
		}
		wg.Wait()
		t.quitCh <- true
	}()
	return t
}

// AddJob add a delay job task.
func (t *Task) AddJob(j *Job) {
	if j.delay > 0 {
		atomic.AddInt64(&t.atomWaitCnt, 1)
		go func(ctx context.Context) {
			defer atomic.AddInt64(&t.atomWaitCnt, -1)
			select {
			case <-time.After(j.delay):
				t.waitCh <- j.id
			case <-ctx.Done():
				t.waitCh <- j.id
			}
		}(t.ctx)
	}
	t.tm.Store(j.id, j)
}

// AddJobFn add a delay job by func to task.
func (t *Task) AddJobFn(id string, fn func(), delay ...time.Duration) {
	j := &Job{
		id: id,
		fn: fn,
	}
	if len(delay) > 0 {
		j.delay = delay[0]
		atomic.AddInt64(&t.atomWaitCnt, 1)
		go func(ctx context.Context) {
			defer atomic.AddInt64(&t.atomWaitCnt, -1)
			select {
			case <-time.After(j.delay):
				t.waitCh <- id
			case <-ctx.Done():
				t.waitCh <- id
			}
		}(t.ctx)
	}
	t.tm.Store(id, j)
}

// Execute the job immediately.
func (t *Task) Execute(id string) {
	t.waitCh <- id
}

// Stop the task.
func (t *Task) Stop() <-chan bool {
	t.cancel()
	for {
		if atomic.LoadInt64(&t.atomWaitCnt) == 0 {
			break
		}
		time.Sleep(time.Millisecond * 10)
	}
	close(t.waitCh)
	return t.quitCh
}
