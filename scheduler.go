package main

import (
	"container/heap"
	"math/rand"
	"time"
)

type job struct {
	data                   *jsonData
	url, selector, twitter string
}

func scheduler(jobs []job) {
	var (
		workCh   = make(chan job)
		resultCh = make(chan job)
		tq       = make(timeQueue, 0, len(jobs))
	)

	for i := 0; i < nWorkers; i++ {
		go worker(workCh, resultCh)
	}

	for {
		var (
			workers chan job
			j       job
		)

		if len(jobs) > 0 {
			j = jobs[0]
			workers = workCh
		}

		select {
		case workers <- j:
			jobs = jobs[1:]
		case j := <-resultCh:
			tq.add(j)
		case <-tq.timer():
			jobs = append(jobs, tq.job())
		}
	}

}

func worker(workCh, resultCh chan job) {
	for j := range workCh {
		data.Update(j)
		resultCh <- j
	}
}

type timedJob struct {
	job
	next  time.Time
	timer <-chan time.Time
}

type timeQueue []timedJob

func (tq timeQueue) Len() int { return len(tq) }

func (tq timeQueue) Less(i, j int) bool {
	return tq[j].next.After(tq[i].next)
}

func (tq timeQueue) Swap(i, j int) { tq[i], tq[j] = tq[j], tq[i] }

func (tq *timeQueue) Push(x interface{}) {
	*tq = append(*tq, x.(timedJob))
}

func (tq *timeQueue) Pop() interface{} {
	old := *tq
	n := len(old) - 1
	result := old[n]
	*tq = old[0:n]
	return result
}

// Returns the next timer to expire, if any
func (tq timeQueue) timer() <-chan time.Time {
	if len(tq) > 0 {
		return tq[0].timer
	}

	return nil
}

func (tq *timeQueue) add(j job) {
	sleep := dSleep - dSleep/2 + time.Duration(rand.Intn(int(dSleep)))
	queueItem := timedJob{j, time.Now().Add(sleep), time.After(sleep)}
	heap.Push(tq, queueItem)
}

// job pops the next job off the timequeue
func (tq *timeQueue) job() job {
	r := heap.Pop(tq).(timedJob)
	return r.job
}
