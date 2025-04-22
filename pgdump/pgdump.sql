CREATE TABLE "refresh_tokens" (
                                  "id" SERIAL PRIMARY KEY,
                                  "user_id" integer,
                                  "token" varchar,
                                  "expires_at" timestamp
);

CREATE TABLE "users" (
                         "id" SERIAL PRIMARY KEY,
                         "name" varchar,
                         "email" varchar UNIQUE,
                         "password" varchar,
                         "registered_at" timestamp
);

CREATE TABLE "medicines" (
                             "id" SERIAL PRIMARY KEY,
                             "ndc" varchar NOT NULL,
                             "name" varchar,
                             "dosage" varchar,
                             "form" varchar,
                             "active_ingredient" varchar,
                             "pharma_company" varchar
);

ALTER TABLE "refresh_tokens" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");
