package command

import (
	"flag"
	"os"
	"strings"

	"github.com/lox/opencoindata/web"
	"github.com/mitchellh/cli"
)

const (
	DEFAULT_BIND = ":8080"
)

// WebCommand loads up the web application server
type WebCommand struct {
	Ui cli.Ui
}

func (c *WebCommand) Help() string {
	helpText := `
Usage: opencoindata web [options]

  Launches a web server running the opencoindata site
`
	return strings.TrimSpace(helpText)
}

func (c *WebCommand) Run(args []string) int {
	var bindArg string

	cmdFlags := flag.NewFlagSet("web", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.Ui.Output(c.Help()) }
	cmdFlags.StringVar(&bindArg, "b", DEFAULT_BIND, "bind")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	config := web.ServeConfig{}

	if os.Getenv("OCD_BIND") != "" {
		config.BindAddress = os.Getenv("OCD_BIND")
	}

	if os.Getenv("OCD_API_HOST") != "" {
		config.ApiHostname = os.Getenv("OCD_API_HOST")
	}

	if os.Getenv("OCD_WS_HOST") != "" {
		config.WsHostname = os.Getenv("OCD_WS_HOST")
	}

	if err := web.Serve(config); err != nil {
		c.Ui.Error(err.Error())
		return 1
	}

	return 0
}

func (c *WebCommand) Synopsis() string {
	return "Collect data from exchanges"
}
