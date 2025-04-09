package controlloop

import (
	"context"
	"fmt"
	"time"
)

const defaultReconcileTime = time.Second * 30
const errorReconcileTime = time.Second * 5

type ControlLoop struct {
	r           Reconcile
	object      ResourceObject
	stopChannel chan struct{}
	l           Logger
}

func New(r Reconcile, object ResourceObject, options ...ClOption) *ControlLoop {
	currentOptions := &opts{}
	for _, o := range options {
		o(currentOptions)
	}
	controlLoop := &ControlLoop{r: r, object: object, stopChannel: make(chan struct{})}

	if currentOptions.logger != nil {
		controlLoop.l = currentOptions.logger
	} else {
		controlLoop.l = &logger{}
	}
	return controlLoop
}

func (cl *ControlLoop) Run() <-chan struct{} {
	exitChannel := make(chan struct{})
	go func() {
		defer func() {
			exitChannel <- struct{}{}
		}()
		object := cl.object
		r := cl.r
		ctx := context.Background()
		ticker := time.NewTicker(defaultReconcileTime)
		result, err := cl.reconcile(ctx, r, object)
		switch {
		case err != nil:
			cl.l.Error(err)
			ticker.Reset(errorReconcileTime)
		case result.RequeueAfter > 0:
			ticker.Reset(result.RequeueAfter)
		default:
			ticker.Reset(defaultReconcileTime)
		}

		for {
			select {
			case <-object.Finalizer():
				return
			default:
				select {
				case <-cl.stopChannel:
					object.SetKillTimestamp(time.Now())
				case <-ticker.C:
				}
			}

			result, err := cl.reconcile(ctx, r, object)
			switch {
			case err != nil:
				cl.l.Error(err)
				ticker.Reset(errorReconcileTime)
			case result.RequeueAfter > 0:
				ticker.Reset(result.RequeueAfter)
			default:
				ticker.Reset(defaultReconcileTime)
			}

		}
	}()

	return exitChannel
}

func (cl *ControlLoop) Stop() {
	cl.stopChannel <- struct{}{}
}

func (cl *ControlLoop) reconcile(ctx context.Context, r Reconcile, object ResourceObject) (Result, error) {
	defer func() {
		if r := recover(); r != nil {
			cl.l.Error(fmt.Errorf("Recovered from panic: %v ", r))
		}
	}()
	return r.Reconcile(ctx, object)
}
