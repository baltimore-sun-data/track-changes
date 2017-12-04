package main

import (
	"container/heap"
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"
)

type job struct {
	owner                  *apiResponse
	data                   *pageInfo
	url, selector, twitter string
}

func (j job) Update() {
	log.Printf("Updating %v", j)

	if j.url != "" {
		j.owner.updateWeb(j.data, j.url, j.selector)
	}

	if j.twitter != "" {
		j.owner.updateTwitter(j.data, j.twitter)
	}
}

func (j job) String() string {
	s := ""
	if j.url != "" {
		s = fmt.Sprintf("{%s||%s}", j.url, j.selector)
	}
	if j.twitter != "" {
		s += fmt.Sprintf("{@%s}", j.twitter)
	}
	return s
}

type jobQueue struct {
	head, length int
	jobs         []job
}

func NewJobQueue(capacity int) *jobQueue {
	return &jobQueue{0, 0, make([]job, capacity)}
}

func (jq *jobQueue) push(nj job) {
	if len(jq.jobs) <= jq.length {
		jq.jobs = append(jq.jobs[jq.head:], append(jq.jobs[:jq.head:jq.head], nj)...)
		jq.jobs = jq.jobs[:cap(jq.jobs)]
		jq.head = 0
		jq.length++
	} else {
		jq.jobs[(jq.head+jq.length)%len(jq.jobs)] = nj
		jq.length++
	}
}

func (jq *jobQueue) shift() bool {
	if jq.length < 1 {
		return false
	}
	jq.head = (jq.head + 1) % len(jq.jobs)
	jq.length--
	return true
}

func (jq jobQueue) first() job {
	if jq.length < 1 {
		return job{}
	}
	return jq.jobs[jq.head]
}

func (jq *jobQueue) start(ctx context.Context) {
	var (
		workCh   = make(chan job)
		resultCh = make(chan job)
		tq       = make(timeQueue, 0, jq.length)
	)

	for i := 0; i < nWorkers; i++ {
		go worker(ctx, workCh, resultCh)
	}

	for {
		workers := workCh
		if jq.length < 1 {
			workers = nil
		}

		select {
		case workers <- jq.first():
			jq.shift()
		case j := <-resultCh:
			tq.add(j)
		case <-tq.timer():
			jq.push(tq.popJob())
		case <-ctx.Done():
			close(workCh)
			return
		}
	}

}

func worker(ctx context.Context, workCh, resultCh chan job) {
	for j := range workCh {
		j.Update()
		select {
		case resultCh <- j:
		case <-ctx.Done():
			return
		}
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

// popJob removes and returns the next job in the timequeue
func (tq *timeQueue) popJob() job {
	r := heap.Pop(tq).(timedJob)
	return r.job
}
