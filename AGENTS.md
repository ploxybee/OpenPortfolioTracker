# Working agreement

## Repository layout

- `backend/`: Go 1.23 service. Keep application code under `internal/` and entrypoints under `cmd/`.
- `frontend/`: Next.js dashboard using TypeScript.
- `docs/`: durable product and architectural decisions.

## Principles

1. This is a read-only portfolio tracker. Do not add trading or order-management capabilities without explicit approval.
2. Broker credentials are server-side secrets. Never expose or commit them, including in fixtures, logs, test snapshots, or `NEXT_PUBLIC_*` environment variables.
3. Keep broker SDK types inside adapter code. Return the neutral `portfolio.Snapshot` contract from providers.
4. Preserve money currency metadata. Do not aggregate across currencies without an explicit FX conversion rule and timestamp.
5. Prefer small, testable Go packages and accessible, responsive UI.

## Quality gates

Before handing off a change, run:

```bash
make format
make test
```

Update `README.md` for setup or user-visible changes, and `docs/architecture.md` when interfaces, boundaries, or data flow change.
