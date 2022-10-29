package main

import (
	"log"

	"gitlab.ozon.dev/cranky4/tg-bot/internal/config"
	repo "gitlab.ozon.dev/cranky4/tg-bot/internal/repository"
	memoryrepo "gitlab.ozon.dev/cranky4/tg-bot/internal/repository/memory"
	sqlrepo "gitlab.ozon.dev/cranky4/tg-bot/internal/repository/sql"
)

func initRepo(conf config.Config) repo.ExpensesRepository {
	var repo repo.ExpensesRepository
	var err error

	switch conf.Storage.Mode {
	case "memory":
		repo = memoryrepo.NewRepository()
	case "sql":
		repo, err = sqlrepo.NewRepository(conf.Database)
		if err != nil {
			log.Fatalf("cannot connect to db %s", err.Error())
		}
	default:
		log.Fatalf("unknown repo mode %s", conf.Storage)
	}

	return repo
}
