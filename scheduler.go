// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gscheduler

import (
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lack-io/gscheduler/cron"
)

const _UNTOUCHED = time.Duration(math.MaxInt64)

var (
	defaultScheduler *Scheduler
	oncedo           sync.Once
	signal           = struct{}{}
	staticJob        = "__staticJob"
)

type Scheduler struct {
	seq         uint64
	store       Store
	count       uint64
	waitJobsNum uint64
	pauseChan   chan struct{}
	resumeChan  chan struct{}
	exitChan    chan struct{}
}

func Default() *Scheduler {
	oncedo.Do(initScheduler)
	return defaultScheduler
}

func initScheduler() {
	defaultScheduler = NewScheduler()
}

func NewScheduler() *Scheduler {
	s := &Scheduler{
		store:      newJobStore(),
		pauseChan:  make(chan struct{}),
		resumeChan: make(chan struct{}),
		exitChan:   make(chan struct{}),
	}
	return s
}

func (s *Scheduler) Start() {
	now := time.Now()
	untouchedJob := Job{
		createTime: now,
		cron:       cron.Every(time.Duration(math.MaxInt64)),
		fn: func() {
			//this jobItem is untouched.
		},
	}

	_, inserted := s.addJob(staticJob, untouchedJob.cron.Next(now), true, untouchedJob.cron, 1, untouchedJob.fn)
	if !inserted {
		panic("[scheduler] internal error.Reason cannot insert job.")
	}
	//开启守护协程
	go s.run()
	s.resume()
}

// 开启任务调度器
func (s *Scheduler) run() {
	var (
		timeout time.Duration
		job     *Job
		timer   = newSafeTimer(_UNTOUCHED)
	)
	defer timer.Stop()
Pause:
	<-s.resumeChan
	for {
		job, _ = s.store.Pop()

		timeout = job.nextTime.Sub(time.Now())
		timer.SafeReset(timeout)
		select {
		case <-timer.C:
			timer.SCR()
			atomic.AddUint64(&s.count, 1)

			if job.isActive {
				job.start(true)
				if job.activeMax == 0 || job.activeMax > job.activeCount {
					job.lastTime = job.nextTime
					job.nextTime = job.cron.Next(job.lastTime)
					s.store.Put(job)
				} else {
					s.removeJob(job)
				}
			} else {
				job.nextTime = job.cron.Next(job.lastTime)
				s.store.Put(job)
			}
		case <-s.pauseChan:
			goto Pause
		case <-s.exitChan:
			goto Exit
		}
	}
Exit:
}

func (s *Scheduler) pause() {
	s.pauseChan <- signal
}

func (s *Scheduler) resume() {
	s.resumeChan <- signal
}

func (s *Scheduler) exit() {
	s.exitChan <- signal
}

func (s *Scheduler) addJob(
	name string,
	nextTime time.Time,
	active bool,
	cron cron.Crontab,
	activeMax uint64,
	fn func(),
) (job *Job, inserted bool) {
	s.seq++
	s.waitJobsNum++
	job = &Job{
		id:         s.seq,
		name:       name,
		createTime: time.Now(),
		isActive:   active,
		lastTime:   time.Now(),
		nextTime:   nextTime,
		cron:       cron,
		activeMax:  activeMax,
		status:     Waiting,
		fn:         fn,
		store:      s.store,
	}
	s.store.Put(job)
	inserted = true
	return
}

func (s *Scheduler) removeJob(job *Job) (removed bool) {
	s.store.Del(job)
	s.waitJobsNum--

	// job.Cancel --> rmJob -->removeJob; schedule -->removeJob
	// it is call repeatly when Job.Cancel
	if atomic.CompareAndSwapInt32(&job.cancelFlag, 0, 1) {
		//job.innerCancel()
	}
	return
}

func (s *Scheduler) rmJob(job *Job) {
	s.pause()
	defer s.resume()

	s.removeJob(job)
	return
}

func (s *Scheduler) cleanJobs() {
	item, _ := s.store.Pop()
	for item != nil {
		item, _ = s.store.Pop()
	}
}

func (s *Scheduler) immediate() {
	for {
		if item, _ := s.store.Pop(); item != nil {
			atomic.AddUint64(&s.count, 1)
			item.start(false)
		} else {
			break
		}
	}
}

// 添加循环任务
// @name:    任务名
// @active:    任务是否活动
// @interval: 任务时间间隔
// @jobFunc: 任务函数，不能为nil
// return
// @Job:　　 任务信息
// @inserted: 任务是否添加成功
func (s *Scheduler) AddIntervalJob(
	name string,
	active bool,
	interval time.Duration,
	jobFunc func(),
) (job *Job, inserted bool) {
	if jobFunc == nil || interval.Nanoseconds() <= 0 {
		return
	}
	s.pause()
	job, inserted = s.addJob(
		name,
		time.Now().Add(interval),
		active,
		cron.Every(interval),
		0,
		jobFunc,
	)
	s.resume()
	return
}

// 添加一次性任务
// @name:    任务名
// @active:    任务是否活动
// @interval: 任务时间间隔
// @jobFunc: 任务函数，不能为nil
// return
// @Job:　　 任务信息
// @inserted: 任务是否添加成功
func (s *Scheduler) AddOnceJob(
	name string,
	active bool,
	interval time.Duration,
	jobFunc func(),
) (job *Job, inserted bool) {
	if jobFunc == nil || interval.Nanoseconds() <= 0 {
		return
	}
	s.pause()
	job, inserted = s.addJob(
		name,
		time.Now().Add(interval),
		active,
		cron.Every(interval),
		1,
		jobFunc,
	)
	s.resume()
	return
}

//
//// 添加多次执行任务
//func (s *Scheduler) AddTimesJob(
//	name string,
//	active bool,
//	interval time.Duration,
//	activeMax uint64,
//	jobFunc func(),
//) (job *Job, inserted bool) {
//	if jobFunc == nil || interval.Nanoseconds() <= 0 {
//		return
//	}
//	s.pause()
//	job, inserted = s.addJob(
//		name,
//		time.Now().Add(interval),
//		active,
//		cron.Every(interval),
//		activeMax,
//		jobFunc,
//	)
//	s.resume()
//	return
//}
//
//// 修改任务名
//func (s *Scheduler) UpdateJobName(jobmsg JobMsg, name string) (updated bool) {
//	if jobmsg == nil || name == "" {
//		return false
//	}
//	item, ok := jobmsg.(*Job)
//	if !ok {
//		return false
//	}
//	s.pause()
//	defer s.resume()
//
//	s.jobQueue.Delete(item)
//	item.name = name
//	s.jobQueue.Insert(item)
//	updated = true
//
//	return
//}
//
//// 修改任务运行状态
//func (s *Scheduler) UpdateJobActive(jobmsg JobMsg, active bool) (updated bool) {
//	if jobmsg == nil {
//		return false
//	}
//	item, ok := jobmsg.(*Job)
//	if !ok {
//		return false
//	}
//	s.pause()
//	defer s.resume()
//
//	s.jobQueue.Delete(item)
//	item.isActive = active
//	s.jobQueue.Insert(item)
//	updated = true
//	return
//}
//
//// 修改任务时间间隔
//func (s *Scheduler) UpdateJobInterval(jobmsg JobMsg, interval time.Duration) (updated bool) {
//	if jobmsg == nil || interval.Nanoseconds() <= 0 {
//		return false
//	}
//	item, ok := jobmsg.(*Job)
//	if !ok {
//		return false
//	}
//	s.pause()
//	defer s.resume()
//
//	s.jobQueue.Delete(item)
//	item.cron = cron.Every(interval)
//	item.nextTime = item.nextTime.Add(interval)
//	s.jobQueue.Insert(item)
//	updated = true
//
//	return
//}
//
//// 修改任务函数
//func (s *Scheduler) UpdateJobFunc(jobmsg JobMsg, fn func()) (updated bool) {
//	if jobmsg == nil {
//		return false
//	}
//	item, ok := jobmsg.(*Job)
//	if !ok {
//		return false
//	}
//	s.pause()
//	defer s.resume()
//
//	s.jobQueue.Delete(item)
//	item.fn = fn
//	s.jobQueue.Insert(item)
//	updated = true
//
//	return
//}

// 删除任务
func (s *Scheduler) RmJob(job *Job) (removed bool) {
	job.Cancel()
	removed = true
	return
}

func (s *Scheduler) GetJobs() []*Job {
	return s.store.GetJobs()
}

// Stop stop clock , and cancel all waiting jobs
func (s *Scheduler) Stop() {
	s.exit()

	s.cleanJobs()
}

// StopGracefull stop clock ,and do once every waiting job including Once\Reapeat
// Note:对于任务队列中，即使安排执行多次或者不限次数的，也仅仅执行一次。
func (s *Scheduler) StopGraceful() {
	s.exit()

	s.immediate()
}

// Count 已经执行的任务数。对于重复任务，会计算多次
func (s *Scheduler) Count() uint64 {
	return atomic.LoadUint64(&s.count)
}

// 重置Clock的内部状态
func (s *Scheduler) Reset() *Scheduler {
	s.exit()
	s.count = 0

	s.cleanJobs()
	s.Start()
	return s
}

// WaitJobs get how much jobs waiting for call
func (s *Scheduler) WaitJobs() uint64 {
	jobs := atomic.LoadUint64(&s.waitJobsNum) - 1
	return jobs
}
