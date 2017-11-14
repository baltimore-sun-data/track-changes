package main

import (
	"container/heap"
	"log"
	"math/rand"
	"time"
)

const (
	nWorkers = 10
	dSleep   = 5 * time.Minute
)

type job struct {
	id, url, selector string
}

type result struct {
	job
	next  time.Time
	timer <-chan time.Time
}

func scheduler(jobs []job) {
	var (
		workCh   = make(chan job)
		resultCh = make(chan result)
		tq       = make(timeQueue, 0, len(jobs))
	)

	for i := 0; i < nWorkers; i++ {
		go worker(workCh, resultCh)
	}

	for {
		var (
			workers chan job
			j       job
			timer   <-chan time.Time
		)

		if len(jobs) > 0 {
			j = jobs[len(jobs)-1]
			workers = workCh
		}

		if len(tq) > 0 {
			timer = tq[0].timer
		}

		select {
		case workers <- j:
			jobs = jobs[:len(jobs)-1]
		case result := <-resultCh:
			heap.Push(&tq, result)
		case <-timer:
			r := heap.Pop(&tq).(result)
			jobs = append(jobs, r.job)
		}
	}

}

func worker(workCh chan job, resultCh chan result) {
	for j := range workCh {
		log.Printf("Starting job %#v", j)
		txt, err := get(j.url, j.selector)
		if err != nil {
			log.Printf("Error for %s: %v", j.id, err)
		} else {
			data.Lock()
			data.m[j.id] = txt
			data.Unlock()
		}

		sleep := dSleep - dSleep/2 + time.Duration(rand.Intn(int(dSleep)))
		resultCh <- result{j, time.Now().Add(sleep), time.After(sleep)}
	}
}

type timeQueue []result

func (tq timeQueue) Len() int { return len(tq) }

func (tq timeQueue) Less(i, j int) bool {
	return tq[j].next.After(tq[i].next)
}

func (tq timeQueue) Swap(i, j int) { tq[i], tq[j] = tq[j], tq[i] }

func (tq *timeQueue) Push(x interface{}) {
	*tq = append(*tq, x.(result))
}

func (tq *timeQueue) Pop() interface{} {
	old := *tq
	n := len(old) - 1
	result := old[n]
	*tq = old[0:n]
	return result
}
