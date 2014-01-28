package command

import (
	"flag"
	"strings"

	"github.com/lox/opencoindata/web"
	"github.com/mitchellh/cli"
)

// WebCommand loads up the web application server
type WebCommand struct {
	Ui cli.Ui
}

func (c *WebCommand) Help() string {
	helpText := `
Usage: opencoindata web [options]

  Launches a web server running the opencoindata site

Options:

  -b=address:port               The interface to bind to, defaults to :8080
`
	return strings.TrimSpace(helpText)
}

func (c *WebCommand) Run(args []string) int {
	var bindArg string

	cmdFlags := flag.NewFlagSet("web", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.Ui.Output(c.Help()) }
	cmdFlags.StringVar(&bindArg, "b", ":8080", "bind")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	if err := web.NewWebServer().Serve(bindArg); err != nil {
		c.Ui.Error(err.Error())
		return 1
	}

	return 0
}

func (c *WebCommand) Synopsis() string {
	return "Collect data from exchanges"
}
