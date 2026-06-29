-- +goose Up
-- +goose StatementBegin
CREATE TABLE rol_component (
	id int4 NOT NULL,
	"name" varchar(100) NULL,
	description text NULL,
	directory text NOT NULL,
	active bool NOT NULL,
	fk_rol_pages_id int4 NOT NULL,
	sort int2 NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
