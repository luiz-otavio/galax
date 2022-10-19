package worker

import (
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type Worker interface {
	Do(f func(*gorm.DB))

	Shutdown()
	Initialize()

	IsShutdown() bool
}

type dbWorkerImpl struct {
	db *gorm.DB

	// Thread
	ch chan func(*gorm.DB)

	// Lock

	shutdown bool
}

func (worker *dbWorkerImpl) Do(f func(*gorm.DB)) {
	worker.ch <- f
}

func (worker *dbWorkerImpl) Shutdown() {
	if worker.shutdown {
		log.Error().Msg("Worker already shutdown")
		return
	}

	if len(worker.ch) > 0 {
		log.Warn().Msg("Worker shutdown with pending tasks")

		// Wait for all tasks to be done
		for len(worker.ch) > 0 {
			time.Sleep(1 * time.Second)
		}
	}

	close(worker.ch)

	worker.shutdown = true
}

func (worker *dbWorkerImpl) IsShutdown() bool {
	return worker.shutdown
}

func (worker *dbWorkerImpl) Initialize() {
	if worker.shutdown {
		log.Error().Msg("Worker already shutdown")
		return
	}

	go func() {
		for {
			time.Sleep(time.Second * 10)

			if len(worker.ch) == 0 {
				continue
			}

			if worker.shutdown {
				log.Error().Msg("Worker already shutdown")
				break
			}

			transaction := worker.db.Begin()

			for i := 0; i < len(worker.ch); i++ {
				f := <-worker.ch

				f(transaction)
			}

			err := transaction.Commit().Error

			if err != nil {
				log.Error().Err(err).Msg("Found an issue with the worker commit.")
			}

		}
	}()
}

func CreateWorker(db *gorm.DB) Worker {
	return &dbWorkerImpl{
		db: db,

		ch: make(chan func(*gorm.DB), 127),

		shutdown: false,
	}
}
