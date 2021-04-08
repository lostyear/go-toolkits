package timerjob

import (
	"sync"

	"github.com/lostyear/go-toolkits/recovery"
)

// var logger logr.Logger

// func init() {
// 	logger = commons.LoggerWithName("manager-logger")
// }

type TimerJobList struct {
	sync.WaitGroup
	Jobs []TimerJob
}

func (l *TimerJobList) Start() {
	for _, job := range l.Jobs {
		l.Add(1)
		go func(j TimerJob) {
			defer recovery.Recovery()
			defer l.Done()

			j.Run()
		}(job)
	}
}

func (l *TimerJobList) Stop() {
	for _, job := range l.Jobs {
		job.Stop()
	}
	l.Wait()
}

func (l *TimerJobList) AddJob(job ...TimerJob) {
	l.Jobs = append(l.Jobs, job...)
}
