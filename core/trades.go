package opencoindata

import (
	"database/sql"
	"log"
	"time"

	"github.com/coopernurse/gorp"
	"github.com/davecgh/go-spew/spew"
	"github.com/lox/babelcoin/core"
	_ "github.com/lox/babelcoin/exchanges/btce"
	_ "github.com/lox/babelcoin/exchanges/cryptsy"
	butil "github.com/lox/babelcoin/util"
	_ "github.com/ziutek/mymysql/godrv"
	"github.com/ziutek/mymysql/mysql"
)

const (
	TRADE_DB = "trades/root/"
)

var db *sql.DB

type tradeRow struct {
	Id       string  `db:"tid"`
	Amount   float64 `db:"amount"`
	Rate     float64 `db:"rate"`
	Unixtime int64   `db:"time"`
}

type tradeCollector struct {
	Pairs       []babelcoin.Pair
	Exchange    babelcoin.Exchange
	ExchangeKey string
	DbMaps      map[babelcoin.Pair]*gorp.DbMap
}

type Trade struct {
	babelcoin.Trade
}

func (t *Trade) String() string {
	return t.Trade.String()
}

func TradeCollectors(exchanges []string) chan tradeCollector {
	ch := make(chan tradeCollector)

	go func() {
		for _, exchangeKey := range exchanges {
			ex, err := babelcoin.NewExchange(exchangeKey, babelcoin.EnvExchangeConfig(exchangeKey))
			if err != nil {
				panic(err)
			}

			pairs, err := ex.Pairs()
			if err != nil {
				panic(err)
			}

			t := tradeCollector{pairs, ex, exchangeKey, map[babelcoin.Pair]*gorp.DbMap{}}

			for _, pair := range pairs {
				dbMap, err := newTradeDb(exchangeKey, pair)
				if err != nil {
					panic(err)
				}

				t.DbMaps[pair] = dbMap
			}

			ch <- t
		}
		close(ch)
	}()

	return ch
}

func (i tradeCollector) Collect(interval time.Duration) chan Trade {
	log.Printf("Collecting trades for %d pairs from %v", len(i.Pairs), i.ExchangeKey)
	source := make(chan babelcoin.Trade, 1)
	out := make(chan Trade, 1)

	butil.HistoryPoller(i.Exchange, i.Pairs, interval, source)
	go func() {
		for t := range source {
			err, skipped := i.save(&Trade{t})
			if err != nil {
				log.Printf("Error saving trade: %s", err)
			}
			if !skipped {
				out <- Trade{t}
			} else {
				log.Printf("Skipping %v, exists in the db", t.Identity())
			}
		}
	}()

	return out
}

func (i tradeCollector) save(t *Trade) (error, bool) {
	trade := &tradeRow{t.Id, t.Amount, t.Rate, t.Timestamp.Unix()}

	if err := i.DbMaps[t.Pair].Insert(trade); err != nil {
		if err.(*mysql.Error).Code == mysql.ER_DUP_ENTRY {
			return nil, true
		}

		return err, false
	}

	return nil, false
}

func dbSingleton() (*sql.DB, error) {
	if db == nil {
		d, err := sql.Open("mymysql", TRADE_DB)
		if err != nil {
			return nil, err
		} else {
			db = d
		}
	}

	return db, nil
}

func newTradeDb(exchange string, pair babelcoin.Pair) (*gorp.DbMap, error) {
	db, err := dbSingleton()
	if err != nil {
		return nil, err
	}

	dbMap := &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}
	tableName := exchange + "-" + pair.String()
	dbMap.AddTableWithName(tradeRow{}, tableName).SetKeys(false, "Id")

	if err = dbMap.CreateTablesIfNotExists(); err != nil {
		return nil, err
	}

	return dbMap, nil
}

func GetPairStatus() (map[babelcoin.Pair]time.Time, error) {
	db, err := dbSingleton()
	if err != nil {
		return map[babelcoin.Pair]time.Time{}, err
	}

	dbMap := &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}
	spew.Dump(dbMap)

	return map[babelcoin.Pair]time.Time{}, nil
}
