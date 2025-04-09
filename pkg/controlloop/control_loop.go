package controlloop

import (
	"context"
	"time"
)

type Result struct {
	RequeueAfter time.Duration
}

type ResourceObject interface {
	GetConditions() []Condition
	GetCondition(name string) (Condition, bool)
	Finalizer() <-chan struct{}
	DoneFinalizer()
	SetKillTimestamp(time time.Time)
}

type Reconcile interface {
	Reconcile(context.Context, ResourceObject) (Result, error)
}
