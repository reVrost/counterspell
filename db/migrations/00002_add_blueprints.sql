-- +goose Up
-- +goose StatementBegin
CREATE TABLE blueprints (
    id SERIAL PRIMARY KEY,                  -- Auto-incrementing unique ID
    blueprint_name VARCHAR(255) NOT NULL UNIQUE, -- Human-readable name for the blueprint
    config TEXT NOT NULL,                   -- The full YAML config as a string
    version INTEGER DEFAULT 1,              -- For versioning blueprints
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB                          -- Optional: Store extra parsed metadata as JSON
);
CREATE INDEX idx_blueprints_created_at ON blueprints (created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS blueprints;
-- +goose StatementEnd
