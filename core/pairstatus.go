package opencoindata

import (
	"strings"
	"time"

	"github.com/coopernurse/gorp"
	"github.com/lox/babelcoin/core"
)

type PairStatus struct {
	Pair      babelcoin.Pair `json:"-"`
	Trades    int64          `json:"trades"`
	LastTrade int64          `json:"lasttrade"`
	Status    string         `json:"status"`
}

// returns the status of all pairs for all exchanges
func GetAllPairStatus() (map[string]map[string]PairStatus, error) {
	db, err := dbSingleton()
	if err != nil {
		return nil, err
	}

	var tables []string
	dbMap := &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}
	if _, err = dbMap.Select(&tables, "SHOW TABLES"); err != nil {
		return nil, err
	}

	var result = map[string]map[string]PairStatus{}
	for _, table := range tables {
		z := strings.SplitN(table, "-", 2)

		if result[z[0]] == nil {
			result[z[0]] = map[string]PairStatus{}
		}

		result[z[0]][z[1]], err = GetPairStatus(z[0], babelcoin.ParsePair(z[1]))
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func GetPairStatus(exchange string, pair babelcoin.Pair) (PairStatus, error) {
	db, err := dbSingleton()
	if err != nil {
		return PairStatus{}, err
	}

	dbMap := &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}
	table := exchange + "-" + pair.String()

	var lastTradeTime int64
	err = dbMap.SelectOne(&lastTradeTime, "SELECT MAX(time) FROM `"+table+"`")
	if err != nil {
		return PairStatus{}, err
	}

	var tradeCount int64
	err = dbMap.SelectOne(&tradeCount, "SELECT COUNT(tid) FROM `"+table+"`")
	if err != nil {
		return PairStatus{}, err
	}

	status := "ok"
	t := time.Unix(lastTradeTime, 0)

	if time.Now().Add(-(time.Minute * 5)).After(t) {
		status = "warning"
	} else if time.Now().Add(-(time.Minute * 15)).After(t) {
		status = "error"
	}

	return PairStatus{
		Pair:      pair,
		Trades:    tradeCount,
		LastTrade: lastTradeTime,
		Status:    status,
	}, nil
}
