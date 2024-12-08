# Online Song Library

Online Song Library Server.

## Installation

1. Go 1.22 or higher required.
2. Setup database, for example:
```bash
# install database
brew install postgresql@16
brew services start postgresql@16

# set database user
psql postgres
create database osl;
create user osl with encrypted password 'osl';
grant all privileges on database osl to osl;
alter database osl owner to osl;
```
3. Set environment variables in .env:
```bash
# Server address: http://localhost:8080
SERVER_URI=:8080
# Database address: user osl, password osl, database osl
DB_URI=postgresql://osl:osl@localhost:5432/osl
# Music info url: http://localhost:8081
MUSIC_INFO_URL=http://localhost:8081
```

## Usage
Test server:
```
make test
```
Run server:
```
make run
```
Try it out: http://localhost:8080/swagger/index.html#/
## License

[MIT](https://choosealicense.com/licenses/mit/)