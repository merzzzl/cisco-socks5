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
	GetGeneration() int64
	IncGeneration()
	SetKillTimestamp(time time.Time)
	GetKillTimestamp() string
	SetDeletionTimestamp(time.Time)
	GetDeletionTimestamp() string
	GetName() ObjectKey
	DeepCopy() ResourceObject
}

type Reconcile interface {
	Reconcile(context.Context, ResourceObject) (Result, error)
}
