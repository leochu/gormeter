package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/leochu/gormeter/summary/commands"
)

var flags = []cli.Flag{
	cli.StringFlag{
		Name:  "path",
		Usage: "Path of log files (required)",
	},
	cli.StringFlag{
		Name:  "out",
		Usage: "Path of out put directory",
	},
	cli.BoolFlag{
		Name:  "xml",
		Usage: "Specifies the xml format for log files",
	},
}

var perfFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "path",
		Usage: "Path of log files",
	},
	cli.StringFlag{
		Name:  "httpPath",
		Usage: "Path to summary file of http",
	},
	cli.StringFlag{
		Name:  "httpsPath",
		Usage: "Path to summary file of https",
	},
}

var cliCommands = []cli.Command{
	{
		Name:   "generate",
		Usage:  "generates summary",
		Action: commands.GenerateSummary,
		Flags:  flags,
	},
	{
		Name:   "analysis",
		Usage:  "Performs analysis on generated summary",
		Action: commands.PerformAnalysis,
		Flags:  perfFlags,
	},
}

func main() {
	fmt.Println()
	app := cli.NewApp()
	app.Name = "jmeter summary"
	app.Commands = cliCommands
	app.CommandNotFound = commandNotFound
	app.Version = "0.1.0"

	app.Run(os.Args)
	os.Exit(0)
}

func commandNotFound(c *cli.Context, cmd string) {
	fmt.Println("Not a valid command:", cmd)
	os.Exit(1)
}
