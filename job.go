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
	"fmt"
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

type Job struct {
	id          uint64
	name        string
	cron        cron.Crontab // 任务时间间隔
	createTime  time.Time    // 任务创建时间
	lastTime    time.Time    // 任务上一次执行时间
	nextTime    time.Time
	isActive    bool   // 任务是否活动
	activeCount uint64 // 任务执行次数
	activeMax   uint64 // 任务允许执行的最大次数，0为无限次
	status      Status // 任务状态
	err         error  // 错误信息
	fn          func() // 任务函数
}

func (j *Job) Less(another rbtree.Item) bool {
	item, ok := another.(*Job)
	if !ok {
		return false
	}
	if !j.nextTime.Equal(item.nextTime) {
		return j.nextTime.Before(item.nextTime)
	}
	if j.id != item.id {
		return j.id < item.id
	}
	now := time.Now()
	if !j.cron.Next(now).Equal(item.cron.Next(now)) {
		return j.cron.Next(now).Before(item.cron.Next(now))
	}
	return j.name < item.name
}

func (j *Job) Start() {
	j.start(true)
}

func (j *Job) start(async bool) {
	j.activeCount++
	if async {
		go j.safeCall()
	} else {
		j.safeCall()
	}
}

func (j *Job) safeCall() {
	j.status = Running
	defer func() {
		j.status = Finished
		if err := recover(); err != nil {
			j.err = fmt.Errorf("[%s] %s: %v", time.Now().Format("2006-01-02 15:04:05"), j.name, err)
		}
	}()
	j.fn()
}
