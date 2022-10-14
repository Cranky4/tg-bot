-- +goose Up
-- +goose StatementBegin
CREATE TABLE expense_categories (
    id uuid not null primary key,
    name varchar(255) not null,
    created_at timestamp not null default now()
);

CREATE TABLE expenses (
    id uuid not null primary key,
    amount int not null,
    datetime timestamp not null,
    category_id uuid not null
        constraint fk_expenses_category_id_expense_category_id
            references expense_categories
            on delete set null,
    created_at timestamp not null default now()
);

CREATE INDEX idx_expenses_category_id on expenses USING hash (category_id); -- потому что выборка, как точное сравнение (либо btree, так как он тодже покрывает равенство)
CREATE INDEX idx_expenses_datetime on expenses (datetime); -- потому что  стандартный btree индекс подходит для диапазонов

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE expenses;
DROP TABLE expense_categories;
-- +goose StatementEnd
