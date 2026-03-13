# Skin Trading Bot Prototype

A prototype bot that acts as a financial helper for your skins.

## Status
Work in progress - backend is done, frontend is partially AI-generated.

## What it does
- Crunches market numbers
- Sends signals when prices are super low
- Tells you when it's a good time to sell

## Tech Stack
- Backend: custom account system (prototype)
- Frontend: 50% AI-generated, 50% handwritten (I'm still learning JS)

## Limitations
- Not a final product yet due to gaps on the dev side
- Frontend is not fully functional

## Notes
- Backend written entirely by me
- Frontend partially generated with AI

## Backend
- Env vars: `DATABASE_URL`, `ACCESS_TOKEN_SECRET`, `COOKIE_SECURE`, `TOKEN`
- `/get/` endpoint is a proxy to Steam and requires external network access
- Auth is prototype-level: passwords are stored as plain text in the database

## Tests
- `test/terminal_test.go` uses a local mock server, so it does not require Steam access
