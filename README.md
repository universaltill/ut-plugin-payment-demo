# Demo Card Terminal (`com.universaltill.payment-demo`)

A simulated tap-to-pay terminal for **Universal Till**: train staff, run
demos and test the whole tender path — approvals, declines, dead-reader
timeouts — with **zero hardware**. Also the reference implementation for
real payment terminal plugins (SumUp, Stripe Terminal, …).

## Test cards

There is no card-entry UI at the till, so the "test card" is encoded in
the **pence amount**, the same trick as Stripe's magic test amounts:

| Amount ends | Outcome | What it simulates |
|---|---|---|
| `.13` (e.g. £5.13) | **DECLINED** | Issuer declines the card |
| `.99` (e.g. £5.99) | **TIMEOUT** | Reader lost connection / never answers |
| anything else | **APPROVED** | Fake auth code `DEMO-xxxxxx` issued |

Every outcome lands in the till's audit trail; approved transactions also
record a result in plugin storage (`txn:<sale_id>` + `last_txn`) via the
WASM host functions — this plugin is the first marketplace consumer of
them (`storage` permission).

## How it runs

`canonical_type: payment` — installing adds a **Demo Card** tender button.
Settling a sale with it publishes `payment.demo.requested`; the till's
in-process wazero runtime executes `bin/plugin.wasm` (WASI command, built
with plain Go `GOOS=wasip1`) with the event JSON on stdin.

## Develop

```bash
scripts/build.sh      # GOOS=wasip1 GOARCH=wasm go build → bin/plugin.wasm
scripts/validate.sh   # manifest sanity
scripts/package.sh    # dist/<id>_<version>_universal.tar.gz (+sha256)
```

Releases: tag `v<version>` (must match `manifest.json`) — CI packages,
publishes to the marketplace and (dev) auto-approves.
