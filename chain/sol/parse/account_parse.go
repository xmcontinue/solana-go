package parse

import (
	"encoding/binary"

	"git.cplus.link/go/akit/util/decimal"
	"github.com/gagliardetto/solana-go/rpc"
)

// GetSwapPrice  获取swap价格（合约内部记录的为价格的开方，故返回时要返回值的平方）
func GetSwapPrice(account *rpc.GetAccountInfoResult) decimal.Decimal {
	dataLen := len(account.Value.Data.GetBinary())
	data := account.Value.Data.GetBinary()[dataLen-96 : dataLen-80]
	contractPrice := PrecisionConversion(decimal.NewFromInt(int64(binary.LittleEndian.Uint64(data))), 12)
	return contractPrice.Mul(contractPrice)
}
