package controlloop

import (
	"sync"
)

type Storage interface {
	Add(item ResourceObject)
	Get(item ResourceObject) ResourceObject
	Update(item ResourceObject) error
	Delete(item ResourceObject)
	getLast() (ResourceObject, bool, error)
}

func NewMemoryStorage(q *Queue[ResourceObject]) *MemoryStorage {
	return &MemoryStorage{
		m:       &sync.RWMutex{},
		Queue:   q,
		objects: make(map[ObjectKey]ResourceObject),
	}
}

type MemoryStorage struct {
	m       *sync.RWMutex
	Queue   *Queue[ResourceObject]
	objects map[ObjectKey]ResourceObject
}

func (s *MemoryStorage) Add(item ResourceObject) {
	s.m.Lock()
	defer s.m.Unlock()
	s.objects[item.GetName()] = item
	s.Queue.add(item)
}

func (s *MemoryStorage) Get(item ResourceObject) ResourceObject {
	s.m.RLock()
	defer s.m.RUnlock()
	val, exist := s.objects[item.GetName()]
	if !exist {
		return nil
	}
	return val.DeepCopy()
}

func (s *MemoryStorage) Update(item ResourceObject) error {
	s.m.Lock()
	defer s.m.Unlock()
	curr, exist := s.objects[item.GetName()]
	if !exist {
		return KetNotExist
	}
	if curr.GetGeneration() > item.GetGeneration() {
		return AlreadyUpdated
	}
	curr.IncGeneration()
	s.objects[item.GetName()] = item
	s.Queue.add(item)
	s.Queue.done(item)
	return nil
}

func (s *MemoryStorage) Delete(item ResourceObject) {
	s.m.Lock()
	defer s.m.Unlock()
	s.Queue.finalize(item)
	delete(s.objects, item.GetName())
}

func (s *MemoryStorage) getLast() (ResourceObject, bool, error) {
	s.m.Lock()
	defer s.m.Unlock()
	name, shutdown := s.Queue.get()
	if shutdown {
		return nil, true, nil
	}
	// object already deleted
	if _, exist := s.objects[name]; !exist {
		return nil, false, KetNotExist
	}
	return s.objects[name].DeepCopy(), false, nil
}
