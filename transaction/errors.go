package transaction

import "errors"

var (
	ErrSignerNotSet         = errors.New("signer not set")
	ErrSenderNotSet         = errors.New("sender not set")
	ErrSuiClientNotSet      = errors.New("sui client not set")
	ErrGasDataNotAllSet     = errors.New("gas data not all set")
	ErrGasPriceNotSet       = errors.New("gas price not set: call SetGasPrice or attach a SuiClient")
	ErrInvalidSuiAddress    = errors.New("invalid sui address")
	ErrInvalidObjectId      = errors.New("invalid object id")
	ErrObjectNotSupportType = errors.New("object not support type")
)
