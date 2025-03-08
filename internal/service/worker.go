package service

import (
	"context"
	"sync"

	"github.com/shopspring/decimal"
	"github.com/sunriseex/test_wallet/internal/logger"
)

type Job struct {
	WalletID string
	Amount   decimal.Decimal
	Ctx      context.Context
}

type WorkerPool struct {
	jobs      chan Job
	wg        sync.WaitGroup
	svc       *WalletServiceImpl
	queueSize int
}

func NewWorkerPool(svc *WalletServiceImpl, workers, queueSize int) *WorkerPool {
	wp := &WorkerPool{
		jobs:      make(chan Job, queueSize),
		svc:       svc,
		queueSize: queueSize,
	}
	for i := 0; i < workers; i++ {
		wp.wg.Add(1)
		go wp.worker()

	}
	return wp

}

func (wp *WorkerPool) worker() {
	defer wp.wg.Done()
	for job := range wp.jobs {
		if err := wp.svc.updateBalance(job.Ctx, job.WalletID, job.Amount); err != nil {
			logger.Log.Errorf("Ошибка обновления баланса для WalletID=%s: %v", job.WalletID, err)
		}

	}
}

func (wp *WorkerPool) AddJob(job Job) bool {
	select {
	case wp.jobs <- job:
		return true
	default:
		return false
	}
}

func (wp *WorkerPool) Shutdown() {
	close(wp.jobs)
	wp.wg.Wait()
}
