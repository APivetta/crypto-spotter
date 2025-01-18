-- +goose Up
-- +goose StatementBegin
CREATE TABLE
    snapshots (
        id SERIAL PRIMARY KEY,
        asset VARCHAR,
        date TIMESTAMP NOT NULL,
        open DOUBLE PRECISION NOT NULL,
        high DOUBLE PRECISION NOT NULL,
        low DOUBLE PRECISION NOT NULL,
        close DOUBLE PRECISION NOT NULL,
        volume DOUBLE PRECISION NOT NULL
    );

CREATE INDEX idx_asset_date ON snapshots (asset, date);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS snapshots;

-- +goose StatementEnd-dir