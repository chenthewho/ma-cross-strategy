# AI Agent Instructions

You are working on **ma-cross-strategy**, a cryptocurrency quantitative
trading backtester being migrated from Python to Go.

## Always

- Read CLAUDE.md first for project context and coding standards.
- All public functions must have godoc comments.
- Use `float64` for all financial calculations.
- Every HTTP call takes `context.Context` for timeout/cancellation.
- Errors propagate up — never silently ignored.
- Write table-driven tests for every new package.
- Keep the Python `src/` code as reference — do not delete it yet.

## Never

- Do not use `panic` outside of `init()` or truly unrecoverable states.
- Do not hardcode API keys or secrets.
- Do not use `float32` in financial math.
- Do not skip error handling with `_`.

## Communication

- Reply in Chinese unless the user switches to English.
- Be concise — no fluff, just facts and code.
