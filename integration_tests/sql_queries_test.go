package integrationtests_test

import (
	"context"
	"database/sql"
	"os"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model"
	expenses_sql_repo "gitlab.ozon.dev/cranky4/tg-bot/internal/repository/sql"

	// init pgsql.
	_ "github.com/jackc/pgx/stdlib"
)

const dateFormat = "2006-01-02 15:04:05"

var _ = Describe("Testing SQL queries", Ordered, func() {
	dsn := os.Getenv("TEST_DB_DSN")
	userId := int64(100)

	db, er := sql.Open("pgx", dsn)
	if er != nil {
		Fail(er.Error())
	}
	categoryID, er := uuid.NewUUID()
	if er != nil {
		Fail(er.Error())
	}

	category := model.ExpenseCategory{
		ID:   categoryID.String(),
		Name: "Дом",
	}

	expenseID1, er := uuid.NewUUID()
	if er != nil {
		Fail(er.Error())
	}
	expense1 := model.Expense{
		ID:         expenseID1.String(),
		Amount:     10000,
		Datetime:   time.Now(),
		Category:   category.Name,
		CategoryID: category.ID,
		UserId:     userId,
	}

	expenseID2, er := uuid.NewUUID()
	if er != nil {
		Fail(er.Error())
	}
	expense2 := model.Expense{
		ID:         expenseID2.String(),
		Amount:     10000,
		Datetime:   time.Now(),
		Category:   category.Name,
		CategoryID: category.ID,
		UserId:     userId,
	}

	It("insert expense category", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		res, err := db.ExecContext(ctx, expenses_sql_repo.ExpenseCategoryInsertSQL, category.ID, category.Name)

		Expect(err).To(BeNil())
		rows, err := res.RowsAffected()
		Expect(err).To(BeNil())
		Expect(int64(1)).To(Equal(rows))
	})

	It("search expense category", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		row := db.QueryRowContext(ctx, expenses_sql_repo.ExpenseCategorySearchSQL, category.Name)
		Expect(row.Err()).To(BeNil())
		var id, name string

		err := row.Scan(&id, &name)
		Expect(err).To(BeNil())

		Expect(category.ID).To(Equal(id))
		Expect(category.Name).To(Equal(name))
	})

	It("insert expenses", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		res, err := db.ExecContext(ctx, expenses_sql_repo.ExpensesInsertSQL, expense1.ID, expense1.Amount, expense1.Datetime, expense1.CategoryID, expense1.UserId)

		Expect(err).To(BeNil())
		rows, err := res.RowsAffected()
		Expect(err).To(BeNil())
		Expect(int64(1)).To(Equal(rows))

		ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		res, err = db.ExecContext(ctx, expenses_sql_repo.ExpensesInsertSQL, expense2.ID, expense2.Amount, expense2.Datetime, expense2.CategoryID, expense2.UserId)

		Expect(err).To(BeNil())
		rows, err = res.RowsAffected()
		Expect(err).To(BeNil())
		Expect(int64(1)).To(Equal(rows))
	})

	It("get expenses count", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		row := db.QueryRowContext(ctx, expenses_sql_repo.ExpensesSelectCountSQL, time.Now().AddDate(0, 0, -1), userId)
		Expect(row.Err()).To(BeNil())
		var count int

		err := row.Scan(&count)
		Expect(err).To(BeNil())

		Expect(2).To(Equal(count))
	})

	It("get expenses", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		rows, err := db.QueryContext(ctx, expenses_sql_repo.ExpensesSelectSQL, time.Now().AddDate(0, 0, -1), userId)

		Expect(err).To(BeNil())
		Expect(rows.Err()).To(BeNil())

		defer func() {
			err = rows.Close()
			Expect(err).To(BeNil())
		}()

		var id, categoryName, categoryId string
		var amount, user int64
		var datetime time.Time

		Expect(rows.Next()).To(BeTrue())
		err = rows.Scan(&id, &amount, &datetime, &categoryId, &categoryName, &user)
		Expect(err).To(BeNil())

		Expect(expense2.ID).To(Equal(id))
		Expect(expense2.Amount).To(Equal(amount))
		Expect(expense2.Datetime.Format(dateFormat)).To(Equal(datetime.Format(dateFormat)))
		Expect(expense2.CategoryID).To(Equal(categoryId))
		Expect(expense2.Category).To(Equal(categoryName))
		Expect(expense2.UserId).To(Equal(user))

		Expect(rows.Next()).To(BeTrue())
		err = rows.Scan(&id, &amount, &datetime, &categoryId, &categoryName, &user)
		Expect(err).To(BeNil())

		Expect(expense1.ID).To(Equal(id))
		Expect(expense1.Amount).To(Equal(amount))
		Expect(expense1.Datetime.Format(dateFormat)).To(Equal(datetime.Format(dateFormat)))
		Expect(expense1.CategoryID).To(Equal(categoryId))
		Expect(expense1.Category).To(Equal(categoryName))
		Expect(expense1.UserId).To(Equal(user))

		Expect(rows.Next()).To(BeFalse())
	})

	It("upsert limit", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		res, err := db.ExecContext(ctx, expenses_sql_repo.UpsertLimitSQL, category.ID, 15000, userId)

		Expect(err).To(BeNil())
		rows, err := res.RowsAffected()
		Expect(err).To(BeNil())
		Expect(int64(1)).To(Equal(rows))
	})

	It("select free limit", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		row := db.QueryRowContext(ctx, expenses_sql_repo.FreeLimitSQL, category.ID, userId)

		Expect(row.Err()).To(BeNil())
		var limit sql.NullInt64

		err := row.Scan(&limit)
		Expect(err).To(BeNil())

		Expect(true).To(Equal(limit.Valid))
		Expect(int64(-5000)).To(Equal(limit.Int64))
	})
})
