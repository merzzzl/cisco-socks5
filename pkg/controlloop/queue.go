package controlloop

import (
	"k8s.io/client-go/util/workqueue"
	"sync"
	"time"
)

type Queue[t comparable] struct {
	queue        workqueue.TypedRateLimitingInterface[ResourceObject]
	existedItems map[ResourceObject]struct{}
	m            *sync.RWMutex
}

func NewQueue() *Queue[ResourceObject] {
	rateLimitingConfig := workqueue.TypedRateLimitingQueueConfig[ResourceObject]{}
	rateLimitingConfig.DelayingQueue = workqueue.NewTypedDelayingQueue[ResourceObject]()
	queue := workqueue.NewTypedRateLimitingQueueWithConfig[ResourceObject](workqueue.NewTypedMaxOfRateLimiter[ResourceObject](), rateLimitingConfig)
	return &Queue[ResourceObject]{queue: queue, existedItems: make(map[ResourceObject]struct{}), m: &sync.RWMutex{}}
}

func (q *Queue[t]) getExistedItems() map[ResourceObject]struct{} {
	q.m.RLock()
	defer q.m.RUnlock()
	return q.existedItems
}

func (q *Queue[t]) len() int {
	q.m.RLock()
	defer q.m.RUnlock()
	return len(q.existedItems)
}

func (q *Queue[t]) AddResource(item ResourceObject) {
	q.m.Lock()
	defer q.m.Unlock()
	q.existedItems[item] = struct{}{}
	q.queue.Add(item)
}

func (q *Queue[t]) add(item ResourceObject) {
	q.queue.Add(item)
}

func (q *Queue[t]) finalize(item ResourceObject) {
	q.m.Lock()
	defer q.m.Unlock()
	delete(q.existedItems, item)
}

func (q *Queue[t]) done(item ResourceObject) {
	q.queue.Done(item)
}

func (q *Queue[t]) addAfter(item ResourceObject, duration time.Duration) {
	q.queue.AddAfter(item, duration)
}

func (q *Queue[t]) addRateLimited(item ResourceObject) {
	q.queue.AddRateLimited(item)
}

func (q *Queue[t]) get() (ResourceObject, bool) {
	return q.queue.Get()
}
