# Trading Bot for XTB WebSocket API

[![CI](https://github.com/artufi/trader/actions/workflows/ci.yaml/badge.svg)](https://github.com/artufi/trader/actions/workflows/ci.yaml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/artufi/trader)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A real-time trading bot written in Go for the XTB WebSocket API. It connects to XTB over
WebSocket for market data and trade execution, manages a pool of concurrent trading clients
per user, and executes trade signals from an external ML model as buy, sell, or no-action trades.

> **Note:** The XTB API is no longer publicly available, so the bot cannot run end-to-end: it
> depends on that API and the connection pool fails to start without it. This project is
> open-sourced as a reference for backend architecture, concurrency, and real-time data
> handling in Go.

## Features

- Real-time communication with the XTB API over WebSocket, across both the request/response and
  streaming endpoints.
- A managed pool of concurrent, per-user trading clients with automatic reconnection.
- Trade signals from an external ML model, consumed over REST and executed as buy, sell, or
  no-action trades.
- Internal tracking of trades with interval-based position closing.
- Structured, context-aware logging with per-request trace IDs.
- Graceful shutdown and PostgreSQL-backed persistence.

## Architecture

The codebase is organised into clear layers:

| Package | Responsibility |
| --- | --- |
| `config` | Application configuration loading (dotenv / YAML). |
| `infrastructure/database` | PostgreSQL connection pool and Goose migrations. |
| `infrastructure/websocket` | WebSocket client and connection lifecycle, pooling, reconnection. |
| `xtb` | XTB protocol: commands, request/response models, processors. |
| `http` | REST controllers, middleware (trace IDs), response helpers. |
| `model` | Domain entities and persistence (users, predictions, orders). |
| `history` | Trade streaming and interval-based order closing. |
| `logging` | Custom `slog` handler that propagates attributes through `context`. |

### Connections

`ConnManager` opens a pool of WebSocket clients per user. Each connection runs in its own
goroutine that logs in, sends keep-alive pings, and reconnects automatically when the
connection drops.

Two client types map to the two XTB endpoints:

- **Request/response client** (`WSClient`) talks to the main endpoint. Each command carries a
  `customTag`, a `pending` map correlates the command with its single reply, and a per-client
  mutex guards the WebSocket's single writer.
- **Streaming client** (`WSClientStream`) talks to the streaming endpoint. It subscribes to a
  feed and receives pushed updates over a channel. It has no request/response correlation and
  authenticates with a stream session ID borrowed from a logged-in `WSClient`.

The pool exists so that every subscribed user keeps a few already-authenticated, long-lived
connections ready to act the moment a signal arrives. Logging in is costly and the API
rate-limits requests, so reusing warm connections lets signals execute in parallel with low
latency instead of opening a fresh session each time. A supervisor goroutine pings each
connection to keep it alive and transparently reconnects on failure, so the pool stays reliable
on its own. The bot also keeps long-lived listening connections open and reacts to pushed
updates, such as a position closing, rather than polling for them.

This connection layer is a reusable pattern: multiplexed request/response over a pooled,
auto-reconnecting WebSocket, applicable to any real-time API client (exchanges, chat, IoT,
message brokers).

### Signal flow

A REST request carrying a prediction resolves a connected XTB client, fetches symbol data,
places a trade transaction, and persists the prediction and order. In the background, one
goroutine streams trade updates into the database, while another closes eligible positions on
a fixed interval.

## Tech Stack

Go, PostgreSQL, Goose (migrations), `gorilla/websocket`, `chi` (routing), `pgx`, Docker Compose.

## Limitations

This is a prototype, not production-ready:

- A single hardcoded test user, no authentication or user management.
- Limited retry and recovery handling on edge cases.
- Partial test coverage, which is being expanded.
- The database schema needs indexing and type tuning for scale.

## Project background

The bot was developed and run on a Raspberry Pi 4. It handled the workload, though performance
was modest. Development was later paused, and the XTB API was eventually discontinued, at which
point the project was open-sourced as an architecture and design reference.

## License

[MIT](LICENSE)
