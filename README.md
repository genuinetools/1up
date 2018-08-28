# 1up

[![Travis CI](https://img.shields.io/travis/genuinetools/1up.svg?style=for-the-badge)](https://travis-ci.org/genuinetools/1up)
[![GoDoc](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=for-the-badge)](https://godoc.org/github.com/genuinetools/1up)
[![Github All Releases](https://img.shields.io/github/downloads/genuinetools/1up/total.svg?style=for-the-badge)](https://github.com/genuinetools/1up/releases)

A custom Gmail spam filter bot.

* [How it works](README.md#how-it-works)
* [Installation](README.md#installation)
   * [Binaries](README.md#binaries)
   * [Via Go](README.md#via-go)
   * [Via Docker](README.md#via-docker)
* [Usage](README.md#usage)

## How it works


The bot will create 3 labels in your Gmail:

- `1up/good`: where you label emails that are "good"
- `1up/bad`: where you label emails that are "bad"
- `1up/quarantine`: where the bot will place emails that it thinks are "bad"
    based off the results of the Bayes classifier

Thanks to [@brendandburns](https://github.com/brendandburns) for pointing me at
Bayes classifiers.

## Installation

#### Binaries

For installation instructions from binaries please visit the [Releases Page](https://github.com/genuinetools/1up/releases).

You will want to follow the steps [here](https://developers.google.com/gmail/api/quickstart/go#step_1_turn_on_the) to turn on the Gmail API and get a credentials file.

#### Via Go

```console
$ go get github.com/genuinetools/1up
```

#### Via Docker

```console
$ docker run --rm -it -v ~/configs/1up:/1up:ro \
    --tmpfs /tmp \
    r.j3ss.co/1up -f /1up/credentials.json
```

## Usage

```console
$ 1up -h
1up -  A custom Gmail spam filter bot.

Usage: 1up <command>

Flags:

  -d, --debug       enable debug logging (default: false)
  -f, --creds-file  Gmail credential file (or env var GMAIL_CREDENTIAL_FILE) (default: <none>)
  -i, --interval    update interval (ex. 5ms, 10s, 1m, 3h) (default: 5m0s)
  --once            run once and exit, do not run as a daemon (default: false)

Commands:

  version  Show the version information.
```
