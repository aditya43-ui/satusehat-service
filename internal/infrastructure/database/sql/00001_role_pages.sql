-- +goose Up
-- +goose StatementBegin
CREATE SEQUENCE IF NOT EXISTS rol_pages_id_seq;

CREATE TABLE IF NOT EXISTS rol_pages (
	id int4 NOT NULL DEFAULT nextval('rol_pages_id_seq'),
	"name" varchar(20) NOT NULL,
	icon varchar(100) NULL,
	url text NOT NULL,
	"level" int2 NOT NULL DEFAULT 0,
	sort int2 NOT NULL DEFAULT 0,
	parent int4 NULL,
	active bool NOT NULL DEFAULT true,
	created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at timestamp NULL,
	deleted_at timestamp NULL,
	CONSTRAINT rol_pages_pkey PRIMARY KEY (id),
	CONSTRAINT rol_pages_parent_fkey FOREIGN KEY (parent) REFERENCES rol_pages(id) ON DELETE CASCADE
);

-- Index untuk performa query
CREATE INDEX idx_rol_pages_parent ON rol_pages(parent) WHERE deleted_at IS NULL;
CREATE INDEX idx_rol_pages_level ON rol_pages(level) WHERE deleted_at IS NULL;
CREATE INDEX idx_rol_pages_active ON rol_pages(active) WHERE deleted_at IS NULL;
CREATE INDEX idx_rol_pages_sort ON rol_pages(sort) WHERE deleted_at IS NULL;

-- Trigger untuk update updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$ language 'plpgsql';

CREATE TRIGGER update_rol_pages_updated_at BEFORE UPDATE ON rol_pages
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_rol_pages_updated_at ON rol_pages;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP INDEX IF EXISTS idx_rol_pages_parent;
DROP INDEX IF EXISTS idx_rol_pages_level;
DROP INDEX IF EXISTS idx_rol_pages_active;
DROP INDEX IF EXISTS idx_rol_pages_sort;
DROP TABLE IF EXISTS rol_pages;
DROP SEQUENCE IF EXISTS rol_pages_id_seq;
-- +goose StatementEnd