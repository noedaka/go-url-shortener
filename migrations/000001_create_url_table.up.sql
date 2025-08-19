CREATE TABLE urls (
	id SERIAL PRIMARY KEY,
    short_url TEXT NOT NULL,
    original_url TEXT NOT NULL
);