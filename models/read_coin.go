package models

type SuiXGetBalanceRequest struct {
	// the owner's Sui address
	Owner string `json:"owner"`
	// optional type name for the coin (e.g., 0x168da5bf1f48dafc111b0a488fa454aca95e0b5e::usdc::USDC), default to 0x2::sui::SUI if not specified.
	CoinType string `json:"coinType"`
}

type SuiXGetAllBalanceRequest struct {
	// the owner's Sui address
	Owner string `json:"owner"`
}

type CoinLockedBalance struct {
	EpochId int `json:"epochId"`
	Number  int `json:"number"`
}

type CoinBalanceResponse struct {
	CoinType              string            `json:"coinType"`
	CoinObjectCount       uint64            `json:"coinObjectCount"`
	TotalBalance          string            `json:"totalBalance"`
	LockedBalance         CoinLockedBalance `json:"lockedBalance"`
	FundsInAddressBalance string            `json:"fundsInAddressBalance"`
}

type CoinAllBalanceResponse []CoinBalanceResponse

type SuiXGetCoinsRequest struct {
	// the owner's Sui address
	Owner string `json:"owner"`
	// optional type name for the coin (e.g., 0x168da5bf1f48dafc111b0a488fa454aca95e0b5e::usdc::USDC), default to 0x2::sui::SUI if not specified.
	CoinType string `json:"coin_type"`
	// optional paging cursor
	Cursor interface{} `json:"cursor"`
	// maximum number of items per page
	Limit uint64 `json:"limit" validate:"lte=50"`
}

type PaginatedCoinsResponse struct {
	Data        []CoinData `json:"data"`
	NextCursor  string     `json:"nextCursor"`
	HasNextPage bool       `json:"hasNextPage"`
}

type CoinData struct {
	CoinType            string `json:"coinType"`
	CoinObjectId        string `json:"coinObjectId"`
	Version             string `json:"version"`
	Digest              string `json:"digest"`
	Balance             string `json:"balance"`
	LockedUntilEpoch    uint64 `json:"lockedUntilEpoch"`
	PreviousTransaction string `json:"previousTransaction"`
}

type SuiXGetAllCoinsRequest struct {
	// the owner's Sui address
	Owner string `json:"owner"`
	// optional paging cursor
	Cursor interface{} `json:"cursor"`
	// maximum number of items per page
	Limit uint64 `json:"limit" validate:"lte=50"`
}

type SuiXGetCoinMetadataRequest struct {
	// type name for the coin (e.g., 0x168da5bf1f48dafc111b0a488fa454aca95e0b5e::usdc::USDC)
	CoinType string `json:"coinType"`
}

type CoinMetadataResponse struct {
	Id          string `json:"id"`
	Decimals    int    `json:"decimals"`
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Description string `json:"description"`
	IconUrl     string `json:"iconUrl"`
}

type SuiXGetTotalSupplyRequest struct {
	// type name for the coin (e.g., 0x168da5bf1f48dafc111b0a488fa454aca95e0b5e::usdc::USDC)
	CoinType string `json:"coinType"`
}

type TotalSupplyResponse struct {
	Value string `json:"value"`
}
