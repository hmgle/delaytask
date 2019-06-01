package delaytask_test

import (
	"log"
	"testing"
	"time"

	"github.com/hmgle/delaytask"
)

var tsk = delaytask.New()

func TestDelayTaskAdd(t *testing.T) {
	tsk.AddJobFn("cancel", func() {
		log.Printf("my ID: cancel\n")
	}, time.Millisecond*50)

	tsk.AddJobFn("foo", func() {
		log.Printf("my ID: foo\n")
		time.Sleep(time.Millisecond * 500)
		log.Printf("foo done!\n")
	}, time.Millisecond*100)

	tsk.AddJobFn("bar", func() {
		log.Printf("my ID: bar\n")
		time.Sleep(time.Millisecond * 100)
		log.Printf("bar done!\n")
	}, time.Millisecond*250)
}

func TestTaskExe(t *testing.T) {
	tsk.Cancel("cancel")
	tsk.Execute("foo")
	tsk.Execute("no this id")
	time.Sleep(time.Millisecond * 50)
	<-tsk.Stop()
}
