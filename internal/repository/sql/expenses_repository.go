package memorystorage

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/config"
	repo "gitlab.ozon.dev/cranky4/tg-bot/internal/repository"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/utils/expenses"
)

const (
	expenseCategorySearchSQL = "SELECT id, name FROM expense_categories WHERE name ILIKE $1 LIMIT 1"
	expenseCategoryInsertSQL = "INSERT INTO expense_categories(id, name) VALUES ($1, $2)"
	expensesInsertSQL        = "INSERT INTO expenses(id, amount, datetime, category_id) VALUES ($1,$2,$3,$4)"
	expensesSelectSQL        = "SELECT e.id, e.amount, e.datetime, c.id as categoryId, c.name FROM expenses e " +
		"INNER JOIN expense_categories c ON e.category_id = c.id WHERE datetime > $1"
	expensesSelectCountSQL = "SELECT COUNT(id) FROM expenses WHERE datetime > $1"
	upsertLimitSQL         = `INSERT INTO expenses_limits (category_id, amount) 
		VALUES($1,$2) ON CONFLICT (category_id) 
		DO UPDATE SET amount = EXCLUDED.amount`
	freeLimitSQL = `SELECT el.amount - SUM(e.amount) FROM expenses e
		LEFT JOIN expenses_limits el  ON e.category_id = el.category_id
		WHERE e.category_id = $1 AND e.datetime >= date_trunc('month', now())
		GROUP BY el.category_id;
	`

	addExpenseErrMsg                = "ошибка в методе addExpense"
	findCategoryErrMsg              = "ошибка в методе findCategory"
	createNewCategoryErrMsg         = "ошибка в методе createNewCategory"
	createNewExpenseErrMsg          = "ошибка в методе createNewExpense"
	getExpensesErrMsg               = "ошибка в методе getExpenses"
	expenseSelectErrMsg             = "ошибка в методе findExpenses"
	expenseSelectCountErrMsg        = "ошибка в методе findCountExpenses"
	setLimitErrMsg                  = "ошибка в методе setLimit"
	upsertLimitErrMsg               = "ошибка в методе upsertLimit"
	limitReachedErrMsg              = "ошибка в методе limitReached"
	freeLimitErrMsg                 = "ошибка в методе findFreeLimit"
	cannotRollbackTransactionErrMsg = "ошибка отката транзакции"
)

type repository struct {
	dsn string
	db  *sql.DB

	categorySearchStmt     *sql.Stmt
	expenseSelectStmt      *sql.Stmt
	expenseSelectCountStmt *sql.Stmt
	freeLimitStmt          *sql.Stmt
}

func NewRepository(conf config.DatabaseConf) repo.ExpensesRepository {
	return &repository{
		dsn: conf.Dsn,
	}
}

func (r *repository) ensureDBConnected() error {
	if r.db != nil {
		return nil
	}

	db, err := sql.Open("pgx", r.dsn)
	if err != nil {
		return err
	}
	r.db = db

	err = r.db.Ping()
	if err != nil {
		return err
	}

	return nil
}

func (r *repository) Add(ctx context.Context, ex expenses.Expense) error {
	var err error

	if err = r.ensureDBConnected(); err != nil {
		return err
	}

	category, found, err := r.findCategory(ctx, ex.Category)
	if err != nil {
		return errors.Wrap(err, addExpenseErrMsg)
	}

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return errors.Wrap(err, addExpenseErrMsg)
	}

	defer func() {
		if err != nil {
			err = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	if !found {
		category, err = r.createNewCategory(ctx, tx, ex.Category)
		if err != nil {
			if tErr := tx.Rollback(); tErr != nil {
				return errors.Wrap(tErr, cannotRollbackTransactionErrMsg)
			}
			return errors.Wrap(err, addExpenseErrMsg)

		}
	}
	ex.CategoryID = category.ID

	if err = r.createExpense(ctx, tx, ex); err != nil {
		return errors.Wrap(err, addExpenseErrMsg)
	}

	return err
}

func (r *repository) GetExpenses(ctx context.Context, p expenses.ExpensePeriod) ([]*expenses.Expense, error) {
	if err := r.ensureDBConnected(); err != nil {
		return []*expenses.Expense{}, err
	}

	exps, err := r.findExpenses(ctx, p.GetStart(time.Now()))
	if err != nil {
		return []*expenses.Expense{}, errors.Wrap(err, getExpensesErrMsg)
	}

	return exps, nil
}

func (r *repository) SetLimit(ctx context.Context, categoryName string, amount int64) error {
	var err error

	if err = r.ensureDBConnected(); err != nil {
		return err
	}

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return errors.Wrap(err, setLimitErrMsg)
	}

	defer func() {
		if err != nil {
			err = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	category, found, err := r.findCategory(ctx, categoryName)
	if err != nil {
		return errors.Wrap(err, setLimitErrMsg)
	}

	if !found {
		category, err = r.createNewCategory(ctx, tx, categoryName)
		if err != nil {
			return errors.Wrap(err, setLimitErrMsg)
		}
	}

	if err = r.upsertLimit(ctx, tx, category.ID, amount); err != nil {
		return errors.Wrap(err, setLimitErrMsg)
	}

	return err
}

func (r *repository) GetFreeLimit(ctx context.Context, categoryName string) (int64, bool, error) {
	if err := r.ensureDBConnected(); err != nil {
		return 0, false, err
	}

	category, found, err := r.findCategory(ctx, categoryName)
	if err != nil {
		return 0, false, errors.Wrap(err, limitReachedErrMsg)
	}

	if !found {
		return 0, false, nil
	}

	return r.findFreeLimit(ctx, category.ID)
}

func (r *repository) findCategory(ctx context.Context, categoryName string) (expenses.ExpenseCategory, bool, error) {
	if r.categorySearchStmt == nil {
		stmt, err := r.db.PrepareContext(ctx, expenseCategorySearchSQL)
		if err != nil {
			return expenses.ExpenseCategory{}, false, errors.Wrap(err, findCategoryErrMsg)
		}
		r.categorySearchStmt = stmt
	}

	row := r.categorySearchStmt.QueryRowContext(ctx, categoryName)

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

func (r *repository) createNewCategory(ctx context.Context, tx *sql.Tx, categoryName string) (expenses.ExpenseCategory, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return expenses.ExpenseCategory{}, errors.Wrap(err, createNewCategoryErrMsg)
	}

	if _, err = tx.ExecContext(ctx, expenseCategoryInsertSQL, id.String(), categoryName); err != nil {
		return expenses.ExpenseCategory{}, errors.Wrap(err, createNewCategoryErrMsg)
	}

	return expenses.ExpenseCategory{
		ID:   id.String(),
		Name: categoryName,
	}, nil
}

func (r *repository) createExpense(ctx context.Context, tx *sql.Tx, ex expenses.Expense) error {
	id, err := uuid.NewUUID()
	if err != nil {
		return errors.Wrap(err, createNewExpenseErrMsg)
	}

	if _, err = tx.ExecContext(ctx, expensesInsertSQL, id.String(), ex.Amount, ex.Datetime, ex.CategoryID); err != nil {
		return errors.Wrap(err, createNewExpenseErrMsg)
	}

	return nil
}

func (r *repository) findCountExpenses(ctx context.Context, from time.Time) (int, error) {
	if r.expenseSelectCountStmt == nil {
		stmt, err := r.db.PrepareContext(ctx, expensesSelectCountSQL)
		if err != nil {
			return 0, errors.Wrap(err, expenseSelectCountErrMsg)
		}
		r.expenseSelectCountStmt = stmt
	}

	row := r.expenseSelectCountStmt.QueryRowContext(ctx, from)
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

func (r *repository) findExpenses(ctx context.Context, from time.Time) ([]*expenses.Expense, error) {
	if r.expenseSelectStmt == nil {
		stmt, err := r.db.PrepareContext(ctx, expensesSelectSQL)
		if err != nil {
			return []*expenses.Expense{}, errors.Wrap(err, expenseSelectErrMsg)
		}
		r.expenseSelectStmt = stmt
	}

	count, err := r.findCountExpenses(ctx, from)
	if err != nil {
		return []*expenses.Expense{}, err
	}

	rows, err := r.expenseSelectStmt.QueryContext(ctx, from)
	if err != nil {
		return []*expenses.Expense{}, errors.Wrap(err, expenseSelectErrMsg)
	}

	defer func() {
		err = rows.Close()
	}()

	exps := make([]*expenses.Expense, 0, count)

	for rows.Next() {
		var id, categoryID, categoryName string
		var amount int64
		var datetime time.Time

		if err = rows.Scan(&id, &amount, &datetime, &categoryID, &categoryName); err != nil {
			return []*expenses.Expense{}, errors.Wrap(err, expenseSelectErrMsg)
		}

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

func (r *repository) upsertLimit(ctx context.Context, tx *sql.Tx, categoryID string, amount int64) error {
	if _, err := tx.ExecContext(ctx, upsertLimitSQL, categoryID, amount); err != nil {
		return errors.Wrap(err, upsertLimitErrMsg)
	}

	return nil
}

func (r *repository) findFreeLimit(ctx context.Context, categoryID string) (int64, bool, error) {
	if r.freeLimitStmt == nil {
		stmt, err := r.db.PrepareContext(ctx, freeLimitSQL)
		if err != nil {
			return 0, false, errors.Wrap(err, freeLimitErrMsg)
		}
		r.freeLimitStmt = stmt
	}

	row := r.freeLimitStmt.QueryRowContext(ctx, categoryID)

	if errors.Is(row.Err(), sql.ErrNoRows) {
		return 0, false, nil
	}

	var freeLimit int64
	if err := row.Scan(&freeLimit); err != nil {
		return 0, true, errors.Wrap(err, freeLimitErrMsg)
	}

	return freeLimit, true, nil
}
