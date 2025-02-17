# crypto-spotter

## Migrate DB

```
# start postgres
docker-compose up -d

# run goose migrations
goose -dir db/migrations $DB_DRIVER $DB_STRING up
```

## Populate DB with snapshots

```
# count is how many coins we want to fetch, they are fetch in popularity order from Binance (1st is BTC)
go run src/cmd/fetch_snapshots/main.go --count=1
```

## Run genetic algorithm to train the stategy weights
```
go run src/cmd/train/main.go --days 3 --count=1
```

## Backtest
```
go run src/cmd/backtest/main.go --days=1 --asset=BTCUSDT
```

## Live run
WIP