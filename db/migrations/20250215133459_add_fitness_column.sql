-- +goose Up
-- +goose StatementBegin
ALTER TABLE genomes ADD COLUMN fitness DOUBLE PRECISION DEFAULT 0.0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE genomes DROP COLUMN fitness;
-- +goose StatementEnd
