# pdt

CLI for the [Portfolio Dividend Tracker](https://portfoliodividendtracker.com) API. Designed for AI agents and automation — all output is JSON, no interactive prompts.

## Install

Download the latest binary for your platform from [Releases](https://github.com/PortfolioDividendTracker/cli/releases).

### macOS (Apple Silicon)

```bash
curl -L https://github.com/PortfolioDividendTracker/cli/releases/latest/download/pdt_darwin_arm64.tar.gz | tar xz
sudo mv pdt /usr/local/bin/
```

### macOS (Intel)

```bash
curl -L https://github.com/PortfolioDividendTracker/cli/releases/latest/download/pdt_darwin_amd64.tar.gz | tar xz
sudo mv pdt /usr/local/bin/
```

### Linux

```bash
curl -L https://github.com/PortfolioDividendTracker/cli/releases/latest/download/pdt_linux_amd64.tar.gz | tar xz
sudo mv pdt /usr/local/bin/
```

## Setup

```bash
# Set your Personal Access Token
pdt config set token pat_your_token_here

# Or use environment variable
export PDT_TOKEN=pat_your_token_here
```

## Usage

```bash
# List your portfolios
pdt list-portfolios

# Get portfolio summary
pdt get-portfolio

# List bookings with pagination
pdt list-bookings --page=1 --perPage=50

# Get a specific dividend
pdt get-dividend --dividendId=123

# Create a booking
pdt create-booking --brokerId=1 --date=2026-01-15 --amount=1000

# Search symbols
pdt search-symbols --query=AAPL

# Update the OpenAPI spec cache
pdt update
```

All commands output JSON to stdout. Errors go to stderr with exit code 1.

## Authentication

Three ways to provide your token (highest priority first):

1. `--token` flag: `pdt list-bookings --token=pat_xxx`
2. `PDT_TOKEN` environment variable
3. Config file: `pdt config set token pat_xxx`

## Configuration

```bash
pdt config set url https://api.portfoliodividendtracker.com/v1
pdt config set token pat_your_token
pdt config get url
pdt config get token
```

Config is stored at `~/.pdt/config.json`.

## Available Commands

Commands are generated dynamically from the API's OpenAPI specification. Run `pdt --help` to see all available commands, or `pdt <command> --help` for details on a specific command.
