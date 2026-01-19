package core

import (
	"time"
	"sync"
	"skynet-net-engine-api/internal/database"
	"skynet-net-engine-api/internal/models"
	"skynet-net-engine-api/pkg/logger"
	"go.uber.org/zap"
)

type Pool struct {
	Workers map[int]*Worker
	Lock    sync.RWMutex
	Ready   sync.WaitGroup
}

var GlobalPool *Pool

func InitPool() {
	GlobalPool = &Pool{
		Workers: make(map[int]*Worker),
	}

	// 1. Fetch Routers
	routers, err := database.GetAllRouters()
	if err != nil {
		logger.Error("Failed to load routers for pool - Continuing with empty pool", zap.Error(err))
	}

	// 2. Spawn Workers
	for _, r := range routers {
		// EXPERIMENTAL FIX: Force Direct IP for Randuagung
		if r.ID == 1 {
			r.Host = "103.156.128.114"
			r.Port = 8728
		}

		GlobalPool.Ready.Add(1) // Expect readiness signal
		logger.Info("Spawning Worker", zap.Int("id", r.ID), zap.String("name", r.Name), zap.String("host", r.Host))
		worker := NewWorker(r, &GlobalPool.Ready)
		
		GlobalPool.Lock.Lock()
		GlobalPool.Workers[r.ID] = worker
		GlobalPool.Lock.Unlock()

		// Start the engine in a persistent Goroutine
		go worker.Start()
	}

	logger.Info("Worker Pool Initialized", zap.Int("workers", len(routers)))
}

func (p *Pool) WaitForReady() {
	logger.Info("Waiting for routers to warmup...")
	
	// Create a channel to signal completion
	done := make(chan struct{})
	go func() {
		p.Ready.Wait()
		close(done)
	}()

	// Wait with timeout
	select {
	case <-done:
		logger.Info("All routers ready!")
	case <-time.After(5 * time.Second):
		logger.Warn("Warmup timed out - Some routers may be effectively offline or slow")
	}
}

func (p *Pool) GetWorker(id int) *Worker {
	p.Lock.RLock()
	defer p.Lock.RUnlock()
	return p.Workers[id]
}

func (p *Pool) GetAllTargets() []models.ActiveUser {
	p.Lock.RLock()
	defer p.Lock.RUnlock()

	total := make([]models.ActiveUser, 0)
	for _, w := range p.Workers {
		w.Lock.RLock()
		// Copy users to avoid race conditions if underlying array changes
		if len(w.ActiveUsers) > 0 {
			batch := make([]models.ActiveUser, len(w.ActiveUsers))
			copy(batch, w.ActiveUsers)
			total = append(total, batch...)
		}
		w.Lock.RUnlock()
	}
	return total
}
