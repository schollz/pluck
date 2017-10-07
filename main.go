package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/schollz/pluck/pluck"
	"github.com/urfave/cli"
)

var version string

func main() {
	app := cli.NewApp()
	app.Version = version
	app.Compiled = time.Now()
	app.Name = "pluck"
	app.Usage = ""
	app.UsageText = `	
1) Pluck all URLs from a website
$ pluck -a '<' -a 'href' -a '"' -d '"' -l -1 -u https://nytimes.com

2) Pluck title from a HTML file
$ pluck -a '<title>' -d '<' -f test.html

3) Pluck using a configuration file. 
$ # Example config file
$ cat config.toml
[[pluck]]
activators = ["<title>"]
deactivator = "</title>"
name = "title"

[[pluck]]
activators = ["<label","Ingredient",">"]
deactivator = "<"
limit = -1
name = "ingredients"
$ pluck -c config.toml -u https://goo.gl/DHmqmv

4) Get headlines from news.google.com 
$ pluck -a 'role="heading"' -a '>' -d '<' -t -s -u 'https://news.google.com/news/?ned=us&hl=en'

5) Pluck items from a block
$ pluck -a 'Section 2' -a '<a' -a 'href' -a '"' -d '"' -p 1 -finisher "Section 3" -u https://cowyo.com/test38/raw
		`
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "file,f",
			Value: "",
			Usage: "file to pluck",
		},
		cli.StringFlag{
			Name:  "url,u",
			Value: "",
			Usage: "url to pluck",
		},
		cli.StringFlag{
			Name:  "config,c",
			Value: "",
			Usage: "specify toml config file",
		},
		cli.StringSliceFlag{
			Name:  "activator,a",
			Usage: "text to find in order to start capture (can specify multiple times)",
		},
		cli.StringFlag{
			Name:  "deactivator,d",
			Value: "",
			Usage: "text to find to restart capturing",
		},
		cli.IntFlag{
			Name:  "permanent,p",
			Value: 0,
			Usage: "number of activators that stay activated (from left to right)",
		},
		cli.StringFlag{
			Name:  "finisher",
			Value: "",
			Usage: "text to find to stop capturing completely",
		},
		cli.IntFlag{
			Name:  "limit,l",
			Value: -1,
			Usage: "maximum number of items to capture",
		},
		cli.BoolFlag{
			Name:  "sanitize,s",
			Usage: "sanitize output (html tag stripping and hex conversion)",
		},
		cli.BoolFlag{
			Name:  "text, t",
			Usage: "output as plain text, not JSON",
		},
		cli.BoolFlag{
			Name:  "verbose",
			Usage: "turn on verbose mode",
		},
		cli.StringFlag{
			Name:  "output,o",
			Value: "",
			Usage: "direct output to file",
		},
	}

	app.Action = func(c *cli.Context) (err error) {
		if c.GlobalString("file") == "" && c.GlobalString("url") == "" {
			fmt.Println("Must specify file or url. For example -u https://nytimes.com.\nSee help and usage with -h")
			return nil
		}
		p, _ := pluck.New()
		if c.GlobalBool("verbose") {
			p.Verbose(true)
		}
		if len(c.GlobalString("config")) > 0 {
			p.Load(c.GlobalString("config"))
		} else {
			if len(c.GlobalStringSlice("activator")) == 0 {
				fmt.Println("Must specify at least one activator. For example -a 'start'.\nSee help and usage with -h")
				return nil
			}
			if len(c.GlobalString("deactivator")) == 0 {
				fmt.Println("Must specify at deactivator. For example -d 'end'.\nSee help and usage with -h")
				return nil
			}
			p.Add(pluck.Config{
				Activators:  c.GlobalStringSlice("activator"),
				Deactivator: c.GlobalString("deactivator"),
				Limit:       c.GlobalInt("limit"),
				Sanitize:    c.GlobalBool("sanitize"),
				Finisher:    c.GlobalString("finisher"),
				Permanent:   c.GlobalInt("permanent"),
			})
		}

		if len(c.GlobalString("file")) > 0 {
			err = p.PluckFile(c.GlobalString("file"))
		} else {
			err = p.PluckURL(c.GlobalString("url"))
		}
		if err != nil {
			return err
		}
		var result string
		if c.GlobalBool("text") {
			results, ok := p.Result()["0"].([]string)
			if !ok {
				results2, ok2 := p.Result()["0"].(string)
				if !ok2 {
					fmt.Println("Error?")
					os.Exit(-1)
				} else {
					result = results2
				}
			} else {
				result = strings.Join(results, "\n\n")
			}
		} else {
			result = p.ResultJSON(true)
		}
		if c.GlobalString("output") != "" {
			return ioutil.WriteFile(c.GlobalString("output"), []byte(result), 0644)
		} else {
			fmt.Println(result)
		}

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Print(err)
	}
}
