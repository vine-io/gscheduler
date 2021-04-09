package gscheduler

import "sync"

var (
	defaultScheduler Scheduler
	oncedo           sync.Once
)

func GetJob(id string) (*Job, error) {
	return defaultScheduler.GetJob(id)
}

func GetJobs() ([]*Job, error) {
	return defaultScheduler.GetJobs()
}

func Count() uint64 {
	return defaultScheduler.Count()
}

func AddJob(job *Job) error {
	return defaultScheduler.AddJob(job)
}

func UpdateJob(job *Job) error {
	return defaultScheduler.UpdateJob(job)
}

func RemoveJob(job *Job) error {
	return defaultScheduler.RemoveJob(job)
}

func Start() {
	oncedo.Do(func() {
		defaultScheduler = NewScheduler()
	})
	defaultScheduler.Start()
}

func Stop() {
	defaultScheduler.Stop()
}

func StopGraceful() {
	defaultScheduler.StopGraceful()
}
func Reset() {
	defaultScheduler.Reset()
}
