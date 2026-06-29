-- +goose Up
-- +goose StatementBegin
CREATE TABLE role_access.rol_master (
	id int4 NOT NULL,
	"name" varchar(100) NULL,
	active bool NULL,
	created_at timestamp NULL,
	updated_at timestamp NULL,
	CONSTRAINT rol_master_role_pk PRIMARY KEY (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
