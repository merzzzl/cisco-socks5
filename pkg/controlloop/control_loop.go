package controlloop

import (
	"context"
	"time"
)

type Result struct {
	RequeueAfter time.Duration
	Requeue      bool
}

type ResourceObject interface {
	GetConditions() []Condition
	GetCondition(name string) (Condition, bool)
	SetKillTimestamp(time time.Time)
	KillTimestamp() string
	SetDeletionTimestamp(time.Time)
	DeletionTimestamp() string
}

type Reconcile interface {
	Reconcile(context.Context, ResourceObject) (Result, error)
}
