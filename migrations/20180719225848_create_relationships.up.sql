CREATE TABLE relationships (
	id serial primary key,
	requestor varchar unique not null,
	target varchar unique not null,
	status varchar not null,
	created_at timestamp not null,
	updated_at timestamp not null
);