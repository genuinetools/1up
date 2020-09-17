# 1up

[![make-all](https://github.com/genuinetools/1up/workflows/make%20all/badge.svg)](https://github.com/genuinetools/1up/actions?query=workflow%3A%22make+all%22)
[![make-image](https://github.com/genuinetools/1up/workflows/make%20image/badge.svg)](https://github.com/genuinetools/1up/actions?query=workflow%3A%22make+image%22)
[![GoDoc](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=for-the-badge)](https://godoc.org/github.com/genuinetools/1up)
[![Github All Releases](https://img.shields.io/github/downloads/genuinetools/1up/total.svg?style=for-the-badge)](https://github.com/genuinetools/1up/releases)

A custom Gmail spam filter bot.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**

- [How it works](#how-it-works)
- [Installation](#installation)
    - [Binaries](#binaries)
    - [Via Go](#via-go)
    - [Via Docker](#via-docker)
- [Usage](#usage)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

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
