create schema smoobites;
use smoobites;

CREATE TABLE users (
	name char NOT NULL,
	email VARCHAR(100) NOT NULL UNIQUE,
    password VARCHAR(100) NOT NULL
);
