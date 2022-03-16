package parse

import (
	"encoding/binary"

	"git.cplus.link/go/akit/util/decimal"
	"github.com/gagliardetto/solana-go/rpc"

	"git.cplus.link/crema/backend/pkg/domain"
)

// GetSwapPrice  获取swap价格（合约内部记录的为价格的开方，故返回时要返回值的平方）
func GetSwapPrice(account *rpc.GetAccountInfoResult, config *domain.SwapConfig) decimal.Decimal {
	dataLen := len(account.Value.Data.GetBinary())
	data := account.Value.Data.GetBinary()[dataLen-96 : dataLen-80]
	tmpPrice := PrecisionConversion(decimal.NewFromInt(int64(binary.LittleEndian.Uint64(data))), 12)
	tmpPrice = tmpPrice.Mul(tmpPrice)

	decimalPrecision := int(config.TokenA.Decimal - config.TokenB.Decimal)

	if decimalPrecision != 0 {
		tmpPrice = PrecisionConversion(tmpPrice, decimalPrecision)
	}

	return tmpPrice
}
