CREATE TABLE goapi.users (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username text UNIQUE NOT NULL,
    status varchar(16) NOT NULL,
    password_hash text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
