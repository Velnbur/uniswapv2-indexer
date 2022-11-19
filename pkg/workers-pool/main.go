package workerspool

import (
	"context"
	"errors"
)

type Task func(ctx context.Context) error

type WorkingPool struct {
	TaskQueue   chan Task
	WorkerNum   int
	TasksAmount int64
}

var RetryError = errors.New("retry")

func NewWorkingPool(workerNum int, tasksAmount int64) *WorkingPool {
	return &WorkingPool{
		TaskQueue:   make(chan Task, tasksAmount),
		WorkerNum:   workerNum,
		TasksAmount: tasksAmount,
	}
}

type errQueue struct {
	err  error
	task Task
}

func (wp *WorkingPool) Run(ctx context.Context) error {
	ctxInner, cancel := context.WithCancel(ctx)
	defer cancel()

	errs := make(chan *errQueue, wp.TasksAmount)

	for i := 0; i < wp.WorkerNum; i++ {
		go wp.startWorker(ctxInner, errs)
	}

	return wp.waitWorkers(ctxInner, errs)
}

func (wp *WorkingPool) startWorker(ctx context.Context, errs chan<- *errQueue) {
	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-wp.TaskQueue:
			if !ok {
				break
			}
			if err := task(ctx); err != nil {
				errs <- &errQueue{err, task}
			}
			errs <- nil
		}
	}
}

func (wp *WorkingPool) waitWorkers(ctx context.Context, errs <-chan *errQueue) error {
	counter := int64(0)
	for {
		select {
		case <-ctx.Done():
			return errors.New("context canceled")
		case errQ := <-errs:
			if errQ != nil {
				if errors.Is(errQ.err, RetryError) {
					wp.TaskQueue <- errQ.task
					continue
				}
				return errQ.err
			}
			counter++
			if counter == wp.TasksAmount {
				close(wp.TaskQueue)
				return nil
			}
		}
	}
}

func (wp *WorkingPool) AddTask(task Task) {
	wp.TaskQueue <- task
}
