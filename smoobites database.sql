drop schema smoobites;
create schema smoobites;
use smoobites;

CREATE TABLE users (
	name VARCHAR(255) NOT NULL,
	email VARCHAR(100) NOT NULL UNIQUE,
    password VARCHAR(100) NOT NULL,
    role VARCHAR(50) NOT NULL
);

insert into users (name, email, password, role)
values ("khoon coffeehouse express", "khoon@smu.com", "$2a$10$g6C1dZ2Sa/MCye.pJW55J.OoIcg9bCwa.71jm7ZqT4WAJWuWfhD3S", "vendor");
