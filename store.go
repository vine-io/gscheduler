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
	"sync"

	"github.com/lack-io/gscheduler/rbtree"
)

type Store interface {
	GetJobs() []*Job
	GetByName(string) (*Job, bool)
	GetById(uint64) (*Job, bool)
	Count() uint
	Put(*Job)
	Del(*Job)
	Min() (*Job, bool)
}

type jobStore struct {
	sync.RWMutex
	store *rbtree.Rbtree
}

func newJobStore() *jobStore {
	return &jobStore{
		store: rbtree.New(),
	}
}

func (s *jobStore) GetJobs() []*Job {
	s.RLock()
	defer s.RUnlock()

	jobs := make([]*Job, 0)
	s.store.Ascend(s.store.Min(), func(item rbtree.Item) bool {
		i, ok := item.(*Job)
		if ok && i.name != staticJob {
			jobs = append(jobs, i)
		}

		return true
	})
	return jobs
}

func (s *jobStore) GetByName(name string) (*Job, bool) {
	s.RLock()
	defer s.RUnlock()

	var job *Job
	s.store.Ascend(s.store.Min(), func(item rbtree.Item) bool {
		i, ok := item.(*Job)
		if !ok || i.name == staticJob {
			return false
		}
		if i.name == name {
			job = i
			return true
		}
		return true
	})
	return job, job != nil
}

func (s *jobStore) GetById(id uint64) (*Job, bool) {
	s.RLock()
	defer s.RUnlock()

	var job *Job
	s.store.Ascend(s.store.Min(), func(item rbtree.Item) bool {
		i, ok := item.(*Job)
		if !ok || i.name == staticJob {
			return false
		}
		if i.id == id {
			job = i
			return true
		}
		return true
	})
	return job, job != nil
}

func (s *jobStore) Count() uint {
	s.RLock()
	defer s.RUnlock()
	return s.store.Len()
}

func (s *jobStore) Put(job *Job) {
	s.Lock()
	s.store.Insert(job)
	s.Unlock()
}

func (s *jobStore) Del(job *Job) {
	s.Lock()
	s.store.Delete(job)
	s.Unlock()
}

func (s *jobStore) Min() (*Job, bool) {
	s.RLock()
	item := s.store.Min()
	if item == nil {
		return nil, false
	}
	job, ok := item.(*Job)
	if !ok {
		s.RUnlock()
		return nil, false
	}
	s.RUnlock()

	return job, true
}
