# delaytask

`delaytask` is a library which provides a way to run the delay task.

## usage

```go
var tsk = delaytask.New()

tsk.AddJobFn("foo", func() {
	println("foo")
}, time.Second)

// tsk.Execute("foo")

<-tsk.Stop()
```
