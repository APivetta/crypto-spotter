-- +goose Up
-- +goose StatementBegin
CREATE TABLE
    trade_results (
        id SERIAL PRIMARY KEY,
        asset VARCHAR,
        date TIMESTAMP NOT NULL,
        result DOUBLE PRECISION NOT NULL
    );

CREATE INDEX idx_trade_results_asset_date ON trade_results (asset, date);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS trade_results;

-- +goose StatementEnd