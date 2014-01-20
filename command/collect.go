package command

import (
	"flag"
	"fmt"
	"strings"
	"time"

	ocd "github.com/lox/opencoindata/core"
	"github.com/mitchellh/cli"
)

// CollectCommand polls exchanges to collect data from them
type CollectCommand struct {
	ShutdownCh <-chan struct{}
	Ui         cli.Ui
}

func (c *CollectCommand) Help() string {
	helpText := `
Usage: opencoindata collect [options]

  Polls supported exchanges to collect new data from them to be persisted locally.

  NOTE: This command will run until terminated with ctrl-c

Options:

  -i=30s                    The interval to use for polling
`
	return strings.TrimSpace(helpText)
}

func (c *CollectCommand) Run(args []string) int {
	var durationArg string

	cmdFlags := flag.NewFlagSet("start", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.Ui.Output(c.Help()) }
	cmdFlags.StringVar(&durationArg, "i", "30s", "The time between polls")

	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	duration, err := time.ParseDuration(durationArg)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error parsing duration: %s", err))
		return 1
	}

	// start up collectors
	go func() {
		for tc := range ocd.TradeCollectors([]string{"btce"}) {
			c.Ui.Output(fmt.Sprintf("Polling %v on %s", tc.Pair, tc.ExchangeKey))
			go func() {
				for trade := range tc.Collect(duration) {
					c.Ui.Output(fmt.Sprintf("%s", trade.String()))
				}
			}()
		}
	}()
	<-c.ShutdownCh

	return 1
}

func (c *CollectCommand) Synopsis() string {
	return "Collect data from exchanges"
}
