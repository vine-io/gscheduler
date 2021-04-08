package gscheduler

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/lack-io/gscheduler/cron"
	"github.com/lack-io/gscheduler/rbtree"
)

var (
	store Store
	job1  *Job
	job2  *Job
)

func init() {
	store = newJobStore()
	job1 = &Job{id: "1", cron: cron.Every(time.Second * 1)}
	job2 = &Job{id: "2", cron: cron.Every(time.Second * 2)}
	store.Put(job1)
	store.Put(job2)
}

func Test_jobStore_Min(t *testing.T) {
	tests := []struct {
		name  string
		store Store
		want  *Job
		want1 error
	}{
		{"TestMin-1", store, job1, nil},
		{"TestMin-2", store, job2, nil},
		{"TestMin-3", store, nil, fmt.Errorf("store empty")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.store
			got, got1 := s.Min()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Pop() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Pop() got1 = %v, want %v", got1, tt.want1)
			}
			if got != nil {
				s.Del(got)
			}
		})
	}
}

func Test_jobStore_GetJobs(t *testing.T) {
	tests := []struct {
		name  string
		store Store
		want  []*Job
	}{
		{"TestGetJobs-1", store, []*Job{job1, job2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.store
			if got, _ := s.GetJobs(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetJobs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_jobStore_GetByName(t *testing.T) {
	aJob := &Job{id: "3", name: "aa", cron: cron.Every(time.Second * 3)}
	bJob := &Job{id: "4", name: "bb", cron: cron.Every(time.Second * 4)}
	store.Put(aJob)
	store.Put(bJob)
	tests := []struct {
		name  string
		store Store
		args  string
		want  *Job
		want1 error
	}{
		{"TestGetByName-1", store, "aa", aJob, nil},
		{"TestGetByName-2", store, "bb", bJob, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.store
			got, got1 := s.GetByName(tt.args)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetByName() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetByName() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_jobStore_GetById(t *testing.T) {
	aJob := &Job{id: "3", name: "aa", cron: cron.Every(time.Second * 3)}
	bJob := &Job{id: "4", name: "bb", cron: cron.Every(time.Second * 4)}
	store.Put(aJob)
	store.Put(bJob)
	tests := []struct {
		name  string
		store Store
		args  string
		want  *Job
		want1 error
	}{
		{"TestGetById-1", store, "3", aJob, nil},
		{"TestGetById-2", store, "4", bJob, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.store
			got, got1 := s.GetById(tt.args)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetById() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetById() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_jobStore_Put(t *testing.T) {
	job3 := &Job{id: "4"}
	store.Put(job3)
	t.Log(store.Count())

	r := rbtree.New()
	r.Insert(job1)
	r.Insert(job2)
	r.Insert(job3)
	t.Log(r.Len())
}
