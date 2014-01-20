package opencoindata

import (
	"database/sql"
	"log"
	"time"

	"github.com/coopernurse/gorp"
	"github.com/lox/babelcoin/core"
	_ "github.com/lox/babelcoin/exchanges/btce"
	butil "github.com/lox/babelcoin/util"
	_ "github.com/ziutek/mymysql/godrv"
	"github.com/ziutek/mymysql/mysql"
)

const (
	TRADE_DB = "trades/root/"
)

type tradeRow struct {
	Id       string  `db:"tid"`
	Amount   float64 `db:"amount"`
	Rate     float64 `db:"rate"`
	Unixtime int64   `db:"time"`
}

type tradeCollector struct {
	babelcoin.Pair
	Exchange    babelcoin.Exchange
	ExchangeKey string
	DbMap       *gorp.DbMap
}

func (i tradeCollector) Collect(interval time.Duration) chan babelcoin.Trade {
	source := make(chan babelcoin.Trade, 1)
	out := make(chan babelcoin.Trade, 1)

	butil.HistoryPoller(i.Exchange, i.Pair, interval, source)
	go func() {
		for t := range source {
			err, skipped := i.save(&tradeRow{t.Id, t.Amount, t.Rate, t.Timestamp.Unix()})
			if err != nil {
				log.Printf("Error saving trade: %s", err)
			}
			if !skipped {
				out <- t
			}
		}
	}()

	return out
}

func (i tradeCollector) save(t *tradeRow) (error, bool) {
	if err := i.DbMap.Insert(t); err != nil {
		if err.(*mysql.Error).Code == mysql.ER_DUP_ENTRY {
			return nil, true
		}

		return err, false
	}

	return nil, false
}

func TradeCollectors(exchanges []string) chan tradeCollector {
	ch := make(chan tradeCollector)

	go func() {
		for _, e := range exchanges {
			ex, err := babelcoin.NewExchange(e, map[string]interface{}{})
			if err != nil {
				panic(err)
			}

			pairs, err := ex.Pairs()
			if err != nil {
				panic(err)
			}

			for _, pair := range pairs {
				dbMap, err := initTradeDb(e, pair)
				if err != nil {
					panic(err)
				}

				ch <- tradeCollector{pair, ex, e, dbMap}
			}
		}
		close(ch)
	}()

	return ch
}

func initTradeDb(exchange string, pair babelcoin.Pair) (*gorp.DbMap, error) {
	db, err := sql.Open("mymysql", TRADE_DB)
	if err != nil {
		return nil, err
	}

	tableName := exchange + "-" + pair.String()
	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}
	dbmap.AddTableWithName(tradeRow{}, tableName).SetKeys(false, "Id")

	if err = dbmap.CreateTablesIfNotExists(); err != nil {
		return nil, err
	}

	return dbmap, nil
}
