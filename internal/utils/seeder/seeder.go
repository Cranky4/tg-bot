package seeder

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"strings"
	"time"

	faker "github.com/go-faker/faker/v4"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model"
)

const (
	expenseCategoryInsertSQL = "INSERT INTO expense_categories(id, name) VALUES %s"
	expensesInsertSQL        = "INSERT INTO expenses(id, amount, datetime, category_id, user_id) VALUES %s"

	cannotConnectToDB               = "не могу подключиться к базе данных"
	cannotStartTransactionErrMsg    = "не могу начать транзакцию"
	cannotRollbackTransactionErrMsg = "не могу откатить транзакцию"
	cannotInsertCategoriesErrMsg    = "не могу добавить категории"

	userId = 100
)

type Seeder interface {
	SeedExpenses(ctx context.Context, expensesCount, categoriesCount int) error
}

type dbSeeder struct {
	dsn string
	db  *sql.DB
}

func NewSeeder(dsn string) Seeder {
	return &dbSeeder{
		dsn: dsn,
	}
}

func (s *dbSeeder) ensureDBConnected() error {
	if s.db != nil {
		return nil
	}

	db, err := sql.Open("pgx", s.dsn)
	if err != nil {
		return err
	}
	s.db = db

	err = s.db.Ping()
	if err != nil {
		return err
	}

	return nil
}

func (s *dbSeeder) SeedExpenses(ctx context.Context, expensesCount, categoriesCount int) error {
	var err error

	if err = s.ensureDBConnected(); err != nil {
		return errors.Wrap(err, cannotConnectToDB)
	}

	categories := make([]model.ExpenseCategory, 0, categoriesCount)
	for i := 0; i < categoriesCount; i++ {
		categories = append(categories, model.ExpenseCategory{
			ID:   faker.UUIDHyphenated(),
			Name: faker.Word(),
		})
	}

	expenses := make([]model.Expense, 0, expensesCount)
	for i := 0; i < expensesCount; i++ {
		minDatetime := time.Now().AddDate(-1, 0, 0).Unix()

		expenses = append(expenses, model.Expense{
			ID:         faker.UUIDHyphenated(),
			CategoryID: categories[rand.Intn(categoriesCount)].ID,
			Amount:     rand.Int63n(99999) + 1,
			Datetime:   time.Unix(rand.Int63n(3600*24*365)+minDatetime, 0),
			UserId:     userId,
		})
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return errors.Wrap(err, cannotStartTransactionErrMsg)
	}

	defer func() {
		if err != nil {
			err = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	if err = insertCategories(ctx, categoriesCount, categories, tx); err != nil {
		if tErr := tx.Rollback(); tErr != nil {
			return errors.Wrap(tErr, cannotRollbackTransactionErrMsg)
		}

		return errors.Wrap(err, cannotInsertCategoriesErrMsg)
	}

	if err = insertExpenses(ctx, expensesCount, expenses, tx); err != nil {
		return errors.Wrap(err, cannotInsertCategoriesErrMsg)
	}

	return err
}

func insertCategories(ctx context.Context, count int, categories []model.ExpenseCategory, tx *sql.Tx) error {
	valuesPlaceholders := make([]string, 0, count)
	paramsCount := 2
	values := make([]interface{}, 0, count*paramsCount)

	for i, c := range categories {
		valuesPlaceholders = append(valuesPlaceholders, fmt.Sprintf("($%d,$%d)", i*paramsCount+1, i*paramsCount+2))
		values = append(values, c.ID, c.Name)
	}

	return doBatchInsert(ctx, tx, expenseCategoryInsertSQL, valuesPlaceholders, values)
}

func insertExpenses(ctx context.Context, count int, expenses []model.Expense, tx *sql.Tx) error {
	valuesPlaceholders := make([]string, 0, count)
	paramsCount := 5
	values := make([]interface{}, 0, count*paramsCount)

	for i, e := range expenses {
		valuesPlaceholders = append(
			valuesPlaceholders,
			fmt.Sprintf("($%d,$%d,$%d,$%d,$%d)", i*paramsCount+1, i*paramsCount+2, i*paramsCount+3, i*paramsCount+4, i*paramsCount+5),
		)
		values = append(values, e.ID, e.Amount, e.Datetime, e.CategoryID, e.UserId)
	}

	return doBatchInsert(ctx, tx, expensesInsertSQL, valuesPlaceholders, values)
}

func doBatchInsert(ctx context.Context, tx *sql.Tx, sql string, placeholders []string, values []interface{}) error {
	query := fmt.Sprintf(sql, strings.Join(placeholders, ","))

	_, err := tx.ExecContext(ctx, query, values...)

	return err
}
