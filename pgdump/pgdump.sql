CREATE TABLE "token" (
                         "id" integer PRIMARY KEY,
                         "user_id" integer,
                         "token" varchar,
                         "expires_at" timestamp
);

CREATE TABLE "user" (
                        "id" integer PRIMARY KEY,
                        "name" varchar,
                        "email" varchar,
                        "password" varchar,
                        "created_at" timestamp
);

CREATE TABLE "medicine" (
                            "id" integer PRIMARY KEY,
                            "global_id" varchar,
                            "name" varchar,
                            "dosage" varchar,
                            "form" varchar,
                            "active_ingredient" varchar,
                            "pharma_company" varchar
);

ALTER TABLE "token" ADD FOREIGN KEY ("user_id") REFERENCES "user" ("id");