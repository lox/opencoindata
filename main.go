package main

import (
	"database/sql"
	"log"
	"time"

	"github.com/coopernurse/gorp"
	"github.com/docopt/docopt.go"
	"github.com/lox/babelcoin/core"
	_ "github.com/lox/babelcoin/exchanges/btce"
	butil "github.com/lox/babelcoin/util"
	_ "github.com/ziutek/mymysql/godrv"
	"github.com/ziutek/mymysql/mysql"
)

type Trade struct {
	Id       string  `db:"tid"`
	Amount   float64 `db:"amount"`
	Rate     float64 `db:"rate"`
	Unixtime int64   `db:"time"`
}

type Pair struct {
	babelcoin.Pair
	Exchange    babelcoin.Exchange
	ExchangeKey string
}

var Exchanges = []string{
	"btce",
}

func main() {
	usage := `Open Coin Data. Cryptocoin exchange data for all. 

Usage:
  opencoindata poll [--interval=<duration>] 
  opencoindata -h | --help
  opencoindata --version

Options:
  -h --help     			Show this screen.
  -i --interval=<duration>  Time interval to use [default: 30s].
  --version     			Show version.
 `

	args, err := docopt.Parse(usage, nil, true, "Open Coin Data", false)
	if err != nil {
		panic(err)
	}

	duration, err := time.ParseDuration(args["--interval"].(string))
	if err != nil {
		panic(err)
	}

	log.Printf("Polling exchanges every %s", duration.String())

	if poll := args["poll"]; poll.(bool) {
		q := make(chan bool)
		for p := range enumPairs() {
			dbmap := initDb(p)
			go func(p Pair) {
				defer dbmap.Db.Close()
				for t := range p.pollTrades(duration) {
					if t.Id == "" {
						panic("Nil trade identifier found")
					}
					err = dbmap.Insert(&Trade{t.Id, t.Amount, t.Rate, t.Timestamp.Unix()})
					if err != nil {
						if err.(*mysql.Error).Code != mysql.ER_DUP_ENTRY {
							log.Printf("Error: %s", err.Error())
						}
					} else {
						log.Printf("%s", t.String())
					}
				}
			}(p)
		}
		<-q
	}
}

func (p Pair) tableName() string {
	return p.ExchangeKey + "-" + p.String()
}

func (p Pair) pollTrades(interval time.Duration) chan babelcoin.Trade {
	ch := make(chan babelcoin.Trade, 1)
	butil.HistoryPoller(p.Exchange, p.Pair, interval, ch)
	return ch
}

func enumPairs() chan Pair {
	ch := make(chan Pair)

	go func() {
		for _, e := range Exchanges {
			ex, err := babelcoin.NewExchange(e, map[string]interface{}{})
			checkErr(err, "Failed to create exchange")

			pairs, err := ex.Pairs()
			checkErr(err, "Failed to get pairs from exchange")

			for _, p := range pairs {
				ch <- Pair{p, ex, e}
			}
		}
		close(ch)
	}()

	return ch
}

func initDb(pair Pair) *gorp.DbMap {
	db, err := sql.Open("mymysql", "trades/root/")
	checkErr(err, "sql.Open failed")

	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}
	dbmap.AddTableWithName(Trade{}, pair.tableName()).SetKeys(false, "Id")

	err = dbmap.CreateTablesIfNotExists()
	checkErr(err, "Create tables failed")

	return dbmap
}

func checkErr(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}
