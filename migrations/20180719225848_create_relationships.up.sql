CREATE TABLE relationships (
	id serial primary key,
	requestor varchar not null,
	target varchar not null,
	status varchar not null,
	created_at timestamp not null,
	updated_at timestamp not null
);