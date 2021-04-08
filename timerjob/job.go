package timerjob

import (
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/go-redis/redis"
)

type TimerJob interface {
	Run()
	Stop()
}

type RedisLockerJob struct {
	BaseTimerJob

	Redis      *redis.Client
	LockID     string
	LockExpire time.Duration
}

const (
	unlockScript = "if redis.call('get', KEYS[1]) == ARGV[1] then return redis.call('del', KEYS[1]) else return 0 end"
)

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
				// defer cronRecovery()
				num := rand.Int()
				j.lock(num)
				j.Worker()
				j.unlock(num)
				wg.Done()
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

func (j *RedisLockerJob) Stop() {
	j.BaseTimerJob.Stop()
}

type BaseTimerJob struct {
	Interval time.Duration
	Worker   func()
	stopCh   chan struct{}
	ticker   *time.Ticker
}

func (j *BaseTimerJob) Run() {
	j.stopCh = make(chan struct{})
	j.ticker = time.NewTicker(j.Interval)

	wg := sync.WaitGroup{}
	for {
		select {
		case <-j.ticker.C:
			wg.Add(1)
			go func() {
				// defer cronRecovery()
				j.Worker()
				wg.Done()
			}()
		case <-j.stopCh:
			wg.Wait()
			return
		}
	}
}

func (j *BaseTimerJob) Stop() {
	j.ticker.Stop()
	close(j.stopCh)
}
