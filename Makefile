train:
	go run src/cmd/train/main.go $(ARGS)

fetch:
	go run src/cmd/fetch_snapshots/main.go $(ARGS)

backtest:
	go run src/cmd/backtest/main.go $(ARGS)

.PHONY: train fetch backtest