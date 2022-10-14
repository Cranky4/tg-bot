-- +goose Up
-- +goose StatementBegin
DROP TABLE expenses_limits;

CREATE TABLE expenses_limits (
    category_id uuid not null
        constraint fk_expenses_limits_category_id_expense_category_id
            references expense_categories
            on delete cascade,
    user_id int not null,
    amount int not null,
    created_at timestamp not null default now(),

    CONSTRAINT pr_expense_category_id_user_id PRIMARY KEY (category_id, user_id)
);

TRUNCATE TABLE expenses;
ALTER TABLE expenses ADD COLUMN user_id int not null;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE expenses_limits DROP COLUMN user_id;
ALTER TABLE expenses DROP user_id;
-- +goose StatementEnd
