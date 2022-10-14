package worker

import (
	"container/list"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type DBWorker struct {
	db *gorm.DB
}

var (
	Queue = list.New()
)

func Initialize(db *gorm.DB) DBWorker {
	worker := DBWorker{db: db}

	go func() {
		for {
			time.Sleep(time.Second * 10)

			if Queue.Len() == 0 {
				continue
			}

			transaction := worker.db.Begin()

			for e := Queue.Front(); e != nil; e = e.Next() {
				f := e.Value.(func(*gorm.DB))

				f(transaction)

				Queue.Remove(e)
			}

			err := transaction.Commit().Error

			if err != nil {
				log.Error().Err(err).Msg("Found an issue with the worker commit.")
			}
		}
	}()

	return worker
}

func Do(f func(*gorm.DB)) {
	Queue.PushBack(f)
}
