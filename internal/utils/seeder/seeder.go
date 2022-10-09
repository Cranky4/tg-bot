package seeder

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	faker "github.com/go-faker/faker/v4"
	"github.com/pkg/errors"
	iternalexpenses "gitlab.ozon.dev/cranky4/tg-bot/internal/model/expenses"
)

const (
	expenseCategoryInsertSQL = "INSERT INTO expense_categories(id, name) VALUES %s"
	expensesInsertSQL        = "INSERT INTO expenses(id, amount, datetime, category_id) VALUES %s"
)

type seeder struct {
	dsn string
	db  *sql.DB
}

func NewSeeder(dsn string) *seeder {
	return &seeder{
		dsn: dsn,
	}
}

func (s *seeder) ensureDBConnected() error {
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

func (s *seeder) SeedExpenses(ctx context.Context, expensesCount, categoriesCount int) error {
	var err error

	if err = s.ensureDBConnected(); err != nil {
		return errors.Wrap(err, "cannot connect to database")
	}

	categories := make([]iternalexpenses.ExpenseCategory, 0, categoriesCount)
	for i := 0; i < categoriesCount; i++ {
		categories = append(categories, iternalexpenses.ExpenseCategory{
			ID:   faker.UUIDHyphenated(),
			Name: faker.Word(),
		})
	}

	expenses := make([]iternalexpenses.Expense, 0, expensesCount)
	for i := 0; i < expensesCount; i++ {
		minDatetime := time.Now().AddDate(-1, 0, 0).Unix()

		expenses = append(expenses, iternalexpenses.Expense{
			ID:         faker.UUIDHyphenated(),
			CategoryID: categories[rand.Intn(categoriesCount)].ID,          //nolint:gosec
			Amount:     rand.Intn(99999) + 1,                               //nolint:gosec
			Datetime:   time.Unix(rand.Int63n(3600*24*365)+minDatetime, 0), //nolint:gosec
		})
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "cannot start transaction")
	}

	if err := insertCategories(ctx, categoriesCount, categories, tx); err != nil {
		tx.Rollback()
		return errors.Wrap(err, "cannot insert categories")
	}

	if err := insertExpenses(ctx, expensesCount, expenses, tx); err != nil {
		tx.Rollback()
		return errors.Wrap(err, "cannot insert expenses")
	}

	return tx.Commit()
}

func insertCategories(ctx context.Context, count int, categories []iternalexpenses.ExpenseCategory, tx *sql.Tx) error {
	valuesPlaceholders := make([]string, 0, count)
	paramsCount := 2
	values := make([]interface{}, 0, count*paramsCount)

	for i, c := range categories {
		valuesPlaceholders = append(valuesPlaceholders, fmt.Sprintf("($%d,$%d)", i*paramsCount+1, i*paramsCount+2))
		values = append(values, c.ID, c.Name)
	}

	return doBatchInsert(ctx, tx, expenseCategoryInsertSQL, valuesPlaceholders, values)
}

func insertExpenses(ctx context.Context, count int, categories []iternalexpenses.Expense, tx *sql.Tx) error {
	valuesPlaceholders := make([]string, 0, count)
	paramsCount := 4
	values := make([]interface{}, 0, count*paramsCount)

	for i, c := range categories {
		valuesPlaceholders = append(
			valuesPlaceholders,
			fmt.Sprintf("($%d,$%d,$%d,$%d)", i*paramsCount+1, i*paramsCount+2, i*paramsCount+3, i*paramsCount+4),
		)
		values = append(values, c.ID, c.Amount, c.Datetime, c.CategoryID)
	}

	return doBatchInsert(ctx, tx, expensesInsertSQL, valuesPlaceholders, values)
}

func doBatchInsert(ctx context.Context, tx *sql.Tx, sql string, placeholders []string, values []interface{}) error {
	query := fmt.Sprintf(sql, strings.Join(placeholders, ","))

	log.Print(query)

	_, err := tx.ExecContext(ctx, query, values...)

	return err
}
