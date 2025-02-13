-- +goose Up
-- +goose StatementBegin
CREATE TABLE
    genomes (
        id SERIAL PRIMARY KEY,
        asset VARCHAR,
        date TIMESTAMP NOT NULL,
        genome JSONB NOT NULL
    );

CREATE INDEX idx_genomes_asset_date ON genomes (asset, date);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS genomes;

-- +goose StatementEnd