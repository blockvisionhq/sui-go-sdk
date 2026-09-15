package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/block-vision/sui-go-sdk/constant"
	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/signer"
	"github.com/block-vision/sui-go-sdk/sui"
	"github.com/block-vision/sui-go-sdk/utils"
)

var ctx = context.Background()
var cli = sui.NewSuiClient(constant.BvTestnetEndpoint)

func main() {
	//SuiGetTotalTransactionBlocks()
	SuiGetTransactionBlock()
	//SuiMultiGetTransactionBlocks()
	//SuiXQueryTransactionBlocks()
	//MoveCall()
	//MergeCoins()
	//SplitCoin()
	//SplitCoinEqual()
	//Publish()
	//TransferObject()
	//TransferSui()
	//Pay()
	//PaySui()
	//PayAllSui()
	//RequestAddStake()
	//RequestWithdrawStake()
	//BatchTransaction()
	//SuiExecuteTransactionBlock()
	//SuiDryRunTransactionBlock()
	//SuiDevInspectTransactionBlock()
	//SignAndExecuteTransactionBlock()
}

func SuiExecuteTransactionBlock() {
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
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp)
}

func SuiDryRunTransactionBlock() {
	rsp, err := cli.SuiDryRunTransactionBlock(ctx, models.SuiDryRunTransactionBlockRequest{
		TxBytes: "AAACAQEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABQEAAAAAAAAAAQEAjgDW4hJZlqvw654RGR3SdndKkdjoC0pzXQLxja/NUahLowQAAAAAACBEQGwClI9RQX68dzbN7PN29/Pw/Sc1hbtZwNAny7wZ+wEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAMKc3VpX3N5c3RlbRZyZXF1ZXN0X3dpdGhkcmF3X3N0YWtlAAIBAAABAQC3+Y0yfxn2dDR+HkBkFAglMULW5+UJOnyW7ajN/X2btQEqzrI5x8BMQ6LjmCSgAykfjisdYCcyTfW79nyzDB/PvtZBpwAAAAAAIAm+IREDziwoZLm7lc4ZKegZ2J5viEgoss9zgrFkHLh6t/mNMn8Z9nQ0fh5AZBQIJTFC1uflCTp8lu2ozf19m7XoAwAAAAAAAFDhjyoAAAAAAA==",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp)
}

func SuiDevInspectTransactionBlock() {
	type moveCallResult struct {
		ReturnValues [2]interface{}
	}
	rsp, err := cli.SuiDevInspectTransactionBlock(ctx, models.SuiDevInspectTransactionBlockRequest{
		Sender:  "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
		TxBytes: "AAUBAFIMicbHjFZu7Q6/JPhUqMItj90GpvFq0B8Qja1/G6rqFSjPCgAAAAAgU+6E0DJf97dwaRV0MdseKxCPU97monpYtyfJSa7jKQAACKCGAQAAAAAAAAiAlpgAAAAAAAABAQEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAYtKc8KAAAAACBhjSVWnhYSnxbZuJWgDlXi2j3pp7gqJJizx29rU/Ly+QEAy/R0ipZdRp6jo2zwzMV0O5bC0K5t7gdi7T7KZfrAf34EcG9vbBBnZXRfbGV2ZWwyX3JhbmdlAgcAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAgNzdWkDU1VJAAf3FSwFkwSAzXQNcxG1uLRcb0iOOlOhHD90pvrDalLg1wZEQlVTREMGREJVU0RDAAUBAAABAQABAgABAwABBAA=",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	var moveCallReturn []moveCallResult
	err = json.Unmarshal(rsp.Results, &moveCallReturn)
	if err != nil {
		fmt.Println("json", err.Error())
		return
	}
	utils.PrettyPrint(rsp)
}

// send signed transactions
func SignAndExecuteTransactionBlock() {
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
		Recipient:   "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
		Amount:      "1",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
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
		return
	}

	utils.PrettyPrint(rsp2)
}

func SuiGetTotalTransactionBlocks() {
	rsp, err := cli.SuiGetTotalTransactionBlocks(ctx)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp)
}

func SuiGetTransactionBlock() {
	rsp, err := cli.SuiGetTransactionBlock(ctx, models.SuiGetTransactionBlockRequest{
		Digest: "DjubVWzkndWDTzcWVd6g3xbQsfG3H2Gt7vMfrm1YjLNJ",
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
}

func SuiMultiGetTransactionBlocks() {
	rsp, err := cli.SuiMultiGetTransactionBlocks(ctx, models.SuiMultiGetTransactionBlocksRequest{
		Digests: []string{"CeVpDXKKU3Gs89efej9pKiYYQyTzifE2BDxWwquUaUht", "C2zZu2dpX2sLQy2234yt6ecRiNTVgQTXeQpgw9GhxGgo"},
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

	for _, transactionBlock := range rsp {
		utils.PrettyPrint(*transactionBlock)
	}
}

func SuiXQueryTransactionBlocks() {
	rsp, err := cli.SuiXQueryTransactionBlocks(ctx, models.SuiXQueryTransactionBlocksRequest{
		SuiTransactionBlockResponseQuery: models.SuiTransactionBlockResponseQuery{
			TransactionFilter: models.TransactionFilter{
				"FromAddress": "0x7d20dcdb2bca4f508ea9613994683eb4e76e9c4ed371169677c1be02aaf0b58e",
			},
			Options: models.SuiTransactionBlockOptions{
				ShowInput:          true,
				ShowRawInput:       true,
				ShowEffects:        true,
				ShowEvents:         true,
				ShowObjectChanges:  true,
				ShowBalanceChanges: true,
			},
		},
		Limit:           5,
		DescendingOrder: false,
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for _, transactionBlock := range rsp.Data {
		utils.PrettyPrint(transactionBlock)
	}
}

func MoveCall() {
	gasObj := "0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1"

	rsp, err := cli.MoveCall(ctx, models.MoveCallRequest{
		Signer:          "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
		PackageObjectId: "0x7d584c9a27ca4a546e8203b005b0e9ae746c9bec6c8c3c0bc84611bcf4ceab5f",
		Module:          "auction",
		Function:        "start_an_auction",
		TypeArguments:   []interface{}{},
		Arguments: []interface{}{
			"0x342e959f8d9d1fa9327a05fd54fefd929bbedad47190bdbb58743d8ba3bd3420",
			"0x3fd0fdedb84cf1f59386b6251ba6dd2cb495094da26e0a5a38239acd9d437f96",
			"0xb3de4235cb04167b473de806d00ba351e5860500253cf8e62d711e578e1d92ae",
			"web3",
			"0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1",
		},
		Gas:       &gasObj,
		GasBudget: "1000",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp)
}

func MergeCoins() {
	gasObj := "0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1"

	rsp, err := cli.MergeCoins(ctx, models.MergeCoinsRequest{
		Signer:      "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
		PrimaryCoin: "0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1",
		CoinToMerge: "0x92f03fdec6e0278dcb6fa3f4467eeee3e0bee1ac41825351ef53431677d2e2f7",
		Gas:         &gasObj,
		GasBudget:   "1000000",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp)
}

func SplitCoin() {
	gasObj := "0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1"

	rsp, err := cli.SplitCoin(ctx, models.SplitCoinRequest{
		Signer:       "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
		CoinObjectId: "0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1",
		SplitAmounts: []string{"1000", "1000"},
		Gas:          &gasObj,
		GasBudget:    "1000",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp)
}

func SplitCoinEqual() {
	gasObj := "0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1"

	rsp, err := cli.SplitCoinEqual(ctx, models.SplitCoinEqualRequest{
		Signer:       "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
		CoinObjectId: "0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1",
		SplitCount:   "2",
		Gas:          &gasObj,
		GasBudget:    "1000",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp)
}

func Publish() {
	gasObj := "0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1"

	rsp, err := cli.Publish(ctx, models.PublishRequest{
		Sender: "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
		CompiledModules: []string{
			"oRzrCwUAAAALAQAOAg4kAzJZBIsBFgWhAZoBB7sC5wEIogQoBsoECgrUBBoM7gSjAQ2RBgQAAAEBAQIBAwEEAQUBBgAHCAAACAgAAAkIAAIKDAEAAQQLAgAGDAIAAxIEAAEUBAEAAQANAAEAAA4CAQAADwMBAAAQBAEAAhUHCAEAAhYJCgEAARcLDAEAARgNCAEAAxkEDgAGGg8QAAUFEQEBCAEVEwgBAAIbFBUBAAMcDgEAAR0BDAEABR4ZAQEIBAYFBgYGBwYKAwsGDAYKFgoXDgYPGAMHCAEHCwMBCAQHCAUAAwYIAgcIAQcIBQEIAAEHCAUCBwsHAQgECwcBCAQBCAQBBgsDAQkAAQMBBwsDAQkAAQcLBwEJAAIHCwcBCQADAQsHAQkAAgcLBwEJAAsHAQkAAQgGAQYIBQEFAgkABQIDCwMBCAQBBgsHAQkAAwcLBwEJAAMHCAUBCwMBCQABCwMBCAQBCAIBCAEBCQAGZG9udXRzB2JhbGFuY2UEY29pbgZvYmplY3QDc3VpCHRyYW5zZmVyCnR4X2NvbnRleHQFRG9udXQJRG9udXRTaG9wDFNob3BPd25lckNhcARDb2luA1NVSQlUeENvbnRleHQJYnV5X2RvbnV0D2NvbGxlY3RfcHJvZml0cwllYXRfZG9udXQEaW5pdAJpZANVSUQFcHJpY2UHQmFsYW5jZQV2YWx1ZQtiYWxhbmNlX211dAVzcGxpdARqb2luA25ldwZzZW5kZXIEdGFrZQZkZWxldGUEemVybwxzaGFyZV9vYmplY3QAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACAwgAAAAAAAAAAAACAREIBgECAxEIBhMDAQsHAQgEAgIBEQgGAAEEAAUmCgEuOAAKABAAFCYDEAsAAQsBAQsCAQcAJwsBOAEMAwsDCgAQABQ4AgwECwAPAQsEOAMBCgIRCBIACwIuEQk4BAIBAQQAEhAKARABOAUMAwsBDwELAwoCOAYMBAsECwIuEQk4BwICAQQADgYLABMADAELARENAgMAAAABDgoAEQgSAgoALhEJOAgLABEIBugDAAAAAAAAOAkSATgKAgEBAQIA",
		},
		Dependencies: []string{"0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1"},
		Gas:          &gasObj,
		GasBudget:    "1000",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp)
}

func TransferObject() {
	gasObj := "0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1"

	rsp, err := cli.TransferObject(ctx, models.TransferObjectRequest{
		Signer:    "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
		ObjectId:  "0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1",
		Gas:       &gasObj,
		GasBudget: "1000",
		Recipient: "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp)
}

func TransferSui() {
	rsp, err := cli.TransferSui(ctx, models.TransferSuiRequest{
		Signer:      "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
		SuiObjectId: "0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1",
		GasBudget:   "1000000",
		Recipient:   "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
		Amount:      "1",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp)
}

func Pay() {
	gasObj := "0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1"

	rsp, err := cli.Pay(ctx, models.PayRequest{
		Signer:      "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
		SuiObjectId: []string{"0x92f03fdec6e0278dcb6fa3f4467eeee3e0bee1ac41825351ef53431677d2e2f7"},
		Recipient:   []string{"0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff"},
		Amount:      []string{"1"},
		Gas:         &gasObj,
		GasBudget:   "1000000",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp)
}

func PaySui() {
	rsp, err := cli.PaySui(ctx, models.PaySuiRequest{
		Signer:      "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
		SuiObjectId: []string{"0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1"},
		Recipient:   []string{"0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff"},
		Amount:      []string{"1"},
		GasBudget:   "1000000",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp)
}

func PayAllSui() {
	rsp, err := cli.PayAllSui(ctx, models.PayAllSuiRequest{
		Signer:      "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
		SuiObjectId: []string{"0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1"},
		Recipient:   "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
		GasBudget:   "1000000",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp)
}

func RequestAddStake() {
	gasObj := "0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1"

	rsp, err := cli.RequestAddStake(ctx, models.AddStakeRequest{
		Signer:    "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
		Coins:     []string{"0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1"},
		Amount:    "1",
		Validator: "0x884515e99dab69c4c28662149db81ca563ed4c36e0c8ce44a58e40e25a0a64a1",
		Gas:       &gasObj,
		GasBudget: "1000",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp)
}

func RequestWithdrawStake() {
	gasObj := "0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1"

	rsp, err := cli.RequestWithdrawStake(ctx, models.WithdrawStakeRequest{
		Signer:         "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
		StakedObjectId: "0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1",
		Gas:            &gasObj,
		GasBudget:      "1000",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp)
}

func BatchTransaction() {
	gasObj := "0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1"

	rsp, err := cli.BatchTransaction(ctx, models.BatchTransactionRequest{
		Signer: "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
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
						"0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1",
					},
				},
			},
			{
				TransferObjectRequestParams: &models.TransferObjectRequest{
					ObjectId:  "0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1",
					Recipient: "0x4ae8be62692d1bbf892b657ee78a59954240ee0525f20a5b5687a70995cf0eff",
				},
			},
		},
		Gas:                            &gasObj,
		GasBudget:                      "1000",
		SuiTransactionBlockBuilderMode: "DevInspect",
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	utils.PrettyPrint(rsp)
}
