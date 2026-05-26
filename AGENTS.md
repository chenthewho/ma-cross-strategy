# Agent Instructions — ma-cross-strategy

You are an AI coding assistant working on a quantitative trading SaaS platform.

## Always

- Read `CLAUDE.md` first — it is the project constitution.
- All functionality is defined in `doc/` (system topology, strategy math engine, evolution engine). Do not implement anything not specified there.
- Strategy code in `internal/strategies/` must be pure functions — no I/O, no network, no database.
- `Step()` is the single source of truth for signal generation — shared by backtest and live trading.
- Backtest and live trading call the same `Step()` implementation.
- GORM Code-First: use `AutoMigrate` only, never hand-write DDL.
- Price calculations use dimensionless expressions (log returns, ratios) when possible.
- Respect SaaS-Strategy-Agent boundaries — do not preemptively decouple.

## Five Iron Rules

1. Strategies must pass compound-interest precondition check
2. Same `Step()` for backtest AND live
3. `Step()` only executes on SaaS side
4. Strategy packages: no I/O (network/db/file)
5. API keys only in `config.agent.yaml`

## Communication

- Reply in Chinese (中文) unless the user switches to English.
- Be concise. Code speaks louder than words.
- After each task: `git commit` → `git push` → WeChat notification.
