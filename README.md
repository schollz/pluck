<p align="center">
<img
    src="logo.png"
    width="260" height="80" border="0" alt="pluck">
<br>
<a href="https://travis-ci.org/schollz/pluck"><img src="https://img.shields.io/travis/schollz/pluck.svg?style=flat-square" alt="Build Status"></a>
<a href="https://github.com/schollz/pluck/releases/latest"><img src="https://img.shields.io/badge/version-1.1.1-brightgreen.svg?style=flat-square" alt="Version"></a>
<img src="https://img.shields.io/badge/coverage-51%25-yellow.svg?style=flat-square" alt="Code Coverage">
</p>

<p align="center">Pluck text from a file.</p>

*pluck* grabs text from a file. Tell it what you are looking for, and where it should begin and end, and it will pluck it. I made this to parse HTML without having to think about XPATHs, but *pluck* is agnostic to type of file (HTML / XML / plain). It is naive, yet fast.

Demo
====

[![asciicast](demo.gif)](https://asciinema.org/a/Oq6enXjipBXqFcugqV7mSvdpR)

Getting Started
===============

## Install

If you have Go1.7+

```
go get github.com/schollz/pluck
```

or just download from the [latest releases](https://github.com/schollz/pluck/releases/latest).

## Basic usage 

```bash
$ pluck -a '<option class="level-0','>' -d '</option' -l 3 file.txt 
{
  "0": [
    "2012",
    "2013",
    "2014"
  ]
}
```

The `-a` specifies *activators*. Once all *activators* are found, in order, the bytes are captured. The `-d` specifies a *deactivator*. Once a *deactivator* is found, then it terminates capturing and reset search. The `-l` specifies the limit (optional), after reaching this limit it stops searching.

Advanced Usage
==============

Use a config file define complicated things to pluck:

```yaml
- 
  activators: 
    - "<option class=\"level-0\" "
    - ">"
  deactivator: <
  limit: 10
  name: options
- 
  activators: 
    - <title>
  deactivator: </title>
  name: title
- 
  activators: 
    - ">Song of the Day: "
  deactivator: <
  limit: -1
  name: songs
```

The `activators` will set the start position, after each one as been found, in order. The `deactivator` will elicit a stop once found and save the corresponding captured bytes. The `limit` (optional) specifies how many times it will pluck if conditions suffice. The `name` (optional) will specify the key in the JSON returned.

Then run 

```bash
$ pluck -c config.yaml test.txt
```

Features
========

This tool was inspired by the following:

- [lxml](https://github.com/warner/magic-wormhole)

*pluck* does not represent a significant innovation over these tools. However, there are some advantages that *pluck* provides:

- Trust. You can run your own *cowyo* server on a domain you trust.
- Direct edting. You can directly edit plaintext documents on the *cowyo* server using the web interface.
- Simplicity. The codebase is < 1k LOC, and is straightforward to understand.

## Benchmark

The [state of the art is `lxml`, based on libxml2](http://lxml.de/performance.html). Here is a comparision for plucking the same data from the same file, run on Windows i7-3770 CPU @ 3.40GHz.

| Language  | Time (ms) |
| ------------- | ------------- |
| Python3.6 `lxml`  | 3.8  |
| Golang `pluck`  | 0.8  |

Development
===========

```
$ go get -u github.com/schollz/pluck
$ cd $GOPATH/src/github.com/schollz/pluck/internal
$ go test -cover
```

License
========

MIT




