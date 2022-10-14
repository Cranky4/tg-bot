-- +goose Up
-- +goose StatementBegin
CREATE TABLE expenses_limits (
    category_id uuid not null primary key
        constraint fk_expenses_limits_category_id_expense_category_id
            references expense_categories
            on delete cascade,
    amount int not null,
    created_at timestamp not null default now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE expenses_limits;
-- +goose StatementEnd
