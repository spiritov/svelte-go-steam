# svelte-go-steam
> based upon https://github.com/dresswithpockets/go-steam-openid-template

> [!NOTE]
> see https://github.com/jump-fortress/site for a more complex example project

## features
- Steam OpenID auth flow
- [Huma api](https://github.com/danielgtaylor/huma) and docs with permissions middleware using sqlite3
- [migrate cli](https://github.com/golang-migrate/migrate) to manage sql migrations
- [sqlc](https://github.com/sqlc-dev/sqlc) to generate Go types and functions from sql
- [openapi-typescript](https://github.com/openapi-ts/openapi-typescript) to generate types from the apis's schema, used with openapi-fetch
- basic example backend
- bring your own Svelte framework! (Typescript & Tailwind are included here..)

## quickstart
1. install go 1.25
2. reference [.env.local.example](api/env/.env.example) to populate `/api/env/.env.local`
3. install `migrate` and `sqlc`, init sqlite3 db
```sh
cd api
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate
migrate -source file://db/migrations -database sqlite3://db/jump.db up
go run .
```
4. run Svelte site
```sh
cd web
npm i
npm run dev
```

## migrations
```sh
cd api
migrate create -ext sql -dir db/migrations change-summary
# move up to next migration version
migrate -source file://db/migrations -database sqlite3://db/jump.db up 1
```

## sqlc
1. write queries in `/db/sql` with [sqlc's basic query annotations](https://docs.sqlc.dev/en/latest/reference/query-annotations.html)
2. generate
```sh
cd api
sqlc generate
```

## filetree
important bits of the api's filetree
- db migrations are found in `/db/migrations/`
- sql queries used with sqlc are found in `/db/sql/`
- see `registerRoutes()` in `/internal/api.go` to see an example of permissions using a "dev" role
- api routes are found in `/internal/routes/routes.go`
- common data models are found in `/models/`
```console
svelte-go-steam/
├── api/
│   ├── db/
│   │   ├── migrations/
│   │   │   ├── 20260302212745_init.down.sql
│   │   │   └── 20260302212745_init.up.sql
│   │   └── sql/
│   │       ├── disallow_token.sql
│   │       ├── openid_nonce.sql
│   │       ├── session.sql
│   │       ├── user.sql
│   │       └── user_role.sql
│   ├── env/
│   │   ├── .env.example
│   │   └── env.go
│   ├── internal/
│   │   ├── api.go
│   │   ├── middleware.go
│   │   ├── routes/
│   │   │   ├── routes.go
│   │   │   └── user.go
│   ├── main.go
│   ├── models/
│   │   ├── conversions.go
│   │   ├── errors.go
│   │   ├── input.go
│   │   ├── models.go
│   │   └── output.go
│   └── sqlc.yaml
```
