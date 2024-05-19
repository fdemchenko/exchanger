# Exchange rate API

`GET /rate` - get USD to UAH exchange rate

`POST /subscribe` - subscribe to exchange rate update (send application/x-www-form-urlencoded email address)


## Running application

Edit `docker-compose.yml` file and put your SMTP credentials, to be able to send updates to subscribers

To start http server and PostgreSQL service run: - `docker compose up`

## TODO

- [ ] Graceful server shutdown
- [ ] SSL encryption
- [ ] Parallel emails sending




