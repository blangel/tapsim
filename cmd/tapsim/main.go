package main

import (
	"fmt"
	"log"
	"os"

	"github.com/halseth/tapsim/script"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "tapsim",
		Usage: "parse and debug bitcoin scripts",
	}

	app.Commands = []*cli.Command{
		{
			Name:        "parse",
			Usage:       "",
			UsageText:   "",
			Description: "",
			ArgsUsage:   "",
			Action:      parse,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "script",
					Usage: "script to parse",
				},
			},
		},
		{
			Name:        "execute",
			Usage:       "",
			UsageText:   "",
			Description: "",
			ArgsUsage:   "",
			Action:      execute,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "script",
					Usage: "output script",
				},
				&cli.StringFlag{
					Name:  "witness",
					Usage: "witness stack",
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func parse(cCtx *cli.Context) error {

	var scriptStr string
	if cCtx.NArg() > 0 {
		scriptStr = cCtx.Args().Get(0)
	} else if cCtx.String("script") != "" {
		scriptStr = cCtx.String("script")
	}

	parsed, err := script.Parse(scriptStr)
	if err != nil {
		return err
	}

	fmt.Printf("Parsed: %x\n", parsed)
	return nil
}

func execute(cCtx *cli.Context) error {
	var scriptStr, witnessStr string
	if cCtx.NArg() > 0 {
		scriptStr = cCtx.Args().Get(0)
	} else if cCtx.String("script") != "" {
		scriptStr = cCtx.String("script")
	}

	if cCtx.NArg() > 1 {
		witnessStr = cCtx.Args().Get(1)
	} else if cCtx.String("witness") != "" {
		witnessStr = cCtx.String("witness")
	}

	fmt.Printf("Script: %s\r\n", scriptStr)
	fmt.Printf("Witness: %s\r\n", witnessStr)

	parsedScript, err := script.Parse(scriptStr)
	if err != nil {
		return err
	}

	parsedWitness, err := script.Parse(witnessStr)
	if err != nil {
		return err
	}

	executeErr := script.Execute(parsedScript, parsedWitness)
	if executeErr != nil {
		fmt.Printf("script exection failed: %s\n", executeErr)
		return nil
	}

	fmt.Printf("script verified\r\n")
	return nil
}