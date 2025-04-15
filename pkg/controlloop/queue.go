package controlloop

import (
	"k8s.io/client-go/util/workqueue"
	"sync"
	"time"
)

type Queue[t comparable] struct {
	queue        workqueue.TypedRateLimitingInterface[ObjectKey]
	existedItems map[ObjectKey]ResourceObject
	m            *sync.RWMutex
}

func NewQueue() *Queue[ResourceObject] {
	rateLimitingConfig := workqueue.TypedRateLimitingQueueConfig[ObjectKey]{}
	rateLimitingConfig.DelayingQueue = workqueue.NewTypedDelayingQueue[ObjectKey]()
	queue := workqueue.NewTypedRateLimitingQueueWithConfig[ObjectKey](workqueue.NewTypedMaxOfRateLimiter[ObjectKey](), rateLimitingConfig)
	return &Queue[ResourceObject]{queue: queue, existedItems: make(map[ObjectKey]ResourceObject), m: &sync.RWMutex{}}
}

func (q *Queue[t]) getExistedItems() map[ObjectKey]ResourceObject {
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
	q.existedItems[item.GetName()] = item
	q.queue.Add(item.GetName())
}

func (q *Queue[t]) GetResource(item ResourceObject) ResourceObject {
	q.m.RLock()
	defer q.m.RUnlock()
	val, exist := q.existedItems[item.GetName()]
	if !exist {
		return nil
	}
	return val.DeepCopy()
}

func (q *Queue[t]) Update(item ResourceObject) error {
	q.m.Lock()
	defer q.m.Unlock()
	curr, exist := q.existedItems[item.GetName()]
	if !exist {
		return KetNotExist
	}
	if curr.GetGeneration() > item.GetGeneration() {
		return AlreadyUpdated
	}
	curr.IncGeneration()
	q.existedItems[item.GetName()] = item
	q.queue.Add(item.GetName())
	return nil
}

func (q *Queue[t]) add(item ResourceObject) {
	q.queue.Add(item.GetName())
}

func (q *Queue[t]) finalize(item ResourceObject) {
	q.m.Lock()
	defer q.m.Unlock()
	delete(q.existedItems, item.GetName())
}

func (q *Queue[t]) done(item ResourceObject) {
	q.queue.Done(item.GetName())
}

func (q *Queue[t]) addAfter(item ResourceObject, duration time.Duration) {
	q.queue.AddAfter(item.GetName(), duration)
}

func (q *Queue[t]) addRateLimited(item ResourceObject) {
	q.queue.AddRateLimited(item.GetName())
}

func (q *Queue[t]) get() (ResourceObject, bool) {
	q.m.Lock()
	defer q.m.Unlock()
	name, shutdown := q.queue.Get()
	if shutdown {
		return nil, true
	}
	return q.existedItems[name].DeepCopy(), false
}
