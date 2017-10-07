<p align="center">
<img
    src="pluck/test/logo.png"
    width="260" height="80" border="0" alt="pluck">
<br>
<a href="https://github.com/schollz/pluck/releases/latest"><img src="https://img.shields.io/badge/version-1.0.0-brightgreen.svg?style=flat-square" alt="Version"></a>
<img src="https://img.shields.io/badge/coverage-92%25-green.svg?style=flat-square" alt="Code Coverage">
<a href="https://godoc.org/github.com/schollz/pluck/pluck"><img src="https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square" alt="Code Coverage"></a>
</p>

<p align="center">Pluck text in a fast and intuitive way. :rooster:</p>

*pluck* makes text extraction intuitive and [fast](https://github.com/schollz/pluck#current-benchmark). You can specify an extraction in nearly the same way you'd tell a person trying to extract the text by hand: "OK Bob, every time you find *X* and then *Y*, copy down everything you see until you encounter *Z*." 

In *pluck*, *X* and *Y* are called *activators* and *Z* is called the *deactivator*. The file/URL being plucked is parsed (or streamed) byte-by-byte into a finite state machine. Once all *activators* are found, the following bytes are saved to a buffer, which is added to a list of results once the *deactivator* is found. Multiple queries are extracted simultaneously and there is no requirement on the file format (e.g. XML/HTML), as long as its text.


# Why?

*pluck* was made as a simple alternative to xpath and regexp. Through simple declarations, *pluck* allows complex procedures like [extracting text in nested HTML tags](https://github.com/schollz/pluck#use-config-file), or [extracting the content of an attribute of a HTML tag](https://github.com/schollz/pluck#basic-usage). *pluck* may not work in all scenarios, so do not consider it a replacement for xpath or regexp.

### Doesn't regex already do this?

Yes basically. Here is [an (simple) example](https://regex101.com/r/xt7fVr/1):

```
(?:(?:X.*Y)|(?:Y.*X))(.*)(?:Z)
```

Basically, this should try and match everything before a `Z` and after we've seen both `X` and `Y`, in any order. This is not a complete example, but it shows the similarity.

The benefit with *pluck* is simplicity. You don't have to worry about escaping the right characters, nor do you need to know any regex syntax (which is not simple). Also *pluck* is hard-coded for matching this specific kind of pattern simultaneously, so there is no cost for generating a new deterministic finite automaton from multiple regex.

### Doesn't cascadia already do this?

Yes, there is already [a command-line tool](https://github.com/suntong/cascadia) to extract structured information from XML/HTML. There are many benefits to *cascadia*, namely you can do a lot more complex things with structured data. If you don't have highly structured data, *pluck* is advantageous (it extracts from any file). Also, with *pluck* you don't need to learn CSS selection.

# Getting Started

## Install

If you have Go1.7+

```
go get github.com/schollz/pluck
```

or just download from the [latest releases](https://github.com/schollz/pluck/releases/latest).

## Basic usage 

Lets say you want to find URLs in a HTML file.

```bash
$ wget nytimes.com -O nytimes.html
$ pluck -a '<' -a 'href' -a '"' -d '"' -l 10 -f nytimes.html
{
    "0": [
        "https://static01.nyt.com/favicon.ico",
        "https://static01.nyt.com/images/icons/ios-ipad-144x144.png",
        "https://static01.nyt.com/images/icons/ios-iphone-114x144.png",
        "https://static01.nyt.com/images/icons/ios-default-homescreen-57x57.png",
        "https://www.nytimes.com",
        "http://www.nytimes.com/services/xml/rss/nyt/HomePage.xml",
        "http://mobile.nytimes.com",
        "http://mobile.nytimes.com",
        "https://typeface.nyt.com/css/zam5nzz.css",
        "https://a1.nyt.com/assets/homepage/20170731-135831/css/homepage/styles.css"
    ]
}
```

The `-a` specifies *activators* and can be specified multiple times. Once all *activators* are found, in order, the bytes are captured. The `-d` specifies a *deactivator*. Once a *deactivator* is found, then it terminates capturing and resets and begins searching again. The `-l` specifies the limit (optional), after reaching the limit (`10` in this example) it stops searching.


## Advanced usage

### Parse URLs or Files

Files can be parsed with `-f FILE` and URLs can be parsed by instead using `-u URL`.

```bash
$ pluck -a '<' -a 'href' -a '"' -d '"' -l 10 -u https://nytimes.com
```

### Use Config file

You can also specify multiple things to pluck, simultaneously, by listing the *activators* and the *deactivator* in a TOML file. For example, lets say we want to parse ingredients and the title of [a recipe](https://goo.gl/DHmqmv). Make a file `config.toml`:

```toml
[[pluck]]
name = "title"
activators = ["<title>"]
deactivator = "</title>"

[[pluck]]
name = "ingredients"
activators = ["<label","Ingredient",">"]
deactivator = "<"
limit = -1
```

The title follows normal HTML and the ingredients were determined by quickly inspecting the HTML source code of the target site. Then, pluck it with,

```bash
$ pluck -c config.toml -u https://goo.gl/DHmqmv
{
    "ingredients": [
        "1 pound medium (26/30) peeled and deveined shrimp, tails removed",
        "2 teaspoons chili powder",
        "Kosher salt",
        "2 tablespoons canola oil",
        "4 scallions, thinly sliced",
        "One 15-ounce can black beans, drained and rinsed well",
        "1/3 cup prepared chipotle mayonnaise ",
        "2 limes, 1 zested and juiced and 1 cut into wedges ",
        "One 14-ounce bag store-bought coleslaw mix (about 6 cups)",
        "1 bunch fresh cilantro, leaves and soft stems roughly chopped",
        "Sour cream or Mexican crema, for serving",
        "8 corn tortillas, warmed "
    ],
    "title": "15-Minute Shrimp Tacos with Spicy Chipotle Slaw Recipe | Food Network Kitchen | Food Network"
}
```

### Extract structured data

Lets say you want to tell Bob "OK Bob, first look for *W*. Then, every time you find *X* and then *Y*, copy down everything you see until you encounter *Z*. Also, stop if you see *U*, even if you are not at the end."  In this case, *W*, *X*, and *Y* are activators but *W* is a "Permanent" activator. Once *W* is found, Bob forgets about looking for it anymore. *U* is a "Finisher" which tells Bob to stop looking for anything and return whatever result was found. 

You can extract information from blocks in *pluck* by using these two keywords: "*permanent*" and "*finisher*". The *permanent* number determines how many of the activators (from the left to right) will stay activated forever, once activated. The *finisher* keyword is a new string that will retire the current plucker when found and not capture anything in the buffer.

For example, suppose you want to only extract `link3` and `link4` from the following: 

```html
<h1>Section 1</h1>
<a href="link1">1</a>
<a href="link2">2</a>
<h1>Section 2</h1>
<a href="link3">3</a>
<a href="link4">4</a>
<h1>Section 3</h1>
<a href="link5">5</a>
<a href="link6">6</a>
```

You can add "Section 2" as an activator and set permanent to `1` so that only the first activator ("Section 2") will continue to remain activated after finding the deactivator. Then you want to finish the plucker when it hits "Section 3", so we can set the finisher keyword as this. Then `config.toml` is

```
[[pluck]]
activators = ["Section 2","a","href",'"']
permanent = 1     # designates that the first 1 activators will persist
deactivator = '"'
finisher = "Section 3"
```

will result in the following:

```json
{
    "0": [
        "link3",
        "link4",
    ]
}
```


### More examples

See [EXAMPLES.md](https://github.com/schollz/pluck/blob/master/EXAMPLES.md) for more examples.

### Use as a Go package

Import pluck as `"github.com/schollz/pluck/pluck"` and you can use it in your own project. See the tests for more info.



# Development

```
$ go get -u github.com/schollz/pluck/...
$ cd $GOPATH/src/github.com/schollz/pluck/pluck
$ go test -cover
```

## Current benchmark

The [state of the art for xpath is `lxml`, based on libxml2](http://lxml.de/performance.html). Here is a comparison for plucking the same data from the same file, run on Intel i5-4310U CPU @ 2.00GHz × 4. (Run Python benchmark `cd pluck/test && python3 main.py`).

| Language  | Rate |
| ------------- | ------------- |
| `lxml` (Python3.5)  | 300 / s  |
| pluck | 1270 / s |

A real-world example I use *pluck* for is processing 1,200 HTML files in parallel, compared to running `lxml` in parallel:

| Language  | Rate |
| ------------- | ------------- |
| `lxml` (Python3.6)  | 25 / s  |
| pluck | 430 / s |

I'd like to benchmark a Perl regex, although I don't know how to write this kind of regex! Send a PR if you do :)

## To Do

- [ ] Allow OR statements (e.g `'|"`). 
- [ ] Quotes match to quotes (single or double)?
- [ ] Allow piping from standard in?
- [x] API to handle strings, e.g. `PluckString(s string)`
- [x] Add parallelism

# License

MIT

# Acknowledgements

<a target="_blank" href="https://www.vecteezy.com">Graphics by: www.vecteezy.com</a>
