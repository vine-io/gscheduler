# gscheduler

## 简介

gscheduler 是`golang`实现的一个简单的任务调度器。 实现功能：

- 任务的添加
- 任务的删除
- 任务的修改
- 任务的查询

## 实例

```golang
package main

import (
	"fmt"
	"time"

	"github.com/lack-io/gscheduler"
)

func main() {
	scheduler := gscheduler.NewScheduler()
	scheduler.Start()

	a := 1

	job1, _ := gscheduler.JobBuilder().Name("job1").Duration(time.Second).Fn(func() {
		fmt.Printf("[%s] a = %d\n", time.Now(), a)
		a++
	}).Out()

	job2, _ := gscheduler.JobBuilder().Name("job2").Duration(time.Second).Fn(func() {
		fmt.Println("job2", time.Now())
	}).Out()

	scheduler.AddJob(job1)
	scheduler.AddJob(job2)

	time.Sleep(time.Second * 10)
}
```

