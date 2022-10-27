package expenses_sql_repo

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/config"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model"
	repo "gitlab.ozon.dev/cranky4/tg-bot/internal/repository"
)

const (
	ExpenseCategorySearchSQL = "SELECT id, name FROM expense_categories WHERE name ILIKE $1 LIMIT 1"
	ExpenseCategoryInsertSQL = "INSERT INTO expense_categories(id, name) VALUES ($1, $2)"

	ExpensesInsertSQL = "INSERT INTO expenses(id, amount, datetime, category_id, user_id) VALUES ($1,$2,$3,$4,$5)"
	ExpensesSelectSQL = "SELECT e.id, e.amount, e.datetime, c.id as categoryId, c.name, e.user_id " +
		"FROM expenses e INNER JOIN expense_categories c ON e.category_id = c.id " +
		"WHERE e.datetime > $1 AND e.user_id = $2 ORDER BY e.created_at DESC"

	ExpensesSelectCountSQL = "SELECT COUNT(id) FROM expenses WHERE datetime > $1 AND user_id = $2"
	UpsertLimitSQL         = `INSERT INTO expenses_limits (category_id, amount, user_id) 
		VALUES($1,$2,$3) ON CONFLICT (category_id, user_id) 
		DO UPDATE SET amount = EXCLUDED.amount`
	FreeLimitSQL = `SELECT el.amount - SUM(e.amount) FROM expenses e
		LEFT JOIN expenses_limits el ON e.category_id = el.category_id AND e.user_id = el.user_id
		WHERE e.category_id = $1 AND e.datetime >= date_trunc('month', now()) AND e.user_id = $2
		GROUP BY el.category_id, el.user_id;
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
	db *sql.DB
}

func NewRepository(conf config.DatabaseConf) (repo.ExpensesRepository, error) {
	db, err := sql.Open("pgx", conf.Dsn)
	if err != nil {
		return nil, err
	}

	return &repository{
		db: db,
	}, nil
}

func (r *repository) Add(ctx context.Context, ex model.Expense) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Add")
	defer span.Finish()

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
			return errors.Wrap(err, addExpenseErrMsg)
		}
	}
	ex.CategoryID = category.ID

	if err = r.createExpense(ctx, tx, ex); err != nil {
		return errors.Wrap(err, addExpenseErrMsg)
	}

	return err
}

func (r *repository) GetExpenses(ctx context.Context, p model.ExpensePeriod, userId int64) ([]*model.Expense, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GetExpenses")
	defer span.Finish()

	exps, err := r.findExpenses(ctx, p.GetStart(time.Now()), userId)
	if err != nil {
		return []*model.Expense{}, errors.Wrap(err, getExpensesErrMsg)
	}

	return exps, nil
}

func (r *repository) SetLimit(ctx context.Context, categoryName string, userId, amount int64) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "SetLimit")
	defer span.Finish()

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

	if err = r.upsertLimit(ctx, tx, category.ID, userId, amount); err != nil {
		return errors.Wrap(err, setLimitErrMsg)
	}

	return err
}

func (r *repository) GetFreeLimit(ctx context.Context, categoryName string, userId int64) (int64, bool, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GetFreeLimit")
	defer span.Finish()

	category, found, err := r.findCategory(ctx, categoryName)
	if err != nil {
		return 0, false, errors.Wrap(err, limitReachedErrMsg)
	}

	if !found {
		return 0, false, nil
	}

	return r.findFreeLimit(ctx, category.ID, userId)
}

func (r *repository) findCategory(ctx context.Context, categoryName string) (model.ExpenseCategory, bool, error) {
	row := r.db.QueryRowContext(ctx, ExpenseCategorySearchSQL, categoryName)

	if errors.Is(row.Err(), sql.ErrNoRows) {
		return model.ExpenseCategory{}, false, nil
	} else if row.Err() != nil {
		return model.ExpenseCategory{}, false, errors.Wrap(row.Err(), findCategoryErrMsg)
	}

	var id, name string
	if err := row.Scan(&id, &name); err != nil {
		return model.ExpenseCategory{}, false, errors.Wrap(row.Err(), findCategoryErrMsg)
	}

	return model.ExpenseCategory{
		ID:   id,
		Name: name,
	}, true, nil
}

func (r *repository) createNewCategory(ctx context.Context, tx *sql.Tx, categoryName string) (model.ExpenseCategory, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return model.ExpenseCategory{}, errors.Wrap(err, createNewCategoryErrMsg)
	}

	if _, err = tx.ExecContext(ctx, ExpenseCategoryInsertSQL, id.String(), categoryName); err != nil {
		return model.ExpenseCategory{}, errors.Wrap(err, createNewCategoryErrMsg)
	}

	return model.ExpenseCategory{
		ID:   id.String(),
		Name: categoryName,
	}, nil
}

func (r *repository) createExpense(ctx context.Context, tx *sql.Tx, ex model.Expense) error {
	id, err := uuid.NewUUID()
	if err != nil {
		return errors.Wrap(err, createNewExpenseErrMsg)
	}

	if _, err = tx.ExecContext(ctx, ExpensesInsertSQL, id.String(), ex.Amount, ex.Datetime, ex.CategoryID, ex.UserId); err != nil {
		return errors.Wrap(err, createNewExpenseErrMsg)
	}

	return nil
}

func (r *repository) findCountExpenses(ctx context.Context, from time.Time, userId int64) (int, error) {
	row := r.db.QueryRowContext(ctx, ExpensesSelectCountSQL, from, userId)
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

func (r *repository) findExpenses(ctx context.Context, from time.Time, userId int64) ([]*model.Expense, error) {
	count, err := r.findCountExpenses(ctx, from, userId)
	if err != nil {
		return []*model.Expense{}, err
	}

	rows, err := r.db.QueryContext(ctx, ExpensesSelectSQL, from, userId)
	if err != nil {
		return []*model.Expense{}, errors.Wrap(err, expenseSelectErrMsg)
	}

	defer rows.Close() //nolint:errcheck

	exps := make([]*model.Expense, 0, count)

	for rows.Next() {
		var id, categoryID, categoryName string
		var userId, amount int64
		var datetime time.Time

		if err = rows.Scan(&id, &amount, &datetime, &categoryID, &categoryName, &userId); err != nil {
			return []*model.Expense{}, errors.Wrap(err, expenseSelectErrMsg)
		}

		exps = append(exps, &model.Expense{
			ID:         id,
			Amount:     amount,
			Datetime:   datetime,
			CategoryID: categoryID,
			Category:   categoryName,
			UserId:     userId,
		})
	}

	return exps, nil
}

func (r *repository) upsertLimit(ctx context.Context, tx *sql.Tx, categoryID string, userId, amount int64) error {
	if _, err := tx.ExecContext(ctx, UpsertLimitSQL, categoryID, amount, userId); err != nil {
		return errors.Wrap(err, upsertLimitErrMsg)
	}

	return nil
}

func (r *repository) findFreeLimit(ctx context.Context, categoryID string, userId int64) (int64, bool, error) {
	row := r.db.QueryRowContext(ctx, FreeLimitSQL, categoryID, userId)

	if errors.Is(row.Err(), sql.ErrNoRows) {
		return 0, false, nil
	}

	var freeLimit sql.NullInt64
	if err := row.Scan(&freeLimit); err != nil {
		return 0, true, errors.Wrap(err, freeLimitErrMsg)
	}

	if !freeLimit.Valid {
		return 0, false, nil
	}

	return freeLimit.Int64, true, nil
}
