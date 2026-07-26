# Quizfreely API

This is the GraphQL API for Quizfreely, a free and open source studying tool.

[quizfreely.org](https://quizfreely.org)

[Codeberg](https://codeberg.org/quizfreely/api) · [GitHub](https://github.com/quizfreely/api)

Quizfreely's frontend web app source code repository is `quizfreely/quizfreely` [on Codeberg](https://codeberg.org/quizfreely/quizfreely) and [on GitHub](https://github.com/quizfreely/quizfreely).

---

### First-time contributor/developer setup

Make sure you have PostgreSQL installed (v15.18 or higher)

Install [dbmate](https://github.com/amacneil/dbmate#installation)

Clone this repository, then copy `.env.example` to `.env`.

Copy `config.example.toml` to `config.toml`

Create PostgreSQL roles named `quizfreely_db_admin` and `quizfreely_api`. You can use `psql`.
```bash
# systemctl start postgresql
sudo -u postgres psql
```
Now, inside your `psql` database shell, create our roles and our database:
```sql
CREATE ROLE quizfreely_db_admin LOGIN PASSWORD 'your_new_password_here';
CREATE ROLE quizfreely_api LOGIN PASSWORD 'password_here';

CREATE DATABASE quizfreely_db OWNER quizfreely_db_admin;
GRANT CONNECT ON DATABASE quizfreely_db TO quizfreely_api;

-- Exit with `\q` when done
\q
```

Now, edit `.env`. Put your new password for `quizfreely_db_admin` and your database name in `DB_MIGRATION_URL`.
```bash
DB_MIGRATION_URL="postgres://quizfreely_db_admin:your_new_password_here@localhost:5432/quizfreely_db?sslmode=disable"
```

Next, edit `config.toml`. Pur your new password for `quizfreely_api` and your same database name in `db_url`.
```toml
db_url = 'postgres://quizfreely_api:password_here@localhost:5432/quizfreely_db'
```

There are lots of helpful and detailed comments inside of config.toml for all the other options.

If you don't change anything else in config.toml or .env, everything should "just work" by default.

Now use dbmate to set up your database. (Run this inside your `api` folder you got after this repository was cloned)
```bash
# cd api

dbmate -e DB_MIGRATION_URL migrate
```

Next, run `subjects.sql` and then `subject_keywords.sql` with `psql`.
```bash
psql -f db/subjects.sql -d postgres://quizfreely_api:password_here@localhost:5432/quizfreely_db
psql -f db/subject_keywords.sql -d postgres://quizfreely_api:password_here@localhost:5432/quizfreely_db
```

After that, we're done and you can run `go run main.go`

### Running the API

Make sure PostgreSQL is running:
```bash
systemctl status postgresql

# start postgres if you need to:
# systemctl start postgresql

# enable postgres if you want it to automatically start on boot:
# systemctl enable postgresql
```

Use `go run` to start the API.
```bash
go run main.go
```

