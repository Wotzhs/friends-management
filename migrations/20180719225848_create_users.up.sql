CREATE TABLE users (
	id uuid primary key,
	email varchar unique not null,
	created_at timestamp not null,
	updated_at timestamp not null
);