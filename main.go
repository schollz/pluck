package main

import (
	"fmt"
	"io/ioutil"
	"os"
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
$ pluck -c config.toml -u http://www.foodnetwork.com/recipes/food-network-kitchen/15-minute-shrimp-tacos-with-spicy-chipotle-slaw-3676441
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
			Usage: "text to find to stop capture",
		},
		cli.IntFlag{
			Name:  "limit,l",
			Value: 1,
			Usage: "maximum number of items to capture",
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
			fmt.Println("Must specify file or url")
			return nil
		}
		p, _ := pluck.New()
		if len(c.GlobalString("config")) > 0 {
			p.Load(c.GlobalString("config"))
		} else {
			if len(c.GlobalStringSlice("activator")) == 0 {
				fmt.Println("Must specify at least one activator")
				return nil
			}
			if len(c.GlobalString("deactivator")) == 0 {
				fmt.Println("Must specify at deactivator")
				return nil
			}
			p.Add(pluck.Config{
				Activators:  c.GlobalStringSlice("activator"),
				Deactivator: c.GlobalString("deactivator"),
				Limit:       c.GlobalInt("limit"),
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
		if c.GlobalString("output") != "" {
			return ioutil.WriteFile(c.GlobalString("output"), []byte(p.ResultJSON()), 0644)
		} else {
			fmt.Println(p.ResultJSON())
		}

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Print(err)
	}
}
