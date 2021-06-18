package timerjob

import (
	"sync"

	"github.com/lostyear/go-toolkits/recovery"
)

// List is a job list
type List struct {
	sync.WaitGroup
	Jobs []TimerJob
}

// Start all jobs in the list
func (l *List) Start() {
	for _, job := range l.Jobs {
		l.Add(1)
		go func(j TimerJob) {
			defer recovery.Recovery()
			defer l.Done()

			j.Run()
		}(job)
	}
}

// Stop run job and wait for all job done
func (l *List) Stop() {
	for _, job := range l.Jobs {
		job.Stop()
	}
	l.Wait()
}

// AddJob to the list
func (l *List) AddJob(job ...TimerJob) {
	l.Jobs = append(l.Jobs, job...)
}
