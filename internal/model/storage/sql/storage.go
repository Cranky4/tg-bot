package memorystorage

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/config"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model/expenses"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model/messages"
)

const (
	expenseCategorySearchSQL = "SELECT id, name FROM expense_categories WHERE name LIKE $1 LIMIT 1"
	expenseCategoryInsertSQL = "INSERT INTO expense_categories(id, name) VALUES ($1, $2)"
	expensesInsertSQL        = "INSERT INTO expenses(id, amount, datetime, category_id) VALUES ($1,$2,$3,$4)"
	expensesSelectSQL        = "SELECT e.id, e.amount, e.datetime, c.id as categoryId, c.name FROM expenses e " +
		"INNER JOIN expense_categories c ON e.category_id = c.id WHERE datetime > $1"
	expensesSelectCountSQL = "SELECT COUNT(id) FROM expenses WHERE datetime > $1"

	addExpenseErrMsg         = "ошибка в методе addExpense"
	findCategoryErrMsg       = "ошибка в методе findCategory"
	createNewCategoryErrMsg  = "ошибка в методе createNewCategory"
	createNewExpenseErrMsg   = "ошибка в методе createNewExpense"
	getExpensesErrMsg        = "ошибка в методе getExpenses"
	expenseSelectErrMsg      = "ошибка в методе findExpenses"
	expenseSelectCountErrMsg = "ошибка в методе findCountExpenses"
)

type storage struct {
	ctx context.Context
	dsn string
	db  *sql.DB

	categorySearchStmt     *sql.Stmt
	categoryInsertStmt     *sql.Stmt
	expenseInsertStmt      *sql.Stmt
	expenseSelectStmt      *sql.Stmt
	expenseSelectCountStmt *sql.Stmt
}

func NewStorage(ctx context.Context, conf config.DatabaseConf) messages.Storage {
	return &storage{
		ctx: ctx,
		dsn: conf.Dsn,
	}
}

func (s *storage) ensureDBConnected() error {
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

func (s *storage) Add(ex expenses.Expense) error {
	if err := s.ensureDBConnected(); err != nil {
		return err
	}

	category, found, err := s.findCategory(ex.Category)
	if err != nil {
		return errors.Wrap(err, addExpenseErrMsg)
	}

	tx, err := s.db.BeginTx(s.ctx, &sql.TxOptions{})
	if err != nil {
		return errors.Wrap(err, addExpenseErrMsg)
	}

	if !found {
		category, err = s.createNewCategory(tx, ex.Category)
		if err != nil {
			tx.Rollback()
			return errors.Wrap(err, addExpenseErrMsg)
		}
	}
	ex.CategoryID = category.ID

	if err := s.createExpense(tx, ex); err != nil {
		tx.Rollback()
		return errors.Wrap(err, addExpenseErrMsg)
	}

	tx.Commit()
	return nil
}

func (s *storage) GetExpenses(p expenses.ExpensePeriod) ([]*expenses.Expense, error) {
	if err := s.ensureDBConnected(); err != nil {
		return []*expenses.Expense{}, err
	}

	exps, err := s.findExpenses(p.GetStart(time.Now()))
	if err != nil {
		return []*expenses.Expense{}, errors.Wrap(err, getExpensesErrMsg)
	}

	return exps, nil
}

func (s *storage) findCategory(categoryName string) (expenses.ExpenseCategory, bool, error) {
	if s.categorySearchStmt == nil {
		stmt, err := s.db.PrepareContext(s.ctx, expenseCategorySearchSQL)
		if err != nil {
			return expenses.ExpenseCategory{}, false, errors.Wrap(err, findCategoryErrMsg)
		}
		s.categorySearchStmt = stmt
	}

	row := s.categorySearchStmt.QueryRowContext(s.ctx, categoryName)

	if errors.Is(row.Err(), sql.ErrNoRows) {
		return expenses.ExpenseCategory{}, false, nil
	} else if row.Err() != nil {
		return expenses.ExpenseCategory{}, false, errors.Wrap(row.Err(), findCategoryErrMsg)
	}

	var id, name string
	if err := row.Scan(&id, &name); err != nil {
		return expenses.ExpenseCategory{}, false, errors.Wrap(row.Err(), findCategoryErrMsg)
	}

	return expenses.ExpenseCategory{
		ID:   id,
		Name: name,
	}, true, nil
}

func (s *storage) createNewCategory(tx *sql.Tx, categoryName string) (expenses.ExpenseCategory, error) {
	if s.categoryInsertStmt == nil {
		stmt, err := tx.PrepareContext(s.ctx, expenseCategoryInsertSQL)
		if err != nil {
			return expenses.ExpenseCategory{}, errors.Wrap(err, createNewCategoryErrMsg)
		}
		s.categoryInsertStmt = stmt
	}

	id, err := uuid.NewUUID()
	if err != nil {
		return expenses.ExpenseCategory{}, errors.Wrap(err, createNewCategoryErrMsg)
	}

	if _, err = s.categoryInsertStmt.ExecContext(s.ctx, id.String(), categoryName); err != nil {
		return expenses.ExpenseCategory{}, errors.Wrap(err, createNewCategoryErrMsg)
	}

	return expenses.ExpenseCategory{
		ID:   id.String(),
		Name: categoryName,
	}, nil
}

func (s *storage) createExpense(tx *sql.Tx, ex expenses.Expense) error {
	if s.expenseInsertStmt == nil {
		stmt, err := tx.PrepareContext(s.ctx, expensesInsertSQL)
		if err != nil {
			return errors.Wrap(err, createNewExpenseErrMsg)
		}
		s.expenseInsertStmt = stmt
	}

	id, err := uuid.NewUUID()
	if err != nil {
		return errors.Wrap(err, createNewExpenseErrMsg)
	}

	if _, err = s.expenseInsertStmt.ExecContext(s.ctx, id.String(), ex.Amount, ex.Datetime, ex.CategoryID); err != nil {
		return errors.Wrap(err, createNewExpenseErrMsg)
	}

	return nil
}

func (s *storage) findCountExpenses(from time.Time) (int, error) {
	if s.expenseSelectCountStmt == nil {
		stmt, err := s.db.PrepareContext(s.ctx, expensesSelectCountSQL)
		if err != nil {
			return 0, errors.Wrap(err, expenseSelectCountErrMsg)
		}
		s.expenseSelectCountStmt = stmt
	}

	row := s.expenseSelectCountStmt.QueryRowContext(s.ctx, from)
	if errors.Is(row.Err(), sql.ErrNoRows) {
		return 0, nil
	} else if row.Err() != nil {
		return 0, errors.Wrap(row.Err(), expenseSelectCountErrMsg)
	}

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, errors.Wrap(row.Err(), expenseSelectCountErrMsg)
	}

	return count, nil
}

func (s *storage) findExpenses(from time.Time) ([]*expenses.Expense, error) {
	if s.expenseSelectStmt == nil {
		stmt, err := s.db.PrepareContext(s.ctx, expensesSelectSQL)
		if err != nil {
			return []*expenses.Expense{}, errors.Wrap(err, expenseSelectErrMsg)
		}
		s.expenseSelectStmt = stmt
	}

	count, err := s.findCountExpenses(from)
	if err != nil {
		return []*expenses.Expense{}, err
	}

	rows, err := s.expenseSelectStmt.QueryContext(s.ctx, from)
	if err != nil {
		return []*expenses.Expense{}, errors.Wrap(err, expenseSelectErrMsg)
	}
	defer rows.Close()

	exps := make([]*expenses.Expense, 0, count)

	for rows.Next() {
		var id, categoryID, categoryName string
		var amount int
		var datetime time.Time

		rows.Scan(&id, &amount, &datetime, &categoryID, &categoryName)

		exps = append(exps, &expenses.Expense{
			ID:         id,
			Amount:     amount,
			Datetime:   datetime,
			CategoryID: categoryID,
			Category:   categoryName,
		})
	}

	return exps, nil
}
