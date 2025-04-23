package controlloop

import (
	"reflect"
	"sync"
	"warp-server/pkg/assertions"
)

type Storages struct {
	mu    sync.RWMutex
	store map[reflect.Type]interface{}
}

func NewStorages() *Storages {
	return &Storages{
		store: make(map[reflect.Type]interface{}),
	}
}

func GetStorage[T ResourceObject[T]](storages *Storages) (T, bool) {
	storages.mu.RLock()
	defer storages.mu.RUnlock()
	var zero T
	raw, ok := storages.store[assertions.TypeOf[T]()]
	if !ok {
		return zero, false
	}
	return assertions.As[T](raw)
}

func SetStorage[T ResourceObject[T]](storages *Storages, store Storage[T]) {
	storages.mu.Lock()
	defer storages.mu.Unlock()
	storages.store[assertions.TypeOf[T]()] = store
}

type Storage[T ResourceObject[T]] interface {
	Add(item T)
	Get(item T) T
	List(key ObjectKey) []T
	Update(item T) error
	Delete(item T)
	getLast() (T, bool, error)
}

func NewMemoryStorage[T ResourceObject[T]](q *Queue[T]) *MemoryStorage[T] {
	return &MemoryStorage[T]{
		m:       &sync.RWMutex{},
		Queue:   q,
		objects: make(map[ObjectKey]T),
	}
}

type MemoryStorage[T ResourceObject[T]] struct {
	m       *sync.RWMutex
	Queue   *Queue[T]
	objects map[ObjectKey]T
}

func (s *MemoryStorage[T]) Add(item T) {
	s.m.Lock()
	defer s.m.Unlock()
	s.objects[item.GetName()] = item
	s.Queue.add(item)
}

func (s *MemoryStorage[T]) Get(item T) T {
	s.m.RLock()
	defer s.m.RUnlock()
	var zero T
	val, exist := s.objects[item.GetName()]
	if !exist {
		return zero
	}
	return val.DeepCopy()
}

func (s *MemoryStorage[T]) List(key ObjectKey) []T {
	s.m.RLock()
	defer s.m.RUnlock()
	res := make([]T, len(s.objects))
	for _, v := range s.objects {
		res = append(res, v.DeepCopy())
	}
	return res
}

func (s *MemoryStorage[T]) Update(item T) error {
	s.m.Lock()
	defer s.m.Unlock()
	curr, exist := s.objects[item.GetName()]
	if !exist {
		return KeyNotExist
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

func (s *MemoryStorage[T]) Delete(item T) {
	s.m.Lock()
	defer s.m.Unlock()
	s.Queue.finalize(item)
	delete(s.objects, item.GetName())
}

func (s *MemoryStorage[T]) getLast() (T, bool, error) {
	s.m.Lock()
	defer s.m.Unlock()
	var zero T
	name, shutdown := s.Queue.get()
	if shutdown {
		return zero, true, nil
	}
	// object already deleted
	if _, exist := s.objects[name]; !exist {
		return zero, false, KeyNotExist
	}
	return s.objects[name].DeepCopy(), false, nil
}
