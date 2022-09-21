package parse

import (
	"encoding/binary"
	"math/big"

	"git.cplus.link/go/akit/util/decimal"
	ag_binary "github.com/gagliardetto/binary"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
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
		tmpPrice = PrecisionConversion(tmpPrice, -decimalPrecision)
	}

	return tmpPrice
}

// GetSwapPriceV2  获取swap价格（合约内部记录的为价格的开方，故返回时要返回值的平方）
func GetSwapPriceV2(account *rpc.GetAccountInfoResult, config *domain.SwapConfig) decimal.Decimal {
	c := &Clmmpool{}
	var err error

	err = bin.NewBinDecoder(account.Value.Data.GetBinary()[8:]).Decode(c)
	if err != nil {
		return decimal.Zero
	}
	baseFloat := big.NewFloat(1)

	currentSqrtPrice, _, _ := baseFloat.Parse(c.CurrentSqrtPrice.String(), 10)
	// 先除以2*pow(64)
	currentSqrtPrice.Quo(currentSqrtPrice, MaxInt64)
	// 再平方
	currentSqrtPrice.Mul(currentSqrtPrice, currentSqrtPrice)

	tmpPrice, err := decimal.NewFromString(currentSqrtPrice.String())
	if err != nil {
		return decimal.Zero
	}
	decimalPrecision := int(config.TokenA.Decimal) - int(config.TokenB.Decimal)

	if decimalPrecision != 0 {
		tmpPrice = PrecisionConversion(tmpPrice, -decimalPrecision)
	}
	return tmpPrice
}

type Clmmpool struct {
	ClmmConfig       solana.PublicKey
	TokenA           solana.PublicKey
	TokenB           solana.PublicKey
	TokenAVault      solana.PublicKey
	TokenBVault      solana.PublicKey
	TickSpacing      uint16
	TickSpacingSeed  uint16
	FeeRate          uint16
	ProtocolFeeRate  uint16 // todo 再最新版本中去掉了 结构体长度也要去掉两个字节
	Liquidity        ag_binary.Uint128
	CurrentSqrtPrice ag_binary.Uint128
	//CurrentTickIndex  int32
	//FeeGrowthGlobalA  decimalU128
	//FeeGrowthGlobalB  decimalU128
	//DeeProtocolTokenA uint64
	//DeeProtocolTokenB uint64
	//Bump              decimalU6412
}
