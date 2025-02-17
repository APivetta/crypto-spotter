train:
	go run src/cmd/train/main.go $(ARGS)

fetch:
	go run src/cmd/fetch_snapshots/main.go $(ARGS)

backtest:
	go run src/cmd/backtest/main.go $(ARGS)

build:
	go build -o bin/trader ./src
	go build -o bin/train ./src/cmd/train/
	go build -o bin/fetch ./src/cmd/fetch_snapshots/

start:
	make fetch ARGS="--count=1"
	make train ARGS="--count=1"
	go run src/main.go

.PHONY: train fetch backtest start build