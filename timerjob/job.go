package timerjob

import (
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/lostyear/go-toolkits/recovery"
)

// TimerJob should run background
type TimerJob interface {
	Run()
	Stop()
}

// RedisLockerJob use redis lock to ensure just one job running
type RedisLockerJob struct {
	BaseTimerJob

	Redis      *redis.Client
	LockID     string
	LockExpire time.Duration
}

const (
	unlockScript = "if redis.call('get', KEYS[1]) == ARGV[1] then return redis.call('del', KEYS[1]) else return 0 end"
)

// Run Start the job running
func (j *RedisLockerJob) Run() {
	j.Worker()
	j.stopCh = make(chan struct{})
	j.ticker = time.NewTicker(j.Interval)

	wg := sync.WaitGroup{}
	for {
		select {
		case <-j.ticker.C:
			wg.Add(1)
			go func() {
				defer recovery.Recovery()
				defer wg.Done()
				num := rand.Int()
				j.lock(num)
				defer j.unlock(num)
				j.Worker()
			}()
		case <-j.stopCh:
			wg.Wait()
			return
		}
	}
}

func (j *RedisLockerJob) lock(uniqueValue int) {
	if ok, err := j.Redis.SetNX(j.LockID, uniqueValue, j.LockExpire).Result(); err != nil || !ok {
		log.Printf("get redis lock %s failed, sleep for next loop...\n", j.LockID)
		return
	}
}

func (j *RedisLockerJob) unlock(uniqueValue int) {
	j.Redis.Eval(unlockScript, []string{j.LockID}, uniqueValue)
}

// Stop the job running
func (j *RedisLockerJob) Stop() {
	j.BaseTimerJob.Stop()
}

// BaseTimerJob is a simple job
type BaseTimerJob struct {
	Interval time.Duration
	Worker   func()
	stopCh   chan struct{}
	ticker   *time.Ticker
}

// Run Start the job running
func (j *BaseTimerJob) Run() {
	j.Worker()
	j.stopCh = make(chan struct{})
	j.ticker = time.NewTicker(j.Interval)

	wg := sync.WaitGroup{}
	for {
		select {
		case <-j.ticker.C:
			wg.Add(1)
			go func() {
				defer recovery.Recovery()
				defer wg.Done()
				j.Worker()
			}()
		case <-j.stopCh:
			wg.Wait()
			return
		}
	}
}

// Stop the job running
func (j *BaseTimerJob) Stop() {
	j.ticker.Stop()
	close(j.stopCh)
}
