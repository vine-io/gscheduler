# gscheduler

## 简介
gscheduler 是`golang`实现的一个简单的任务调度器。
实现功能：
- 任务的添加
- 任务的删除
- 任务的时间间隔修改
- 任务的名称修改
- 任务的执行函数
- 任务的信息json输出

## 实例
```golang
package main

import (
	"fmt"
	"time"
	"github.com/xingyys/gscheduler"
)

func main() {
	scheduler := gscheduler.NewScheduler()
	scheduler.Start()

	// 添加循环时间
	scheduler.AddIntervalJob("job1", true, time.Second * 1, func() {
		fmt.Println("runing job1")
	})

	// 添加一次性时间
	scheduler.AddOnceJob("job2", true, time.Second * 1, func() {
		fmt.Println("running once job")
	})

	// 添加执行次数任务
	v := 1
	scheduler.AddTimesJob("job3", true, time.Second * 2, 3, func() {
		fmt.Println("running job at time ", v)
		v++
	})
	time.Sleep(time.Second * 10)
	// 结束任务
	scheduler.Stop()
}
```

## 感谢
[https://github.com/alex023/clock][1]


  [1]: https://github.com/alex023/clock