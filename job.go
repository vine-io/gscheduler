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
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/lack-io/gscheduler/cron"
	"github.com/lack-io/gscheduler/rbtree"
)

type Status string

const (
	Waiting  Status = "waiting"
	Running  Status = "running"
	Finished Status = "finished"
)

type JobMsg interface {
	C() <-chan JobMsg
}

type Job struct {
	id          uint64
	name        string
	cron        cron.Crontab // 任务时间间隔
	createTime  time.Time    // 任务创建时间
	lastTime    time.Time    // 任务上一次执行时间
	nextTime    time.Time    // 任务下一次执行时间
	isActive    bool         // 任务是否活动
	activeCount uint64       // 任务执行次数
	activeMax   uint64       // 任务允许执行的最大次数，0为无限次
	status      Status       // 任务状态
	errorMsg    string       // 错误信息
	fn          func()       // 任务函数
	msgChan     chan JobMsg  // 任务输出信息
	cancelFlag  int32
	scheduler   *Scheduler
}

func (j *Job) Less(another rbtree.Item) bool {
	item, ok := another.(*Job)
	if !ok {
		return false
	}
	if !j.nextTime.Equal(item.nextTime) {
		return j.nextTime.Before(item.nextTime)
	}
	return j.id < item.id
}

func (j *Job) start(async bool) {
	j.activeCount++
	if async {
		go j.safeCall()
	} else {
		j.safeCall()
	}
}

// 删除任务
func (j *Job) Cancel() {
	if atomic.CompareAndSwapInt32(&j.cancelFlag, 0, 1) {
		j.scheduler.rmJob(j)
		j.innerCancel()
	}
}

type Data struct {
	Id   uint64
	Name string
}

type jobDetails struct {
	Id          uint64    `json:"id"`
	Name        string    `json:"name"`
	CreateTime  time.Time `json:"createTime"`
	LastTime    time.Time `json:"lastTime"`
	NextTime    time.Time `json:"nextTime"`
	IsActive    bool      `json:"isActive"`
	ActiveCount uint64    `json:"activeCount"`
	ActiveMax   uint64    `json:"activeMax"`
	Status      Status    `json:"status"`
	ErrorMsg    string    `json:"errorMsg"`
}

func (j Job) Details() []byte {
	jobdetails := &jobDetails{}
	jobdetails.Id = j.id
	jobdetails.Name = j.name
	jobdetails.CreateTime = j.createTime
	jobdetails.LastTime = j.lastTime
	jobdetails.NextTime = j.nextTime
	jobdetails.IsActive = j.isActive
	jobdetails.ActiveCount = j.activeCount
	jobdetails.ActiveMax = j.activeMax
	jobdetails.Status = j.status
	jobdetails.ErrorMsg = j.errorMsg
	data, _ := json.Marshal(jobdetails)
	return data
}

func (j *Job) C() <-chan JobMsg {
	return j.msgChan
}

func (j *Job) innerCancel() {
	j.scheduler = nil
	close(j.msgChan)
}

func (j *Job) safeCall() {
	j.status = Running
	defer func() {
		j.status = Finished
		if err := recover(); err != nil {
			j.errorMsg = fmt.Sprintf("[%v] error: %v at %v", j.name, err, time.Now().Format("2006-01-02 15:04:05"))
		}
	}()
	j.fn()
}
