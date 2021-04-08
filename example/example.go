package main

import (
	"github.com/lack-io/gscheduler"
)

func main() {
	gscheduler.Start()
	//
	//a := 1
	//
	//job1, _ := gscheduler.JobBuilder().Name("job1").Duration(time.Second).Fn(func() {
	//	fmt.Printf("[%s] a = %d\n", time.Now(), a)
	//	a++
	//}).Out()
	//
	//gscheduler.AddJob(job1)
	//
	//time.Sleep(time.Second * 4)
}
