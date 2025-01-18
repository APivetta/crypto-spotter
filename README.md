# crypto-spotter

## Migrate DB

```
# start postgres
docker-compose up -d

# run goose migrations
goose -dir db/migrations $DB_DRIVER $DB_STRING up
```


