
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE scores (
	"id" SERIAL PRIMARY KEY,
	"created_at" TIMESTAMP DEFAULT now(),
	"name" TEXT NOT NULL,
	"score" BIGINT NOT NULL
);

CREATE TABLE api_keys (
	"id" SERIAL PRIMARY KEY,
	"apikey" TEXT NOT NULL,
	"email" TEXT NOT NULL
);


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE scores;
DROP TABLE api_keys;

