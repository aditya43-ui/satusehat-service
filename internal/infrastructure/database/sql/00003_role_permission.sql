-- +goose Up
-- +goose StatementBegin
CREATE TABLE rol_permission (
	id int4 NOT NULL,
	"create" bool NULL,
	"read" bool NULL,
	"update" bool NULL,
	"disable" bool NULL,
	"delete" bool NULL,
	active bool NULL,
	fk_rol_pages_id int4 NULL,
	role_keycloak _text NULL,
	group_keycloak _text NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
