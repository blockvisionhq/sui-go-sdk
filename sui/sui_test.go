// Copyright (c) BlockVision, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package sui

import (
	"context"
	"fmt"
	"github.com/block-vision/sui-go-sdk/constant"
	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/utils"
	"testing"
)

var ctx = context.Background()
var cli = NewSuiClientWithCustomClient(constant.BvMainnetEndpoint, newThrottledTestHTTPClient())

func TestOnReadSystemFromSui(t *testing.T) {
	t.Run("test on sui_getCheckpoint", func(t *testing.T) {
		rsp, err := cli.SuiGetCheckpoint(ctx, models.SuiGetCheckpointRequest{
			CheckpointID: "1628214",
		})

		if err != nil {
			t.Error(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)
	})

	t.Run("test on sui_getCheckpoints", func(t *testing.T) {
		rsp, err := cli.SuiGetCheckpoints(ctx, models.SuiGetCheckpointsRequest{
			Limit:           5,
			DescendingOrder: true,
		})

		if err != nil {
			t.Error(err.Error())
			t.FailNow()
		}

		for _, checkpoint := range rsp.Data {
			utils.PrettyPrint(checkpoint)
		}
	})

	t.Run("test on sui_getLatestCheckpointSequenceNumber", func(t *testing.T) {
		rsp, err := cli.SuiGetLatestCheckpointSequenceNumber(ctx)

		if err != nil {
			t.Error(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)
	})

	t.Run("test on suix_getReferenceGasPrice", func(t *testing.T) {
		rsp, err := cli.SuiXGetReferenceGasPrice(ctx)

		if err != nil {
			t.Error(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)
	})

	t.Run("test on suix_getCommitteeInfo", func(t *testing.T) {
		rsp, err := cli.SuiXGetCommitteeInfo(ctx, models.SuiXGetCommitteeInfoRequest{
			Epoch: "39",
		})

		if err != nil {
			t.Error(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)
	})

	t.Run("test on suix_getReferenceGasPrice", func(t *testing.T) {
		rsp, err := cli.SuiXGetReferenceGasPrice(ctx)

		if err != nil {
			t.Error(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)
	})

	t.Run("test on suix_getStakes", func(t *testing.T) {
		rsp, err := cli.SuiXGetStakes(ctx, models.SuiXGetStakesRequest{
			Owner: "0xd939e3fe7ea4d503f84767dca0c58b7ec1c71f085638a4c0611aa64aa71b5fcf",
		})

		if err != nil {
			t.Error(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)
	})

	t.Run("test on suix_getStakesByIds", func(t *testing.T) {
		rsp, err := cli.SuiXGetStakesByIds(ctx, models.SuiXGetStakesByIdsRequest{
			StakedSuiIds: []string{"0x02cfd8057d8a499bcd936ba65efd65889e66874b3819cb251fe9b9799048f1ed"},
		})

		if err != nil {
			t.Error(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)
	})

	t.Run("test on suix_getLatestSuiSystemState", func(t *testing.T) {
		rsp, err := cli.SuiXGetLatestSuiSystemState(ctx)

		if err != nil {
			t.Error(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)
	})

	t.Run("test on sui_getChainIdentifier", func(t *testing.T) {
		rsp, err := cli.SuiGetChainIdentifier(ctx)

		if err != nil {
			t.Error(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)
	})

	t.Run("test on suix_getValidatorsApy", func(t *testing.T) {
		rsp, err := cli.SuiXGetValidatorsApy(ctx)

		if err != nil {
			t.Error(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)
	})

	t.Run("test on sui_getProtocolConfig", func(t *testing.T) {
		rsp, err := cli.SuiGetProtocolConfig(ctx, models.SuiGetProtocolConfigRequest{})

		if err != nil {
			t.Error(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)
	})
}

func TestOnReadCoinFromSui(t *testing.T) {

	t.Run("test on suix_getBalance", func(t *testing.T) {
		rsp, err := cli.SuiXGetBalance(ctx, models.SuiXGetBalanceRequest{
			Owner:    "0xd939e3fe7ea4d503f84767dca0c58b7ec1c71f085638a4c0611aa64aa71b5fcf",
			CoinType: "0x2::sui::SUI",
		})

		if err != nil {
			t.Error(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)
	})

	t.Run("test on suix_getAllBalances", func(t *testing.T) {
		rsp, err := cli.SuiXGetAllBalance(ctx, models.SuiXGetAllBalanceRequest{
			Owner: "0xd939e3fe7ea4d503f84767dca0c58b7ec1c71f085638a4c0611aa64aa71b5fcf",
		})

		if err != nil {
			t.Error(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)
	})

	t.Run("test on suix_getCoins", func(t *testing.T) {
		rsp, err := cli.SuiXGetCoins(ctx, models.SuiXGetCoinsRequest{
			Owner:    "0xd939e3fe7ea4d503f84767dca0c58b7ec1c71f085638a4c0611aa64aa71b5fcf",
			CoinType: "0x2::sui::SUI",
			Limit:    5,
		})

		if err != nil {
			t.Error(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)
	})

	t.Run("test on suix_getAllCoins", func(t *testing.T) {
		rsp, err := cli.SuiXGetAllCoins(ctx, models.SuiXGetAllCoinsRequest{
			Owner: "0xd939e3fe7ea4d503f84767dca0c58b7ec1c71f085638a4c0611aa64aa71b5fcf",
			Limit: 5,
		})

		if err != nil {
			t.Error(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)
	})

	t.Run("test on suix_getCoinMetadata", func(t *testing.T) {
		rsp, err := cli.SuiXGetCoinMetadata(ctx, models.SuiXGetCoinMetadataRequest{
			CoinType: "0x06864a6f921804860930db6ddbe2e16acdf8504495ea7481637a1c8b9a8fe54b::cetus::CETUS",
		})

		if err != nil {
			t.Error(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)
	})

	t.Run("test on suix_getTotalSupply", func(t *testing.T) {
		rsp, err := cli.SuiXGetTotalSupply(ctx, models.SuiXGetTotalSupplyRequest{
			CoinType: "0x06864a6f921804860930db6ddbe2e16acdf8504495ea7481637a1c8b9a8fe54b::cetus::CETUS",
		})

		if err != nil {
			t.Error(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)
	})
}

func TestOnReadTransactionFromSui(t *testing.T) {

	t.Run("test on sui_getTotalTransactionBlocks", func(t *testing.T) {
		rsp, err := cli.SuiGetTotalTransactionBlocks(ctx)

		if err != nil {
			fmt.Println()
			t.FailNow()
		}

		utils.PrettyPrint(rsp)
	})

	t.Run("test on sui_getTransactionBlock", func(t *testing.T) {
		rsp, err := cli.SuiGetTransactionBlock(ctx, models.SuiGetTransactionBlockRequest{
			Digest: "2LYaFDf5oU64xguKAjSiH7TarPSkxc35sN6rPc8RsoWf",
			Options: models.SuiTransactionBlockOptions{
				ShowInput:    true,
				ShowRawInput: true,
				ShowEffects:  true,
				ShowEvents:   true,
			},
		})

		if err != nil {
			fmt.Println(err)
			t.FailNow()
		}

		utils.PrettyPrint(rsp)
	})

	t.Run("test on sui_multiGetTransactionBlocks", func(t *testing.T) {
		rsp, err := cli.SuiMultiGetTransactionBlocks(ctx, models.SuiMultiGetTransactionBlocksRequest{
			Digests: []string{"2LYaFDf5oU64xguKAjSiH7TarPSkxc35sN6rPc8RsoWf", "rHrHjircrKMRP5fRb6YW7ez2cdL3JhQhzoZDv4THEFk"},
			Options: models.SuiTransactionBlockOptions{
				ShowInput:    true,
				ShowRawInput: true,
				ShowEffects:  true,
			},
		})

		if err != nil {
			fmt.Println(err)
			t.FailNow()
		}

		for _, transactionBlock := range rsp {
			utils.PrettyPrint(*transactionBlock)
		}
	})

	t.Run("test on suix_queryTransactionBlocks", func(t *testing.T) {
		rsp, err := cli.SuiXQueryTransactionBlocks(ctx, models.SuiXQueryTransactionBlocksRequest{
			SuiTransactionBlockResponseQuery: models.SuiTransactionBlockResponseQuery{
				TransactionFilter: models.TransactionFilter{
					"FromAddress": "0x02bcc205ccf48ac87f081f907ddbd46de66f847afbb6a8b11801240132f4eec5",
				},
				Options: models.SuiTransactionBlockOptions{
					ShowInput:    true,
					ShowRawInput: true,
					ShowEffects:  true,
				},
			},
			Limit:           5,
			DescendingOrder: false,
		})
		if err != nil {
			fmt.Println(err)
			t.FailNow()
		}

		for _, transactionBlock := range rsp.Data {
			utils.PrettyPrint(transactionBlock)
		}
	})
}

func TestOnReadObjectFromSui(t *testing.T) {
	t.Run("test on sui_getObject", func(t *testing.T) {
		rsp, err := cli.SuiGetObject(ctx, models.SuiGetObjectRequest{
			ObjectId: "0x02cfd8057d8a499bcd936ba65efd65889e66874b3819cb251fe9b9799048f1ed",
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
			t.Error(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)
	})

	t.Run("test on suix_getOwnedObjects", func(t *testing.T) {
		suiObjectResponseQuery := models.SuiObjectResponseQuery{
			Options: models.SuiObjectDataOptions{
				ShowType:                true,
				ShowContent:             true,
				ShowBcs:                 true,
				ShowOwner:               true,
				ShowPreviousTransaction: true,
				ShowStorageRebate:       true,
				ShowDisplay:             true,
			},
		}

		rsp, err := cli.SuiXGetOwnedObjects(ctx, models.SuiXGetOwnedObjectsRequest{
			Address: "0xd939e3fe7ea4d503f84767dca0c58b7ec1c71f085638a4c0611aa64aa71b5fcf",
			Query:   suiObjectResponseQuery,
			Limit:   5,
		})

		if err != nil {
			t.Error(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)
	})

	t.Run("test on sui_multiGetObjects", func(t *testing.T) {
		rsp, err := cli.SuiMultiGetObjects(ctx, models.SuiMultiGetObjectsRequest{
			ObjectIds: []string{"0x02cfd8057d8a499bcd936ba65efd65889e66874b3819cb251fe9b9799048f1ed"},
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
			t.Error(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)

	})

	t.Run("test on suix_getDynamicFields", func(t *testing.T) {
		rsp, err := cli.SuiXGetDynamicField(ctx, models.SuiXGetDynamicFieldRequest{
			ObjectId: "0x02cfd8057d8a499bcd936ba65efd65889e66874b3819cb251fe9b9799048f1ed",
			Limit:    5,
		})

		if err != nil {
			t.Error(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)

	})

	t.Run("test on sui_tryGetPastObject", func(t *testing.T) {
		rsp, err := cli.SuiTryGetPastObject(ctx, models.SuiTryGetPastObjectRequest{
			ObjectId: "0x02cfd8057d8a499bcd936ba65efd65889e66874b3819cb251fe9b9799048f1ed",
			Version:  9636,
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
			t.Error(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)

	})

	t.Run("test on sui_tryMultiGetPastObjects", func(t *testing.T) {
		rsp, err := cli.SuiTryMultiGetPastObjects(ctx, models.SuiTryMultiGetPastObjectsRequest{
			MultiGetPastObjects: []*models.PastObject{
				{
					ObjectId: "0xfe3e114168d65ca9c86e43ce0f8dc4f8e0fa5a03634a4c6bf292679f6d73ec72",
					Version:  "22945798",
				},
				{
					ObjectId: "0xbf67e84fef313e6f1756411b095ba07868804852c939691b300a7e1e45d0251f",
					Version:  "23119685",
				},
			},
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
			t.Error(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)

	})
}

func TestOnReadEventFromSui(t *testing.T) {

	t.Run("test on sui_getEvents", func(t *testing.T) {
		rsp, err := cli.SuiGetEvents(ctx, models.SuiGetEventsRequest{
			Digest: "4ErUvWjWdXY5zdVkRCqgQFJZQDTgJmHo55RFJW2FWcs2",
		})

		if err != nil {
			t.Error(err.Error())
			t.FailNow()
		}

		for _, event := range rsp {
			utils.PrettyPrint(*event)
		}
	})

	t.Run("test on suix_queryEvents", func(t *testing.T) {
		rsp, err := cli.SuiXQueryEvents(ctx, models.SuiXQueryEventsRequest{
			SuiEventFilter: models.EventFilterByMoveEventType{
				MoveEventType: "0x3::validator::StakingRequestEvent",
			},
			Limit:           5,
			DescendingOrder: false,
		})

		if err != nil {
			t.Error(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)
	})
}

/*
func TestOnWriteTransactionToSui(t *testing.T) {

	t.Run("test on sui_executeTransactionBlock", func(t *testing.T) {
		rsp, err := cli.SuiExecuteTransactionBlock(ctx, models.SuiExecuteTransactionBlockRequest{
			TxBytes:   "AAACAQEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABQEAAAAAAAAAAQEAjgDW4hJZlqvw654RGR3SdndKkdjoC0pzXQLxja/NUahLowQAAAAAACBEQGwClI9RQX68dzbN7PN29/Pw/Sc1hbtZwNAny7wZ+wEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAMKc3VpX3N5c3RlbRZyZXF1ZXN0X3dpdGhkcmF3X3N0YWtlAAIBAAABAQC3+Y0yfxn2dDR+HkBkFAglMULW5+UJOnyW7ajN/X2btQEqzrI5x8BMQ6LjmCSgAykfjisdYCcyTfW79nyzDB/PvtZBpwAAAAAAIAm+IREDziwoZLm7lc4ZKegZ2J5viEgoss9zgrFkHLh6t/mNMn8Z9nQ0fh5AZBQIJTFC1uflCTp8lu2ozf19m7XoAwAAAAAAAFDhjyoAAAAAAA==",
			Signature: []string{"ALISOTYXKlmBvQ1Uc4UrlWieczU9cGwkyT0Mg65Y2r6VvriElBGB63JDjqg1714Z8B0m84g3S4yrJIIws+leugKOjKY5Wf3dV/la/GVL26whJPWy7WsrWUH2wtmlmcgN6w=="},
			Options: models.SuiTransactionBlockOptions{
				ShowInput:    true,
				ShowRawInput: true,
				ShowEffects:  true,
				ShowEvents:   true,
			},
			RequestType: "WaitForLocalExecution",
		})

		if err != nil {
			t.Error(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)
	})

	t.Run("test on unsafe_moveCall", func(t *testing.T) {
		rsp, err := cli.MoveCall(ctx, models.MoveCallRequest{
			Signer:          "0xb7f98d327f19f674347e1e40641408253142d6e7e5093a7c96eda8cdfd7d9bb5",
			PackageObjectId: "0x7d584c9a27ca4a546e8203b005b0e9ae746c9bec6c8c3c0bc84611bcf4ceab5f",
			Module:          "auction",
			Function:        "start_an_auction",
			TypeArguments:   []interface{}{},
			Arguments: []interface{}{
				"0x342e959f8d9d1fa9327a05fd54fefd929bbedad47190bdbb58743d8ba3bd3420",
				"0x3fd0fdedb84cf1f59386b6251ba6dd2cb495094da26e0a5a38239acd9d437f96",
				"0xb3de4235cb04167b473de806d00ba351e5860500253cf8e62d711e578e1d92ae",
				"web3",
				"0x2aceb239c7c04c43a2e39824a003291f8e2b1d6027324df5bbf67cb30c1fcfbe",
			},
			Gas:       "0x2aceb239c7c04c43a2e39824a003291f8e2b1d6027324df5bbf67cb30c1fcfbe",
			GasBudget: "1000",
		})

		if err != nil {
			t.Error(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)

	})

	t.Run("test on unsafe_mergeCoins", func(t *testing.T) {
		rsp, err := cli.MergeCoins(ctx, models.MergeCoinsRequest{
			Signer:      "0xb7f98d327f19f674347e1e40641408253142d6e7e5093a7c96eda8cdfd7d9bb5",
			PrimaryCoin: "0x2aceb239c7c04c43a2e39824a003291f8e2b1d6027324df5bbf67cb30c1fcfbe",
			CoinToMerge: "0x92f03fdec6e0278dcb6fa3f4467eeee3e0bee1ac41825351ef53431677d2e2f7",
			Gas:         "0x2aceb239c7c04c43a2e39824a003291f8e2b1d6027324df5bbf67cb30c1fcfbe",
			GasBudget:   "1000",
		})

		if err != nil {
			fmt.Println(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)
	})

	t.Run("test on unsafe_publish", func(t *testing.T) {
		rsp, err := cli.Publish(ctx, models.PublishRequest{
			Sender: "0xb7f98d327f19f674347e1e40641408253142d6e7e5093a7c96eda8cdfd7d9bb5",
			CompiledModules: []string{
				"oRzrCwUAAAALAQAOAg4kAzJZBIsBFgWhAZoBB7sC5wEIogQoBsoECgrUBBoM7gSjAQ2RBgQAAAEBAQIBAwEEAQUBBgAHCAAACAgAAAkIAAIKDAEAAQQLAgAGDAIAAxIEAAEUBAEAAQANAAEAAA4CAQAADwMBAAAQBAEAAhUHCAEAAhYJCgEAARcLDAEAARgNCAEAAxkEDgAGGg8QAAUFEQEBCAEVEwgBAAIbFBUBAAMcDgEAAR0BDAEABR4ZAQEIBAYFBgYGBwYKAwsGDAYKFgoXDgYPGAMHCAEHCwMBCAQHCAUAAwYIAgcIAQcIBQEIAAEHCAUCBwsHAQgECwcBCAQBCAQBBgsDAQkAAQMBBwsDAQkAAQcLBwEJAAIHCwcBCQADAQsHAQkAAgcLBwEJAAsHAQkAAQgGAQYIBQEFAgkABQIDCwMBCAQBBgsHAQkAAwcLBwEJAAMHCAUBCwMBCQABCwMBCAQBCAIBCAEBCQAGZG9udXRzB2JhbGFuY2UEY29pbgZvYmplY3QDc3VpCHRyYW5zZmVyCnR4X2NvbnRleHQFRG9udXQJRG9udXRTaG9wDFNob3BPd25lckNhcARDb2luA1NVSQlUeENvbnRleHQJYnV5X2RvbnV0D2NvbGxlY3RfcHJvZml0cwllYXRfZG9udXQEaW5pdAJpZANVSUQFcHJpY2UHQmFsYW5jZQV2YWx1ZQtiYWxhbmNlX211dAVzcGxpdARqb2luA25ldwZzZW5kZXIEdGFrZQZkZWxldGUEemVybwxzaGFyZV9vYmplY3QAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACAwgAAAAAAAAAAAACAREIBgECAxEIBhMDAQsHAQgEAgIBEQgGAAEEAAUmCgEuOAAKABAAFCYDEAsAAQsBAQsCAQcAJwsBOAEMAwsDCgAQABQ4AgwECwAPAQsEOAMBCgIRCBIACwIuEQk4BAIBAQQAEhAKARABOAUMAwsBDwELAwoCOAYMBAsECwIuEQk4BwICAQQADgYLABMADAELARENAgMAAAABDgoAEQgSAgoALhEJOAgLABEIBugDAAAAAAAAOAkSATgKAgEBAQIA",
			},
			Dependencies: []string{"0x2aceb239c7c04c43a2e39824a003291f8e2b1d6027324df5bbf67cb30c1fcfbe"},
			Gas:          "0x2aceb239c7c04c43a2e39824a003291f8e2b1d6027324df5bbf67cb30c1fcfbe",
			GasBudget:    "1000",
		})

		if err != nil {
			fmt.Println(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)
	})

	t.Run("test on unsafe_splitCoin", func(t *testing.T) {
		rsp, err := cli.SplitCoin(ctx, models.SplitCoinRequest{
			Signer:       "0xb7f98d327f19f674347e1e40641408253142d6e7e5093a7c96eda8cdfd7d9bb5",
			CoinObjectId: "0x2aceb239c7c04c43a2e39824a003291f8e2b1d6027324df5bbf67cb30c1fcfbe",
			SplitAmounts: []string{"1000", "1000"},
			Gas:          "0x2aceb239c7c04c43a2e39824a003291f8e2b1d6027324df5bbf67cb30c1fcfbe",
			GasBudget:    "1000",
		})

		if err != nil {
			fmt.Println(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)
	})

	t.Run("test on unsafe_splitCoinEqual", func(t *testing.T) {
		rsp, err := cli.SplitCoinEqual(ctx, models.SplitCoinEqualRequest{
			Signer:       "0xb7f98d327f19f674347e1e40641408253142d6e7e5093a7c96eda8cdfd7d9bb5",
			CoinObjectId: "0x2aceb239c7c04c43a2e39824a003291f8e2b1d6027324df5bbf67cb30c1fcfbe",
			SplitCount:   "2",
			Gas:          "0x2aceb239c7c04c43a2e39824a003291f8e2b1d6027324df5bbf67cb30c1fcfbe",
			GasBudget:    "1000",
		})

		if err != nil {
			fmt.Println(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)
	})

	t.Run("test on unsafe_transferObject", func(t *testing.T) {
		rsp, err := cli.TransferObject(ctx, models.TransferObjectRequest{
			Signer:    "0xb7f98d327f19f674347e1e40641408253142d6e7e5093a7c96eda8cdfd7d9bb5",
			ObjectId:  "0x2aceb239c7c04c43a2e39824a003291f8e2b1d6027324df5bbf67cb30c1fcfbe",
			Gas:       "0x2aceb239c7c04c43a2e39824a003291f8e2b1d6027324df5bbf67cb30c1fcfbe",
			GasBudget: "1000",
			Recipient: "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
		})

		if err != nil {
			fmt.Println(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)
	})

	t.Run("test on unsafe_transferSui", func(t *testing.T) {
		rsp, err := cli.TransferSui(ctx, models.TransferSuiRequest{
			Signer:      "0xb7f98d327f19f674347e1e40641408253142d6e7e5093a7c96eda8cdfd7d9bb5",
			SuiObjectId: "0x2aceb239c7c04c43a2e39824a003291f8e2b1d6027324df5bbf67cb30c1fcfbe",
			GasBudget:   "1000",
			Recipient:   "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
			Amount:      "1",
		})

		if err != nil {
			fmt.Println(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)
	})

	t.Run("test on unsafe_pay", func(t *testing.T) {
		rsp, err := cli.Pay(ctx, models.PayRequest{
			Signer:      "0xb7f98d327f19f674347e1e40641408253142d6e7e5093a7c96eda8cdfd7d9bb5",
			SuiObjectId: []string{"0x92f03fdec6e0278dcb6fa3f4467eeee3e0bee1ac41825351ef53431677d2e2f7"},
			Recipient:   []string{"0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff"},
			Amount:      []string{"1"},
			Gas:         "0x2aceb239c7c04c43a2e39824a003291f8e2b1d6027324df5bbf67cb30c1fcfbe",
			GasBudget:   "1000",
		})

		if err != nil {
			fmt.Println(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)
	})

	t.Run("test on unsafe_paySui", func(t *testing.T) {
		rsp, err := cli.PaySui(ctx, models.PaySuiRequest{
			Signer:      "0xb7f98d327f19f674347e1e40641408253142d6e7e5093a7c96eda8cdfd7d9bb5",
			SuiObjectId: []string{"0x2aceb239c7c04c43a2e39824a003291f8e2b1d6027324df5bbf67cb30c1fcfbe"},
			Recipient:   []string{"0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff"},
			Amount:      []string{"1"},
			GasBudget:   "1000",
		})

		if err != nil {
			fmt.Println(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)
	})

	t.Run("test on unsafe_payAllSui", func(t *testing.T) {
		rsp, err := cli.PayAllSui(ctx, models.PayAllSuiRequest{
			Signer:      "0xb7f98d327f19f674347e1e40641408253142d6e7e5093a7c96eda8cdfd7d9bb5",
			SuiObjectId: []string{"0x2aceb239c7c04c43a2e39824a003291f8e2b1d6027324df5bbf67cb30c1fcfbe"},
			Recipient:   "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
			GasBudget:   "1000",
		})

		if err != nil {
			fmt.Println(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)
	})

	t.Run("test on unsafe_batchTransaction", func(t *testing.T) {
		rsp, err := cli.BatchTransaction(ctx, models.BatchTransactionRequest{
			Signer: "0xb7f98d327f19f674347e1e40641408253142d6e7e5093a7c96eda8cdfd7d9bb5",
			RPCTransactionRequestParams: []models.RPCTransactionRequestParams{
				{
					MoveCallRequestParams: &models.MoveCallRequest{
						PackageObjectId: "0x7d584c9a27ca4a546e8203b005b0e9ae746c9bec6c8c3c0bc84611bcf4ceab5f",
						Module:          "auction",
						Function:        "start_an_auction",
						TypeArguments:   []interface{}{},
						Arguments: []interface{}{
							"0x342e959f8d9d1fa9327a05fd54fefd929bbedad47190bdbb58743d8ba3bd3420",
							"0x3fd0fdedb84cf1f59386b6251ba6dd2cb495094da26e0a5a38239acd9d437f96",
							"0xb3de4235cb04167b473de806d00ba351e5860500253cf8e62d711e578e1d92ae",
							"web3",
							"0x2aceb239c7c04c43a2e39824a003291f8e2b1d6027324df5bbf67cb30c1fcfbe",
						},
					},
				},
				{
					TransferObjectRequestParams: &models.TransferObjectRequest{
						ObjectId:  "0x2aceb239c7c04c43a2e39824a003291f8e2b1d6027324df5bbf67cb30c1fcfbe",
						Recipient: "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
					},
				},
			},
			Gas:                            "0x2aceb239c7c04c43a2e39824a003291f8e2b1d6027324df5bbf67cb30c1fcfbe",
			GasBudget:                      "1000",
			SuiTransactionBlockBuilderMode: "DevInspect",
		})

		if err != nil {
			fmt.Println(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)
	})


	t.Run("test on SignAndExecuteTransactionBlock", func(t *testing.T) {
		signerAccount, err := signer.NewSignertWithMnemonic("input your mnemonic")
		if err != nil {
			fmt.Println(err.Error())
			t.FailNow()
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
			t.FailNow()
		}

		// see the successful transaction url: https://suivision.xyz/txblock/C7iYsH4tU5RdY1KBeNax4mCBn3XLZ5UswsuDpKrVkcH6
		rsp2, err := cli.SignAndExecuteTransactionBlock(ctx, models.SignAndExecuteTransactionBlockRequest{
			TxnMetaData: rsp,
			PriKey:      priKey,
			Options: models.SuiTransactionBlockOptions{
				ShowInput:    true,
				ShowRawInput: true,
				ShowEffects:  true,
			},
			RequestType: "WaitForLocalExecution",
		})

		if err != nil {
			fmt.Println(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp2)
	})

}
*/
func TestOnReadMoveDataFromSui(t *testing.T) {
	t.Run("test on sui_getMoveFunctionArgTypes", func(t *testing.T) {
		rsp, err := cli.SuiGetMoveFunctionArgTypes(ctx, models.GetMoveFunctionArgTypesRequest{
			Package:  "0x9fe1780ac27ec50c9c441fb31822f5c148f841f09ee455c6a0daf7c634a30a27",
			Module:   "aifrens",
			Function: "claim",
		})

		if err != nil {
			fmt.Println(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)
	})

	t.Run("test on sui_getNormalizedMoveModulesByPackage", func(t *testing.T) {
		rsp, err := cli.SuiGetNormalizedMoveModulesByPackage(ctx, models.GetNormalizedMoveModulesByPackageRequest{
			Package: "0x9fe1780ac27ec50c9c441fb31822f5c148f841f09ee455c6a0daf7c634a30a27",
		})

		if err != nil {
			fmt.Println(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)
	})

	t.Run("test on sui_getNormalizedMoveModule", func(t *testing.T) {
		rsp, err := cli.SuiGetNormalizedMoveModule(ctx, models.GetNormalizedMoveModuleRequest{
			Package:    "0x9fe1780ac27ec50c9c441fb31822f5c148f841f09ee455c6a0daf7c634a30a27",
			ModuleName: "aifrens",
		})

		if err != nil {
			fmt.Println(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)
	})

	t.Run("test on sui_getNormalizedMoveStruct", func(t *testing.T) {
		rsp, err := cli.SuiGetNormalizedMoveStruct(ctx, models.GetNormalizedMoveStructRequest{
			Package:    "0x9fe1780ac27ec50c9c441fb31822f5c148f841f09ee455c6a0daf7c634a30a27",
			ModuleName: "aifrens",
			StructName: "AifrensPool",
		})

		if err != nil {
			fmt.Println(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)
	})

	t.Run("test on sui_getNormalizedMoveFunction", func(t *testing.T) {
		rsp, err := cli.SuiGetNormalizedMoveFunction(ctx, models.GetNormalizedMoveFunctionRequest{
			Package:      "0x9fe1780ac27ec50c9c441fb31822f5c148f841f09ee455c6a0daf7c634a30a27",
			ModuleName:   "aifrens",
			FunctionName: "claim",
		})

		if err != nil {
			fmt.Println(err.Error())
			t.FailNow()
		}

		utils.PrettyPrint(rsp)
	})

}
