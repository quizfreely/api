-- migrate:up
ALTER TYPE auth_type_enum RENAME TO auth_type;

-- migrate:down
ALTER TYPE auth_type RENAME TO auth_type_enum;

