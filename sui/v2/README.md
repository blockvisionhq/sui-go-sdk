# Sui Client V2 (Unified JSON-RPC / gRPC Contract)

`sui/v2` exposes one public client contract on top of two transport backends:

- JSON-RPC via `HttpConn`
- gRPC via `grpcconn.SuiGrpcClient`

The core value of v2 is that callers use the same request/response model regardless
of the underlying transport. The backend is selected once at client construction
time; the rest of the application code can stay unchanged.

## Design Goals

- Provide one stable public API for both JSON-RPC and gRPC.
- Reuse the same request option structs and response structs across transports.
- Make backend switching, fallback, and A/B comparison easy.
- Normalize transport-specific payload shape differences where practical.
- Keep remaining incompatibilities explicit instead of pretending the two backends
  are byte-for-byte identical.

## Backend Selection

`NewClient` accepts either a gRPC client or an HTTP JSON-RPC connection:

```go
package main

import (
    "context"

    "github.com/block-vision/sui-go-sdk/common/grpcconn"
    "github.com/block-vision/sui-go-sdk/common/httpconn"
    suiV2 "github.com/block-vision/sui-go-sdk/sui/v2"
)

func newGrpcClient() (suiV2.Client, error) {
    gc := grpcconn.NewSuiGrpcClient("your-grpc-endpoint")
    return suiV2.NewClient(suiV2.ClientOptions{
        GrpcClient: gc,
    })
}

func newJSONRPCClient() (suiV2.Client, error) {
    hc := httpconn.NewHttpConn("your-json-rpc-endpoint", nil)
    return suiV2.NewClient(suiV2.ClientOptions{
        HttpConn: hc,
    })
}

func main() {
    var client suiV2.Client
    var err error

    client, err = newGrpcClient()
    // client, err = newJSONRPCClient()
    if err != nil {
        panic(err)
    }

    _, _ = client.GetReferenceGasPrice(context.Background())
}
```

Switching transports should only require changing `ClientOptions`. Call sites keep
using the same `suiV2.Client` interface.

## Public Contract

The following parts of the v2 API are intended to be transport-agnostic:

- The `suiV2.Client` method set.
- Request option types such as `GetObjectsOptions`, `ListCoinsOptions`,
  `GetTransactionOptions`, and `SimulateTransactionOptions`.
- Shared response types re-exported from `sui/v2/types`.
- The meaning of `Include` flags for optional transaction/object fields.

In practice, this means callers can rely on the same top-level shapes for:

- Objects, owned objects, coins, balances, dynamic fields, system state, and Move
  metadata.
- Transaction responses via `TransactionResult` / `SimulateTransactionResult`.
- Parsed transaction sub-structures such as `TransactionData`,
  `TransactionEffects`, `ChangedObject`, `Event`, and `BalanceChange`.

v2 also normalizes several payload-shape differences for you:

- JSON-RPC and gRPC both map into the same exported Go structs.
- Transaction command keys are exposed in v2 naming (for example
  `typeArguments`, `result`, `subresult`).
- Coin/object type strings are normalized where the JSON-RPC source uses
  shorter or differently formatted type strings.
- Raw transaction bytes are exposed through `transaction.bcs` when requested and
  available from the selected backend.

## Example

```go
tx, err := client.GetTransaction(ctx, suiV2.GetTransactionOptions{
    Digest: "3CK7Fv9CUp3QetDhszToqwnDHgzpYPauP1H2N5iyuijh",
    Include: suiV2.TransactionInclude{
        Transaction:    true,
        Effects:        true,
        Events:         true,
        BalanceChanges: true,
        Bcs:            true,
    },
})
if err != nil {
    panic(err)
}

_ = tx.Transaction.TransactionData
_ = tx.Transaction.Effects
_ = tx.Transaction.Events
_ = tx.Transaction.BCS
```

## Typed Errors

v2 exposes a public typed error model through `sui/v2`:

- `*suiV2.SDKError`
- `suiV2.ErrBackendRequired`
- `suiV2.ErrGrpcClientRequired`
- `suiV2.ErrHttpConnRequired`
- `suiV2.ErrTransactionNotFound`
- `suiV2.ErrSystemStateNotFound`
- `suiV2.ErrChainIdentifierNotFound`

Use `errors.Is` for category checks and `errors.As` when you want extra context:

```go
resp, err := client.GetTransaction(ctx, suiV2.GetTransactionOptions{
    Digest: "some-digest",
})
if err != nil {
    if errors.Is(err, suiV2.ErrTransactionNotFound) {
        // handle not found
    }

    var sdkErr *suiV2.SDKError
    if errors.As(err, &sdkErr) {
        fmt.Printf("code=%s transport=%s method=%s operation=%s\n",
            sdkErr.Code, sdkErr.Transport, sdkErr.Method, sdkErr.Operation)
    }
    return
}

_ = resp
```

Transport-level failures are wrapped as `SDKError` with:

- `Code = transport_error`
- `Transport = grpc` or `json-rpc`
- `Method = public v2 method name`
- `Operation = underlying backend operation name`

This lets callers keep one error-handling path even when they switch transports.

## Running Examples

The repository includes `v2_examples/main.go`, which now supports both backends.

Run with gRPC:

```shell
V2_TRANSPORT=grpc \
BLOCKVISION_API_KEY=your_api_key \
go run ./v2_examples
```

Run with JSON-RPC:

```shell
V2_TRANSPORT=json-rpc \
go run ./v2_examples GetBalance
```

Optional environment variables:

- `V2_GRPC_ENDPOINT`
- `V2_RPC_URL`
- `BLOCKVISION_API_KEY` (required for `grpc`)

## Known Differences / Compatibility Notes

v2 targets semantic compatibility across transports, not guaranteed byte-for-byte
identity of every upstream payload. Consumers should rely on the normalized field
meaning, and treat the following differences as intentional compatibility notes:

- `GetChainIdentifier`: the two backends currently expose different upstream
  notions of chain identity. gRPC returns the genesis checkpoint digest from
  service info, while JSON-RPC returns the result of `sui_getChainIdentifier`.
  The value is surfaced through the same response field, but should not be
  assumed to be string-identical across transports.
- `transaction.bcs`: both backends expose raw transaction bytes at
  `transaction.bcs` when `include.bcs = true`, but they come from different
  upstream fields. gRPC uses transaction BCS from the protobuf response; JSON-RPC
  uses `rawTransaction`. Presence depends on whether the selected backend returns
  raw bytes for that request.
- `effects.eventsDigest`: gRPC can populate `effects.eventsDigest` directly.
  JSON-RPC does not currently provide an equivalent field in the normalized v2
  response, so callers must treat it as optional.
- `effects.changedObjects[].inputDigest`: gRPC can provide `inputDigest`
  directly for changed objects. JSON-RPC supplements it when it can derive the
  value from shared objects or gas payment inputs, but owned-object
  `inputDigest` is not always recoverable from JSON-RPC data and may be absent.
- `transaction.transaction.commands[].*.typeArguments`: v2 normalizes the field
  name itself to `typeArguments`, but the type argument strings may still differ
  in address formatting between backends, such as short addresses like `0x2::...`
  versus zero-padded canonical addresses. If you compare type arguments across
  transports, canonicalize the address portion first.
- `transaction.transaction.inputs` pure input representation: JSON-RPC
  historically exposes pure inputs as `{type, value, valueType}` and v2
  normalizes them toward raw BCS bytes when possible. gRPC may expose either a
  typed literal form (for example `PureU64Input`, `PureAddressInput`,
  `PureVecInput`) or fall back to `PureBCSInput` depending on what metadata the
  upstream protobuf response carries. These representations are semantically
  equivalent PTB inputs, but they are not guaranteed to serialize identically.

## Scope

v2 is a contract-normalization layer, not a promise that JSON-RPC and gRPC are
implemented by the Sui network as identical wire protocols. When a transport
difference cannot be losslessly removed, the v2 API prefers:

- one shared exported field name,
- best-effort normalization,
- and explicit documentation of the remaining difference.
