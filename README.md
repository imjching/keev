## HashDB

Work-in-progress.
A simple key-value store backed by hash tables

## Architecture

High-level overview:

                 ┌ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┐
                                      Clients
                 └ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┘
                                        ▲
                                        │
                                        │ Google Protocol Buffers 
                                        │
                                        ▼
                 ┌───────────────────────────────────────────────┐
                 │                    gRPC                       │
                 └───────────────────────────────────────────────┘
                 ┌───────────────────────────────────────────────┐
                 │                 RAM or disk                   │
                 └───────────────────────────────────────────────┘

Database current accepts the following commands:
- SET key value (valid if key is not present)
- UPDATE key value (valid if key is present)
- HAS key
- UNSET key
- GET key
- COUNT
- SHOW KEYS
- SHOW DATA
- SHOW NAMESPACES
- USE namespace

Restrictions:
* Both `key` and `value` cannot contain spaces.
* `key` cannot contain dots.
