// Demo card terminal — a WASI command (GOOS=wasip1 GOARCH=wasm) run
// in-process by the till's wazero runtime for every payment.demo.requested
// event. It simulates a certified tap-to-pay terminal with DETERMINISTIC
// outcomes so staff training, demos and E2E tests need no hardware.
//
// The "test cards" are encoded in the pence amount (there is no card-entry
// UI at the till — same idea as Stripe's magic test amounts):
//
//	amount ends .13  → DECLINED  (card declined by issuer)
//	amount ends .99  → TIMEOUT   (terminal never answers; the till's event
//	                              deadline kills the handler — the real-world
//	                              "reader lost connection" path)
//	anything else    → APPROVED  (fake auth code emitted)
//
// Every outcome is written to plugin storage via the `ut` host functions
// (permission "storage") — the first marketplace plugin exercising host
// functions end-to-end.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
	"unsafe"
)

//go:wasmimport ut log_write
func utLogWrite(ptr, n uint32)

//go:wasmimport ut storage_set
func utStorageSet(kPtr, kLen, vPtr, vLen uint32) int32

func ptrOf(b []byte) (uint32, uint32) {
	if len(b) == 0 {
		return 0, 0
	}
	return uint32(uintptr(unsafe.Pointer(&b[0]))), uint32(len(b))
}

func logf(format string, args ...any) {
	msg := []byte(fmt.Sprintf(format, args...))
	p, n := ptrOf(msg)
	utLogWrite(p, n)
}

func store(key string, val []byte) int32 {
	kp, kl := ptrOf([]byte(key))
	vp, vl := ptrOf(val)
	return utStorageSet(kp, kl, vp, vl)
}

type event struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Payload struct {
		SaleID    string `json:"sale_id"`
		Method    string `json:"method"`
		Amount    int64  `json:"amount"`
		Reference string `json:"reference"`
	} `json:"payload"`
}

func main() {
	raw, _ := io.ReadAll(os.Stdin)
	var ev event
	if err := json.Unmarshal(raw, &ev); err != nil {
		fmt.Fprintf(os.Stderr, "demo-card: bad event: %v\n", err)
		os.Exit(1)
	}

	amount := ev.Payload.Amount

	// payment.demo.authorize — BLOCKING, runs BEFORE the sale completes.
	// Exit 0 = approved (tender proceeds); non-zero = declined (the till
	// refuses the sale). The timeout card never answers, so the runtime's
	// deadline kills us — the dead-reader path.
	if ev.Type == "payment.demo.authorize" {
		switch amount % 100 {
		case 13:
			logf("demo-card: DECLINED %d minor units", amount)
			result, _ := json.Marshal(map[string]any{"terminal": "demo", "amount": amount, "outcome": "declined"})
			_ = store("last_txn", result)
			os.Exit(2)
		case 99:
			logf("demo-card: simulating terminal timeout (%d minor units)", amount)
			time.Sleep(30 * time.Second)
			os.Exit(1)
		}
		auth := fmt.Sprintf("DEMO-%06d", amount%1000000)
		result, _ := json.Marshal(map[string]any{"terminal": "demo", "amount": amount, "outcome": "approved", "auth_code": auth})
		_ = store("last_txn", result)
		logf("demo-card: APPROVED %d minor units (%s)", amount, auth)
		_, _ = os.Stdout.Write(append(result, '\n'))
		return
	}

	// payment.demo.requested — post-settle: the sale exists now; file the
	// transaction under its id.
	result, _ := json.Marshal(map[string]any{
		"terminal": "demo", "sale_id": ev.Payload.SaleID,
		"amount": amount, "outcome": "settled",
	})
	if code := store("txn:"+ev.Payload.SaleID, result); code != 0 {
		logf("demo-card: storing settle failed (%d)", code)
	}
	logf("demo-card: settled sale %s (%d minor units)", ev.Payload.SaleID, amount)
	_, _ = os.Stdout.Write(append(result, '\n'))
}
