# Sui-Go-SDK

<p align="center">
    <a href="https://github.com/block-vision/sui-go-sdk/blob/main/.github/workflows/ci.yml"><img src="https://github.com/block-vision/sui-go-sdk/actions/workflows/ci.yml/badge.svg"></a>
    <a href="LICENSE"><img src="https://img.shields.io/badge/License-Apache_2.0-red.svg"></a>
    <a href="https://goreportcard.com/report/github.com/block-vision/sui-go-sdk"><img src="https://goreportcard.com/badge/github.com/securego/gosec"></a>
    <a href="https://pkg.go.dev/github.com/block-vision/sui-go-sdk"> <img src="https://pkg.go.dev/badge/github.com/block-vision/sui-go-sdk.svg"></a>
    <a href="https://discord.gg/Re6prK86Tr"><img src="https://img.shields.io/badge/chat-on%20discord-7289da.svg?sanitize=true"></a>
</p>

## Overview

The Sui-Go-SDK aims to offer access to all Sui RPC methods with Golang and also offers some additional features that make the integration easier. Powered by [SuiVision](https://suivision.xyz/) team.

> Check this [doc](https://docs.blockvision.org/reference/grpc-for-sui) to get your API Key and tutorials for Sui gRPC integration.

### V2 Client (Recommended)

We now offer a v2 client (`sui/v2`) with a unified API that works with both JSON-RPC and gRPC backends. Use the same request/response types regardless of transport—switch between JSON-RPC and gRPC by changing client options only.

See the [V2 README](sui/v2/README.md) for design goals, usage examples, and compatibility notes.

### Features

+ Support the mainstream methods in the Object, Coin, Event, Read Transaction Blocks, System Data, and Write Transaction
  Blocks modules.
+ **gRPC Support**: Full support for Sui gRPC services including Ledger Service, Name Service, Transaction Execution Service, Subscription Service, Live Data Service, Move Package Service, and Signature Verification Service.
+ Customized request method `SuiCall`.
+ Unsigned methods can be executed without loading your keystore file.
+ Provide the method `SignAndExecuteTransactionBlock` to send signed transaction.
+ Support subscriptions to events or transactions via websockets.

## Quick Start

### Install

```shell
go get github.com/block-vision/sui-go-sdk
```

### Go Version

| Golang Version |
|----------------|
| \>= 1.20       | 

## Examples

### Connecting to Sui Network

You can use the NewSuiClient method and pass in the RPC URL as an argument to easily connect to the Sui network.
BlockVision provides the following free and fast Sui network endpoints:

+ MainNet: https://sui-mainnet-endpoint.blockvision.org
+ TestNet: https://sui-testnet-endpoint.blockvision.org

```go
package main

import (
	"context"
	"fmt"

	"github.com/block-vision/sui-go-sdk/sui"
)

func main() {
	// configure your endpoint here or use BlockVision's free Sui RPC endpoint
	cli := sui.NewSuiClient("https://sui-testnet-endpoint.blockvision.org")
}

```

### Connecting to Sui Network via gRPC

You can also connect to the Sui network using gRPC for better performance and real-time data access:

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/block-vision/sui-go-sdk/common/grpcconn"
	"github.com/block-vision/sui-go-sdk/constant"
	"github.com/block-vision/sui-go-sdk/pb/sui/rpc/v2beta2"
)

func main() {
	// Create gRPC client with authentication
	client := grpcconn.NewSuiGrpcClientWithAuth(
		constant.SuiMainnetGrpcEndpoint,
		constant.SuiMainnetGrpcToken,
	)
	defer client.Close()

	ctx := context.Background()

	// Get ledger service
	ledgerService, err := client.LedgerService(ctx)
	if err != nil {
		log.Fatalf("Failed to get ledger service: %v", err)
	}

	// Example: Get service info
	info, err := ledgerService.GetServiceInfo(ctx, &v2beta2.GetServiceInfoRequest{})
	if err != nil {
		log.Fatalf("Failed to get service info: %v", err)
	}

	fmt.Printf("Service info: %+v\n", info)
}

```
### Getting coins from the faucet
You can request sui from the faucet when running against devnet, testnet, or localnet

```go
package main

import (
	"fmt"
	"github.com/block-vision/sui-go-sdk/constant"
	"github.com/block-vision/sui-go-sdk/sui"
)

func main() {
	RequestDevNetSuiFromFaucet()
}

func RequestDevNetSuiFromFaucet() {
	faucetHost, err := sui.GetFaucetHost(constant.SuiDevnet)
	if err != nil {
		fmt.Println("GetFaucetHost err:", err)
		return
	}

	fmt.Println("faucetHost:", faucetHost)

	recipient := "0xaf9f4d20c205f26051a7e1758601c4c47a9f99df3f9823f70926c17c80882d36"

	header := map[string]string{}
	err = sui.RequestSuiFromFaucet(faucetHost, recipient, header)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// the successful transaction block url: https://suiexplorer.com/txblock/91moaxbXsQnJYScLP2LpbMXV43ZfngS2xnRgj1CT7jLQ?network=devnet
	fmt.Println("Request DevNet Sui From Faucet success")
}

```


### Writing Transaction Blocks to Sui

#### Transfer Object

```go
package main

import (
  "context"
  "fmt"
  "github.com/block-vision/sui-go-sdk/constant"
  "github.com/block-vision/sui-go-sdk/models"
  "github.com/block-vision/sui-go-sdk/signer"
  "github.com/block-vision/sui-go-sdk/sui"
  "github.com/block-vision/sui-go-sdk/utils"
)

func main() {
  var ctx = context.Background()
  var cli = sui.NewSuiClient(constant.BvTestnetEndpoint)

  signerAccount, err := signer.NewSignertWithMnemonic("input your mnemonic")

  if err != nil {
    fmt.Println(err.Error())
    return
  }

  priKey := signerAccount.PriKey
  fmt.Printf("signerAccount.Address: %s\n", signerAccount.Address)

  gasObj := "0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1"

  rsp, err := cli.TransferObject(ctx, models.TransferObjectRequest{
    Signer:    signerAccount.Address,
    ObjectId:  "0x99b51302b66bd65b070cdb549b86e4b9aa7370cfddc70211c2b5a478140c7999",
    Gas:       &gasObj,
    GasBudget: "100000000",
    Recipient: "0xaf9f4d20c205f26051a7e1758601c4c47a9f99df3f9823f70926c17c80882d36",
  })

  if err != nil {
    fmt.Println(err.Error())
    return
  }

  // see the successful transaction url: https://explorer.sui.io/txblock/71xJsyNDRpGV96Dpw2FtjbWgJn2b5yP7KkYC13TGC5n9?network=testnet
  rsp2, err := cli.SignAndExecuteTransactionBlock(ctx, models.SignAndExecuteTransactionBlockRequest{
    TxnMetaData: rsp,
    PriKey:      priKey,
    // only fetch the effects field
    Options: models.SuiTransactionBlockOptions{
      ShowInput:    true,
      ShowRawInput: true,
      ShowEffects:  true,
    },
    RequestType: "WaitForLocalExecution",
  })

  if err != nil {
    fmt.Println(err.Error())
    return
  }

  utils.PrettyPrint(rsp2)
}

```

#### Transfer Sui

```go
package main

import (
	"context"
	"fmt"
	"github.com/block-vision/sui-go-sdk/constant"
	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/signer"
	"github.com/block-vision/sui-go-sdk/sui"
	"github.com/block-vision/sui-go-sdk/utils"
)

func main() {
	var ctx = context.Background()
	var cli = sui.NewSuiClient(constant.BvTestnetEndpoint)

	signerAccount, err := signer.NewSignertWithMnemonic("input your mnemonic")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	priKey := signerAccount.PriKey
	fmt.Printf("signerAccount.Address: %s\n", signerAccount.Address)

	rsp, err := cli.TransferSui(ctx, models.TransferSuiRequest{
		Signer:      signerAccount.Address,
		SuiObjectId: "0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1",
		GasBudget:   "100000000",
		Recipient:   "0xb7f98d327f19f674347e1e40641408253142d6e7e5093a7c96eda8cdfd7d9bb5",
		Amount:      "1",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// see the successful transaction url: https://explorer.sui.io/txblock/C7iYsH4tU5RdY1KBeNax4mCBn3XLZ5UswsuDpKrVkcH6?network=testnet
	rsp2, err := cli.SignAndExecuteTransactionBlock(ctx, models.SignAndExecuteTransactionBlockRequest{
		TxnMetaData: rsp,
		PriKey:      priKey,
      // only fetch the effects field
		Options: models.SuiTransactionBlockOptions{
			ShowInput:    true,
			ShowRawInput: true,
			ShowEffects:  true,
		},
		RequestType: "WaitForLocalExecution",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp2)
}

```

#### Move Call

```go

package main

import (
	"context"
	"fmt"
	"github.com/block-vision/sui-go-sdk/constant"
	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/signer"
	"github.com/block-vision/sui-go-sdk/sui"
	"github.com/block-vision/sui-go-sdk/utils"
)

func main() {
	var ctx = context.Background()
	var cli = sui.NewSuiClient(constant.BvTestnetEndpoint)

	signerAccount, err := signer.NewSignertWithMnemonic("input your mnemonic")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	priKey := signerAccount.PriKey
	fmt.Printf("signerAccount.Address: %s\n", signerAccount.Address)

  gasObj := "0x58c103930dc52c0ab86319d99218e301596fda6fd80c4efafd7f4c9df1d0b6d0"

	rsp, err := cli.MoveCall(ctx, models.MoveCallRequest{
		Signer:          signerAccount.Address,
		PackageObjectId: "0x7d584c9a27ca4a546e8203b005b0e9ae746c9bec6c8c3c0bc84611bcf4ceab5f",
		Module:          "auction",
		Function:        "start_an_auction",
		TypeArguments:   []interface{}{},
		Arguments: []interface{}{
			"0x342e959f8d9d1fa9327a05fd54fefd929bbedad47190bdbb58743d8ba3bd3420",
			"0x3fd0fdedb84cf1f59386b6251ba6dd2cb495094da26e0a5a38239acd9d437f96",
			"0xb3de4235cb04167b473de806d00ba351e5860500253cf8e62d711e578e1d92ae",
			"BlockVision",
			"0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1",
		},
		Gas:       &gasObj,
		GasBudget: "100000000",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// see the successful transaction url: https://explorer.sui.io/txblock/CD5hFB4bWFThhb6FtvKq3xAxRri72vsYLJAVd7p9t2sR?network=testnet
	rsp2, err := cli.SignAndExecuteTransactionBlock(ctx, models.SignAndExecuteTransactionBlockRequest{
		TxnMetaData: rsp,
		PriKey:      priKey,
      // only fetch the effects field
		Options: models.SuiTransactionBlockOptions{
			ShowInput:    true,
			ShowRawInput: true,
			ShowEffects:  true,
		},
		RequestType: "WaitForLocalExecution",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp2)
}

```

#### Merge coins

```go

package main

import (
  "context"
  "fmt"
  "github.com/block-vision/sui-go-sdk/constant"
  "github.com/block-vision/sui-go-sdk/models"
  "github.com/block-vision/sui-go-sdk/signer"
  "github.com/block-vision/sui-go-sdk/sui"
  "github.com/block-vision/sui-go-sdk/utils"
)

func main() {
  var ctx = context.Background()
  var cli = sui.NewSuiClient(constant.BvTestnetEndpoint)

  signerAccount, err := signer.NewSignertWithMnemonic("input your mnemonic")

  if err != nil {
    fmt.Println(err.Error())
    return
  }

  priKey := signerAccount.PriKey
  fmt.Printf("signerAccount.Address: %s\n", signerAccount.Address)

  gasObj := "0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1"

  rsp, err := cli.MergeCoins(ctx, models.MergeCoinsRequest{
    Signer:      signerAccount.Address,
    PrimaryCoin: "0x180fe0c159644fe4b376e4488498e524b2a564919775cb2719734a4699ae7b28",
    CoinToMerge: "0x3b4644f82b4dc339c17ed5f786f4050e1f765b38e9297ffdacdfc5ead482669f",
    Gas:         &gasObj,
    GasBudget:   "100000000",
  })

  if err != nil {
    fmt.Println(err.Error())
    return
  }

  utils.PrettyPrint(rsp)

  // see the successful transaction url: https://explorer.sui.io/txblock/DZrnvnk67b27KQXisQA7VfnBUnga2SMnWNMY3UCZPG5a?network=testnet
  rsp2, err := cli.SignAndExecuteTransactionBlock(ctx, models.SignAndExecuteTransactionBlockRequest{
    TxnMetaData: rsp,
    PriKey:      priKey,
    // only fetch the effects field
    Options: models.SuiTransactionBlockOptions{
      ShowInput:    true,
      ShowRawInput: true,
      ShowEffects:  true,
    },
    RequestType: "WaitForLocalExecution",
  })

  if err != nil {
    fmt.Println(err.Error())
    return
  }

  utils.PrettyPrint(rsp2)
}

```

### Reading Data from Sui

#### Get the address all balance

Fetch all balance owned by the address `0xb7f98d327f19f674347e1e40641408253142d6e7e5093a7c96eda8cdfd7d9bb5`.

```go
package main

import (
	"context"
	"fmt"

	"github.com/block-vision/sui-go-sdk/constant"
	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/sui"
	"github.com/block-vision/sui-go-sdk/utils"
)

func main() {
	var ctx = context.Background()
	var cli = sui.NewSuiClient(constant.BvTestnetEndpoint)

	rsp, err := cli.SuiXGetAllBalance(ctx, models.SuiXGetAllBalanceRequest{
		Owner: "0xb7f98d327f19f674347e1e40641408253142d6e7e5093a7c96eda8cdfd7d9bb5",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp)

	// If you want to request for original json response data, you can use SuiCall().
	callRsp, err := cli.SuiCall(ctx, "suix_getAllBalances", "0xb7f98d327f19f674347e1e40641408253142d6e7e5093a7c96eda8cdfd7d9bb5")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(callRsp)
}

```

#### Get Coin Metadata

Fetch coin metadata by the CoinType `0xf7a0b8cc24808220226301e102dae27464ea46e0d74bb968828318b9e3a921fa`.

```go
package main

import (
	"context"
	"fmt"

	"github.com/block-vision/sui-go-sdk/constant"
	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/sui"
	"github.com/block-vision/sui-go-sdk/utils"
)

func main() {
	var ctx = context.Background()
	var cli = sui.NewSuiClient(constant.BvTestnetEndpoint)

	rsp, err := cli.SuiXGetCoinMetadata(ctx, models.SuiXGetCoinMetadataRequest{
		CoinType: "0xf7a0b8cc24808220226301e102dae27464ea46e0d74bb968828318b9e3a921fa::busd::BUSD",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp)
}

```

#### Get Owned Objects

Fetch objects owned by the address `0xb7f98d327f19f674347e1e40641408253142d6e7e5093a7c96eda8cdfd7d9bb5`.

```go
package main

import (
  "context"
  "fmt"

  "github.com/block-vision/sui-go-sdk/constant"
  "github.com/block-vision/sui-go-sdk/models"
  "github.com/block-vision/sui-go-sdk/sui"
  "github.com/block-vision/sui-go-sdk/utils"
)

func main() {
  var ctx = context.Background()
  var cli = sui.NewSuiClient(constant.BvTestnetEndpoint)

  suiObjectResponseQuery := models.SuiObjectResponseQuery{
    // only fetch the effects field
    Options: models.SuiObjectDataOptions{
      ShowType:    true,
      ShowContent: true,
      ShowBcs:     true,
      ShowOwner:   true,
    },
  }

  rsp, err := cli.SuiXGetOwnedObjects(ctx, models.SuiXGetOwnedObjectsRequest{
    Address: "0xb7f98d327f19f674347e1e40641408253142d6e7e5093a7c96eda8cdfd7d9bb5",
    Query:   suiObjectResponseQuery,
    Limit:   5,
  })

  if err != nil {
    fmt.Println(err.Error())
    return
  }

  utils.PrettyPrint(rsp)
}

```

#### Get Object

Fetch object details for the object with id `0xeeb964d1e640219c8ddb791cc8548f3242a3392b143ff47484a3753291cad898`.


```go
package main

import (
  "context"
  "fmt"

  "github.com/block-vision/sui-go-sdk/constant"
  "github.com/block-vision/sui-go-sdk/models"
  "github.com/block-vision/sui-go-sdk/sui"
  "github.com/block-vision/sui-go-sdk/utils"
)

func main() {
  var ctx = context.Background()
  var cli = sui.NewSuiClient(constant.BvTestnetEndpoint)

  rsp, err := cli.SuiGetObject(ctx, models.SuiGetObjectRequest{
    ObjectId: "0xeeb964d1e640219c8ddb791cc8548f3242a3392b143ff47484a3753291cad898",
    // only fetch the effects field
    Options: models.SuiObjectDataOptions{
      ShowContent:             true,
      ShowDisplay:             true,
      ShowType:                true,
      ShowBcs:                 true,
      ShowOwner:               true,
      ShowPreviousTransaction: true,
      ShowStorageRebate:       true,
    },
  })

  if err != nil {
    fmt.Println(err.Error())
    return
  }

  utils.PrettyPrint(rsp)
}

```

#### Get Transaction

Fetch transaction details with digests `CeVpDXKKU3Gs89efej9pKiYYQyTzifE2BDxWwquUaUht`.

```go
package main

import (
  "context"
  "fmt"

  "github.com/block-vision/sui-go-sdk/constant"
  "github.com/block-vision/sui-go-sdk/models"
  "github.com/block-vision/sui-go-sdk/sui"
  "github.com/block-vision/sui-go-sdk/utils"
)

func main() {
  var ctx = context.Background()
  var cli = sui.NewSuiClient(constant.BvTestnetEndpoint)

  rsp, err := cli.SuiGetTransactionBlock(ctx, models.SuiGetTransactionBlockRequest{
    Digest: "CeVpDXKKU3Gs89efej9pKiYYQyTzifE2BDxWwquUaUht",
    // only fetch the effects field
    Options: models.SuiTransactionBlockOptions{
      ShowInput:          true,
      ShowRawInput:       true,
      ShowEffects:        true,
      ShowEvents:         true,
      ShowBalanceChanges: true,
      ShowObjectChanges:  true,
    },
  })

  if err != nil {
    fmt.Println(err.Error())
    return
  }

  utils.PrettyPrint(rsp)

  // fetch multiple transactions in one batch request
  rsp2, err := cli.SuiMultiGetTransactionBlocks(ctx, models.SuiMultiGetTransactionBlocksRequest{
    Digests: []string{"CeVpDXKKU3Gs89efej9pKiYYQyTzifE2BDxWwquUaUht", "C2zZu2dpX2sLQy2234yt6ecRiNTVgQTXeQpgw9GhxGgo"},
    // only fetch the effects field
    Options: models.SuiTransactionBlockOptions{
      ShowInput:          true,
      ShowRawInput:       true,
      ShowEffects:        true,
      ShowEvents:         true,
      ShowObjectChanges:  true,
      ShowBalanceChanges: true,
    },
  })

  if err != nil {
    fmt.Println(err.Error())
    return
  }

  for _, transactionBlock := range rsp2 {
    utils.PrettyPrint(*transactionBlock)
  }

}

```

#### Get Checkpoints

Get latest 10 Checkpoints in descending order.

```go
package main

import (
  "context"
  "fmt"

  "github.com/block-vision/sui-go-sdk/constant"
  "github.com/block-vision/sui-go-sdk/models"
  "github.com/block-vision/sui-go-sdk/sui"
  "github.com/block-vision/sui-go-sdk/utils"
)

func main() {
  var ctx = context.Background()
  var cli = sui.NewSuiClient(constant.BvTestnetEndpoint)

  rsp, err := cli.SuiGetCheckpoints(ctx, models.SuiGetCheckpointsRequest{
    Limit:           10,
    DescendingOrder: true,
  })

  if err != nil {
    fmt.Println(err.Error())
    return
  }

  utils.PrettyPrint(rsp)

  // fetch Checkpoint 1628214 and print details.
  rsp2, err := cli.SuiGetCheckpoint(ctx, models.SuiGetCheckpointRequest{
    CheckpointID: "1628214",
  })

  if err != nil {
    fmt.Println(err.Error())
    return
  }

  utils.PrettyPrint(rsp2)

}

```

#### Query Events

Fetch event details with digests `CeVpDXKKU3Gs89efej9pKiYYQyTzifE2BDxWwquUaUht`.

```go
package main

import (
  "context"
  "fmt"

  "github.com/block-vision/sui-go-sdk/constant"
  "github.com/block-vision/sui-go-sdk/models"
  "github.com/block-vision/sui-go-sdk/sui"
  "github.com/block-vision/sui-go-sdk/utils"
)

func main() {
  var ctx = context.Background()
  var cli = sui.NewSuiClient(constant.BvTestnetEndpoint)

  rsp, err := cli.SuiGetEvents(ctx, models.SuiGetEventsRequest{
    Digest: "HATq5p7MNynkBL5bLsdVqL3K38PxWHbqs7vndGiz5qrA",
  })

  if err != nil {
    fmt.Println(err.Error())
    return
  }

  utils.PrettyPrint(rsp)

  // fetch list of events for a specified query criteria.
  rsp2, err := cli.SuiXQueryEvents(ctx, models.SuiXQueryEventsRequest{
    SuiEventFilter: models.EventFilterByMoveEventType{
      MoveEventType: "0x3::validator::StakingRequestEvent",
    },
    Limit:           5,
    DescendingOrder: true,
  })

  if err != nil {
    fmt.Println(err.Error())
    return
  }

  utils.PrettyPrint(rsp2)

}

```

### Subscribe API

#### Subscribe event API

Subscribe to a stream of Sui event.

```go

package main

import (
  "context"
  "github.com/block-vision/sui-go-sdk/constant"
  "github.com/block-vision/sui-go-sdk/models"
  "github.com/block-vision/sui-go-sdk/sui"
  "github.com/block-vision/sui-go-sdk/utils"
)

func main() {
  var ctx = context.Background()
  // create a websocket client, connect to the mainnet websocket endpoint
  var cli = sui.NewSuiWebsocketClient(constant.WssBvMainnetEndpoint)

  // receiveMsgCh is a channel to receive Sui event
  receiveMsgCh := make(chan models.SuiEventResponse, 10)
  
  // SubscribeEvent implements the method `suix_subscribeEvent`, subscribe to a stream of Sui event.
  err := cli.SubscribeEvent(ctx, models.SuiXSubscribeEventsRequest{
    SuiEventFilter: map[string]interface{}{
      "All": []string{},
    },
  }, receiveMsgCh)
  if err != nil {
    panic(err)
  }

  for {
    select {
	// receive Sui event
    case msg := <-receiveMsgCh:
      utils.PrettyPrint(msg)
    case <-ctx.Done():
      return
    }
  }

}

```

#### Subscribe transaction API

Subscribe to a stream of Sui transaction effects.

```go

package main

import (
  "context"
  "github.com/block-vision/sui-go-sdk/constant"
  "github.com/block-vision/sui-go-sdk/models"
  "github.com/block-vision/sui-go-sdk/sui"
  "github.com/block-vision/sui-go-sdk/utils"
)

func main() {
  var ctx = context.Background()
  // create a websocket client, connect to the mainnet websocket endpoint
  var cli = sui.NewSuiWebsocketClient(constant.WssBvMainnetEndpoint)

  // receiveMsgCh is a channel to receive Sui transaction effects
  receiveMsgCh := make(chan models.SuiEffects, 10)

  // SubscribeTransaction implements the method `suix_subscribeTransaction`, subscribe to a stream of Sui transaction effects.
  err := cli.SubscribeTransaction(ctx, models.SuiXSubscribeTransactionsRequest{
    TransactionFilter: models.TransactionFilterByFromAddress{
      FromAddress: "0x0000000000000000000000000000000000000000000000000000000000000000",
    },
  }, receiveMsgCh)
  if err != nil {
    panic(err)
  }

  for {
    select {
    // receive Sui transaction effects
    case msg := <-receiveMsgCh:
      utils.PrettyPrint(msg)
    case <-ctx.Done():
      return
    }
  }

}

```

## API Documentation
The Go Client SDK API documentation is currently available at [pkg.go.dev](https://pkg.go.dev/github.com/block-vision/sui-go-sdk).

## Contribution

+ We welcome your suggestions, comments (including criticisms) and contributions.
+ Please follow the PR/issue template provided to ensure that your contributions are clear and easy to understand.
+ Thank you to all the people who participate in building better infrastructure!

## Resources

+ [SDK Examples](https://github.com/block-vision/sui-go-sdk/tree/main/examples)
+ [Sui](https://github.com/MystenLabs/sui)

## License

[Apache 2.0 license](LICENSE)




