package indicators

type Action int

const (
	STRONG_SELL Action = -2
	SELL        Action = -1
	HOLD        Action = 0
	BUY         Action = 1
	STRONG_BUY  Action = 2
)
