# TimingWheel模块使用指南

`timingwheel`模块是一个用于管理定时任务的模块，它提供了一种高效的方式来处理大量的定时任务。这个模块的主要组件是`TimingWheel`结构体，它提供了一系列的方法来添加、执行和取消定时任务。

## 初始化TimingWheel

首先，你需要创建一个`TimingWheel`实例。你可以使用`NewTimingWheel`函数来创建一个新的实例，这个函数接受两个参数：`tick`和`wheelSize`。

```go
tw := timingwheel.NewTimingWheel(1*time.Second, 30)
```

这里，`tick`参数表示时间轮的最小时间单位，`wheelSize`参数表示时间轮的大小。

## 启动TimingWheel

创建了`TimingWheel`实例之后，你需要调用`start`方法来启动它。

```go
tw.start()
```

## 添加定时任务

要添加一个定时任务，你可以使用`AfterFunc`方法。这个方法接受两个参数：`d`和`f`。`d`参数表示定时任务的延迟时间，`f`参数是一个函数，它将在延迟时间过后被执行。

```go
tw.AfterFunc(1*time.Second, func() {
    fmt.Println("Hello, world!")
})
```

这里，我们添加了一个定时任务，它将在1秒后打印出"Hello, world!"。

## 停止TimingWheel

当你不再需要`TimingWheel`时，你可以调用`stop`方法来停止它。

```go
tw.stop()
```

注意，`stop`方法会阻塞直到所有的定时任务都已经被处理完毕。

## 使用ScheduleFunc添加定时任务

除了`AfterFunc`方法之外，你还可以使用`ScheduleFunc`方法来添加定时任务。这个方法接受两个参数：`s`和`f`。`s`参数是一个实现了`Scheduler`接口的对象，`f`参数是一个函数，它将在定时任务的执行时间到来时被执行。

```go
scheduler := &timingwheel.ScheduleFunc{
    NextFunc: func(time.Time) time.Time {
        return time.Now().Add(1 * time.Second)
    },
}

tw.ScheduleFunc(scheduler, func() {
    fmt.Println("Hello, world!")
})
```

这里，我们添加了一个定时任务，它将在1秒后打印出"Hello, world!"。`NextFunc`函数用于计算下一次定时任务的执行时间。

# 工作流程

1. `NewTimingWheel(tick time.Duration, wheelSize int64) *TimingWheel`：

这个函数用于创建一个新的TimingWheel实例。它接收两个参数，一个是tick，表示时间轮的最小时间单位，另一个是wheelSize，表示时间轮的大小。在函数内部，首先将tick转换为毫秒，然后获取当前时间的毫秒数。然后，创建一个bucket数组，数组的大小就是wheelSize。最后，创建一个TimingWheel实例，设置其tick、wheelSize、currentTime（当前时间，取自startMs并且被tick整除）、interval（时间轮的总时间跨度，等于tick乘以wheelSize）、buckets（刚刚创建的bucket数组）和queue（一个新的delayqueue实例）。

2. `start()`：

这个方法用于启动TimingWheel。它首先调用waitGroup的Wrap方法，将queue的Poll方法作为参数传入。Poll方法会持续从delayqueue中获取到期的bucket，并将其发送到queue的C通道。然后，再次调用waitGroup的Wrap方法，创建一个新的goroutine，这个goroutine会持续从queue的C通道中读取到期的bucket，然后调用advanceClock方法推进时间轮的时间，并调用bucket的Flush方法处理到期的bucket。

3. `add(t *Timer) bool`：

这个方法用于将一个定时器添加到TimingWheel。首先，获取当前时间，然后判断定时器的过期时间是否在当前时间轮的范围内。如果在当前时间轮的范围内，就将定时器添加到对应的bucket中，并将bucket添加到delayqueue中（如果bucket的过期时间发生了变化）。如果不在当前时间轮的范围内，就将定时器添加到上层的时间轮中（如果上层的时间轮还不存在，就创建一个新的上层时间轮）。

4. `addOrRun(t *Timer)`：

这个方法用于将一个定时器添加到TimingWheel，如果定时器已经过期，就直接运行定时器的任务。它首先尝试调用add方法将定时器添加到TimingWheel，如果添加失败（说明定时器已经过期），就在一个新的goroutine中运行定时器的任务。

5. `advanceClock(expiration int64)`：

这个方法用于推进时间轮的时间。它接收一个参数，表示新的过期时间。如果新的过期时间大于当前时间加上一个tick，就将当前时间设置为新的过期时间（被tick整除）。然后，尝试推进上层时间轮的时间（如果上层时间轮存在）。

6. `stop()`：

这个方法用于停止TimingWheel。它首先关闭exitC通道，然后等待所有的goroutine结束。

7. `AfterFunc(d time.Duration, f func()) *Timer` 和 `ScheduleFunc(s Scheduler, f func()) (t *Timer)`：

这两个方法用于创建一个定时任务，该任务在指定的时间后执行（AfterFunc）或者按照指定的调度计划执行（ScheduleFunc）。它们都会创建一个新的Timer实例，并调用addOrRun方法将其添加到TimingWheel。