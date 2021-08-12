package main

import (
	"fmt"
	"log"
	"time"

	"github.com/vine-io/gscheduler"
)

func main() {
	gscheduler.Start()

	a := 1

	job1, _ := gscheduler.JobBuilder().Name("job1").Duration(time.Second).Fn(func() {
		fmt.Printf("[%s] a = %d\n", time.Now(), a)
		a++
	}).Out()

	gscheduler.AddJob(job1)

	time.Sleep(time.Second * 4)

	_, err := gscheduler.GetJob(job1.ID())
	if err != nil {
		log.Fatal(err)
	}
}
