package controlloop

import (
	"sync"
	"time"
)

type Resource struct {
	conditions    []Condition
	m             *sync.RWMutex
	killTimestamp string
	killMutex     *sync.Mutex
	finalizer     chan struct{}
}

type Condition struct {
	Type    string
	Status  string
	Reason  string
	Message string
}

func NewResource() Resource {
	return Resource{m: &sync.RWMutex{}, killMutex: &sync.Mutex{}, finalizer: make(chan struct{})}
}

func (r *Resource) Finalizer() <-chan struct{} {
	return r.finalizer
}
func (r *Resource) DoneFinalizer() {
	close(r.finalizer)
}

func (r *Resource) SetKillTimestamp(time time.Time) {
	r.killMutex.Lock()
	defer r.killMutex.Unlock()
	if r.killTimestamp == "" {
		r.killTimestamp = time.Format("2006-01-02 15:04:05")
	}
}

func (r *Resource) KillTimestamp() string {
	r.killMutex.Lock()
	defer r.killMutex.Unlock()
	return r.killTimestamp
}

func (r *Resource) setCondition(name, status, reason, message string) {
	r.m.Lock()
	defer r.m.Unlock()
	exist := func(slice []Condition, name string) bool {
		for _, item := range slice {
			if item.Type == name {
				return true
			}
		}
		return false
	}
	if exist(r.conditions, name) {
		for i := range r.conditions {
			if r.conditions[i].Type == name {
				r.conditions[i].Reason = reason
				r.conditions[i].Message = message
				r.conditions[i].Status = status
			}
		}
		return
	}
	r.conditions = append(r.conditions, Condition{Type: name, Status: status, Reason: reason, Message: message})
}

func (r *Resource) GetConditions() []Condition {
	r.m.RLock()
	defer r.m.RUnlock()
	return r.conditions
}

func (r *Resource) GetCondition(name string) (Condition, bool) {
	for _, c := range r.GetConditions() {
		if c.Type == name {
			return c, true
		}
	}
	return Condition{}, false
}

func (r *Resource) MarkFalse(name, reason, message string) {
	r.setCondition(name, "False", reason, message)
}

func (r *Resource) MarkTrue(name string) {
	r.setCondition(name, "True", "", "")
}
