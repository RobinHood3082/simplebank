# SimpleBank

Banking API sample. (Work-in-progress, not very organized yet).

## Features

- User registration and authentication
- Account creation and management
- Deposit and withdrawal functionality
- Transaction history tracking
- Email notification
- Role-based admin control

## Installation

To run the project, follow these steps:

1. Run `make postgres` to run the postgres server in docker with the necessary envs.
2. Run `make createdb` to create the DB in postgres.
3. Run `make migrateup` to run the DB migrations.
4. Now you are ready to use `make server` any time you want to run the server.

To add new migrations:
`migrate create -ext sql -dir internal/db/migration -seq <migration_name>`
