### Database setup

- Generated ERD: https://dbdiagram.io/d/63cb8d49296d97641d7b24f9
- Run docker for postgres: `docker compose up`, db container name should be _go-simple-bank-db-1_
- To connect to the db in docker: `docker exec -it go-simple-bank-db-1 psql -U root`
- Inside db, run `select now();` to show date time, run `\q` to quit pg container
- To show logs: `docker logs go-simple-bank-db-1`
- Use TablePlus to open database as user `root`
- Run the exported sql query from _dbdiagram_ to create tables
