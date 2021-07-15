package delaytask

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// Job defines a job.
type Job struct {
	ID    string
	Fn    func()
	Delay time.Duration

	timer *time.Timer
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
					j.Fn()
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
	if j.Delay > 0 {
		j.timer = time.NewTimer(j.Delay)
		atomic.AddInt64(&t.atomWaitCnt, 1)
		go func(ctx context.Context) {
			defer atomic.AddInt64(&t.atomWaitCnt, -1)
			select {
			case <-j.timer.C:
				t.waitCh <- j.ID
			case <-ctx.Done():
				t.waitCh <- j.ID
			}
		}(t.ctx)
	}
	t.tm.Store(j.ID, j)
}

// AddJobFn add a delay job by func to task.
func (t *Task) AddJobFn(id string, fn func(), delay ...time.Duration) {
	j := &Job{
		ID: id,
		Fn: fn,
	}
	if len(delay) > 0 {
		j.Delay = delay[0]
		j.timer = time.NewTimer(j.Delay)
		atomic.AddInt64(&t.atomWaitCnt, 1)
		go func(ctx context.Context) {
			defer atomic.AddInt64(&t.atomWaitCnt, -1)
			select {
			case <-j.timer.C:
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

// Cancel the job.
func (t *Task) Cancel(id string) {
	t.tm.Delete(id)
}

// Reset changes the job's timer to expire after duration d.
// It returns true if the timer had been active, false if the timer had
// expired or been stopped.
func (t *Task) Reset(id string, d time.Duration) bool {
	v, ok := t.tm.Load(id)
	if !ok {
		return false
	}
	j := v.(*Job)
	if j.timer != nil {
		return j.timer.Reset(d)
	}
	return false
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
