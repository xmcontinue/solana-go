package sol

import (
	"context"
	"math"
	"strconv"
	"time"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	"git.cplus.link/go/akit/util/decimal"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"

	model "git.cplus.link/crema/backend/internal/model/market"

	"git.cplus.link/crema/backend/pkg/domain"
)

const DAY_SECONDS float64 = 86400

type TVL struct {
	transactionCache map[string]*rpc.TransactionWithMeta
	signatureList    []*rpc.TransactionSignature
	tokenAVolume     uint64
	tokenBVolume     uint64
	// client           *rpc.Client
	tokenABalance uint64
	tokenBBalance uint64
	txNum         uint64
	util          *solana.Signature
	*SwapConfig
}

type SwapConfig struct {
	Name          string `json:"name" mapstructure:"name"`
	Fee           string `json:"fee" mapstructure:"fee"`
	SwapAccount   string `json:"swap_account" mapstructure:"swap_account"`
	SwapPublicKey solana.PublicKey
	TokenA        Token `json:"token_a" mapstructure:"token_a"`
	TokenB        Token `json:"token_b" mapstructure:"token_b"`
}

type Token struct {
	Symbol             string `json:"symbol" mapstructure:"symbol"`
	TokenMint          string `json:"token_mint" mapstructure:"token_mint"`
	SwapTokenAccount   string `json:"swap_token_account" mapstructure:"swap_token_account"`
	SwapTokenPublicKey solana.PublicKey
	Decimal            uint8 `json:"decimal" mapstructure:"decimal"`
}

// func Init(config *config.Config) error {
// 	var rErr error
// 	once.Do(func() {
// 		stopChan := make(chan struct{})
// 		resChan, err := etcd.Watch("/crema/swap-pairs", stopChan)
// 		if err != nil {
// 			rErr = errors.Wrap(err)
// 			return
// 		}
//
// 		go func() {
// 			for {
// 				select {
// 				case res := <-resChan:
// 					err = json.Unmarshal(res.Value, &swapConfigList)
// 					if err != nil {
// 						rErr = errors.Wrap(err)
// 						return
// 					}
//
// 					// 加载配置
// 					for _, v := range swapConfigList {
// 						v.SwapPublicKey = solana.MustPublicKeyFromBase58(v.SwapAccount)
// 						v.TokenA.SwapTokenPublicKey = solana.MustPublicKeyFromBase58(v.TokenA.SwapTokenAccount)
// 						v.TokenB.SwapTokenPublicKey = solana.MustPublicKeyFromBase58(v.TokenB.SwapTokenAccount)
// 					}
// 				}
// 			}
// 		}()
//
// 		time.Sleep(time.Second) // todo
//
// 		err = config.UnmarshalKey("chain_net_rpc", &chainNetRpc)
// 		if err != nil {
// 			rErr = errors.Wrap(err)
// 			return
// 		}
//
// 	})
// 	return rErr
// }

func NewTVL(swapConfig *SwapConfig) *TVL {
	// TODO 若重启时是否由数据库中读取last transaction至 tvl中
	// net := rpc.MainNetBeta_RPC
	// if chainNet == "dev" {
	//	//net = rpc.DevNet_RPC
	// }

	// chainNetRpc := "https://connect.runnode.com/?apikey=PMkQIG6CxY0ybWmaHRHJ"

	return &TVL{
		transactionCache: make(map[string]*rpc.TransactionWithMeta),
		signatureList:    make([]*rpc.TransactionSignature, 0),
		tokenAVolume:     0,
		tokenBVolume:     0,
		SwapConfig:       swapConfig,
	}
}

func (tvl *TVL) work() error {
	if len(tvl.signatureList) != 0 {
		tvl.util = &tvl.signatureList[0].Signature
	} else {
		tvl.util = nil
	}

	logger.Info("tvl sync: pullLastSignature", logger.String("swap_address:", tvl.SwapAccount))
	tvl.pullLastSignature()

	logger.Info("tvl sync: removeOldSignature", logger.String("swap_address:", tvl.SwapAccount))
	tvl.removeOldSignature()

	logger.Info("tvl sync: calculate", logger.String("swap_address:", tvl.SwapAccount))
	tvl.calculate()

	logger.Info("tvl sync: getTvl", logger.String("swap_address:", tvl.SwapAccount))
	err := tvl.getTvl()

	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}
func (tvl *TVL) Start() error {
	logger.Info("tvl syncing ......", logger.String("swap_address:", tvl.SwapAccount))
	tvl.tokenAVolume = 0
	tvl.tokenBVolume = 0
	err := tvl.work()

	if err != nil {
		logger.Error("tvl sync fail ......", logger.Errorv(err))
		return errors.Wrap(err)
	}

	// 存入数据库
	// transactionsByte, _ := json.Marshal(tvl.transactionCache)
	// signaturesByte, _ := json.Marshal(tvl.signatureList)

	tokenAVolume, _ := decimal.NewFromString(strconv.FormatUint(tvl.tokenAVolume, 10))
	tokenBVolume, _ := decimal.NewFromString(strconv.FormatUint(tvl.tokenBVolume, 10))
	tokenABalance, _ := decimal.NewFromString(strconv.FormatUint(tvl.tokenABalance, 10))
	tokenBBalance, _ := decimal.NewFromString(strconv.FormatUint(tvl.tokenBBalance, 10))

	swapPairCount := &domain.SwapPairCount{
		TokenAVolume:      precisionConversion(tokenAVolume, int(tvl.TokenA.Decimal)),
		TokenBVolume:      precisionConversion(tokenBVolume, int(tvl.TokenB.Decimal)),
		TokenABalance:     precisionConversion(tokenABalance, int(tvl.TokenA.Decimal)),
		TokenBBalance:     precisionConversion(tokenBBalance, int(tvl.TokenB.Decimal)),
		TokenAPoolAddress: tvl.TokenA.SwapTokenAccount,
		TokenBPoolAddress: tvl.TokenB.SwapTokenAccount,
		TokenSwapAddress:  tvl.SwapAccount,
		// LastTransaction:   string(transactionsByte),
		// Signature:     string(signaturesByte),
		TxNum:         tvl.txNum,
		PairName:      tvl.Name,
		TokenASymbol:  tvl.TokenA.Symbol,
		TokenBSymbol:  tvl.TokenB.Symbol,
		TokenADecimal: int(tvl.TokenA.Decimal),
		TokenBDecimal: int(tvl.TokenB.Decimal),
	}

	err = model.CreateSwapPairCount(context.Background(), swapPairCount)
	if err != nil {
		logger.Error("tvl sync fail ......", logger.Errorv(err))
		return errors.Wrap(err)
	}

	logger.Info("tvl sync complete!", logger.String("swap_address:", tvl.SwapAccount))
	return nil
}

func (tvl *TVL) pullLastSignature() {
	limit := 1000
	var before *solana.Signature = nil
	var opts *rpc.GetSignaturesForAddressOpts
	pullResult := make([]*rpc.TransactionSignature, 0)
	for {
		if tvl.util == nil && before == nil {
			opts = &rpc.GetSignaturesForAddressOpts{
				Limit:      &limit,
				Commitment: rpc.CommitmentFinalized,
			}
		} else if tvl.util != nil && before == nil {
			opts = &rpc.GetSignaturesForAddressOpts{
				Limit:      &limit,
				Until:      *tvl.util,
				Commitment: rpc.CommitmentFinalized,
			}
		} else if tvl.util == nil && before != nil {
			opts = &rpc.GetSignaturesForAddressOpts{
				Limit:      &limit,
				Before:     *before,
				Commitment: rpc.CommitmentFinalized,
			}
		} else {
			opts = &rpc.GetSignaturesForAddressOpts{
				Limit:      &limit,
				Before:     *before,
				Until:      *tvl.util,
				Commitment: rpc.CommitmentFinalized,
			}
		}
		out, err := GetRpcClient().GetSignaturesForAddressWithOpts(
			context.TODO(),
			tvl.SwapPublicKey,
			opts,
		)
		if err != nil {
			break
		}
		size := len(out)
		if size < limit {
			pullResult = append(pullResult, out...)
			break
		}
		if tvl.util == nil {
			blockTime := out[size-1].BlockTime
			currentTime := time.Now()
			duration := currentTime.Sub(blockTime.Time())
			flapsSeconds := duration.Seconds()
			if flapsSeconds > DAY_SECONDS {
				pullResult = append(pullResult, out[0:size]...)
				break
			} else {
				before = &(out[size-1].Signature)
				pullResult = append(pullResult, out[0:size]...)
				continue
			}
		} else {
			before = &(out[size-1].Signature)
			pullResult = append(pullResult, out[0:(size)]...)
		}

	}
	totalSize := len(pullResult)
	validPullResult := make([]*rpc.TransactionSignature, 0)

	for _, v := range pullResult {
		if v.Err == nil {
			validPullResult = append(validPullResult, v)
		}
	}
	logger.Info("tvl sync: signature size", logger.Int("total size:", totalSize), logger.Int("valid size:", len(validPullResult)), logger.String("swap_address:", tvl.SwapAccount))
	finalResult := make([]*rpc.TransactionSignature, 0)
	for k, value := range validPullResult {
		if k%100 == 0 && k > 0 {
			logger.Info("tvl sync: transaction num", logger.Int("now size:", k), logger.Int("valid size:", len(validPullResult)), logger.String("swap_address:", tvl.SwapAccount))
		}
		out, err := GetRpcClient().GetConfirmedTransaction(
			context.TODO(),
			value.Signature,
		)
		if err != nil || out.Meta.Err != nil {
			continue
		}
		key := value.Signature.String()

		instructionLen := getInstructionLen(out.Transaction.Message.Instructions)
		if instructionLen == 9 || instructionLen == 41 || instructionLen == 50 || instructionLen == 52 {
			continue
		}
		
		tvl.transactionCache[key] = out
		finalResult = append(finalResult, value)
	}
	tvl.signatureList = append(finalResult, tvl.signatureList...)
}

func (tvl *TVL) removeOldSignature() {
	currentTime := time.Now()
	end := 0
	for index, value := range tvl.signatureList {
		blockTime := value.BlockTime.Time()
		duration := currentTime.Sub(blockTime)
		flapsSeconds := duration.Seconds()
		if flapsSeconds > DAY_SECONDS {
			delete(tvl.transactionCache, value.Signature.String())
		} else {
			end = index + 1
		}
	}
	tvl.signatureList = tvl.signatureList[0:end]
}

func (tvl *TVL) calculate() {
	tvl.tokenAVolume = 0
	tvl.tokenBVolume = 0
	tvl.txNum = 0

	for _, meta := range tvl.transactionCache {
		tokenAVolumeTmp, tokenBVolumeTmp := tvl.getSwapVolume(meta, tvl.TokenA.SwapTokenPublicKey, tvl.TokenB.SwapTokenPublicKey)
		if tokenAVolumeTmp < 0 || tokenBVolumeTmp < 0 {
			tvl.txNum += 1
		}
		if tokenAVolumeTmp < 0 {
			tvl.tokenAVolume = tvl.tokenAVolume + uint64(abs(tokenAVolumeTmp))
		}
		if tokenBVolumeTmp < 0 {
			tvl.tokenBVolume = tvl.tokenBVolume + uint64(abs(tokenBVolumeTmp))
		}
	}

	// for _, meta := range tvl.transactionCache {
	// 	tokenAVolumeTmp, tokenBVolumeTmp := tvl.getSwapVolume(meta, tvl.TokenA.SwapTokenPublicKey, tvl.TokenB.SwapTokenPublicKey)
	// 	tvl.tokenAVolume = tvl.tokenAVolume + uint64(abs(tokenAVolumeTmp))
	// 	tvl.tokenBVolume = tvl.tokenBVolume + uint64(abs(tokenBVolumeTmp))
	// }
}

func abs(n int64) int64 {
	y := n >> 63
	return (n ^ y) - y
}

// precisionConversion 精度转换
func precisionConversion(num decimal.Decimal, precision int) decimal.Decimal {
	return num.Div(decimal.NewFromFloat(math.Pow10(precision)))
}

func (tvl TVL) getSwapVolume(meta *rpc.TransactionWithMeta, tokenAPoolAddress solana.PublicKey, tokenBPoolAddress solana.PublicKey) (int64, int64) {
	var tokenAPreBalanceStr string
	var tokenBPreBalanceStr string
	var tokenAPostBalanceStr string
	var tokenBPostBalanceStr string

	for _, tokenBalance := range meta.Meta.PreTokenBalances {
		keyIndex := tokenBalance.AccountIndex
		key := meta.Transaction.Message.AccountKeys[keyIndex]
		if key.Equals(tokenAPoolAddress) {
			tokenAPreBalanceStr = tokenBalance.UiTokenAmount.Amount
			continue
		}
		if key.Equals(tokenBPoolAddress) {
			tokenBPreBalanceStr = tokenBalance.UiTokenAmount.Amount
			continue
		}
	}

	for _, tokenBalance := range meta.Meta.PostTokenBalances {
		keyIndex := tokenBalance.AccountIndex
		key := meta.Transaction.Message.AccountKeys[keyIndex]
		if key.Equals(tokenAPoolAddress) {
			tokenAPostBalanceStr = tokenBalance.UiTokenAmount.Amount
			continue
		}
		if key.Equals(tokenBPoolAddress) {
			tokenBPostBalanceStr = tokenBalance.UiTokenAmount.Amount
			continue
		}
	}

	tokenAPreBalance, _ := strconv.ParseInt(tokenAPreBalanceStr, 10, 64)
	tokenAPostBalance, _ := strconv.ParseInt(tokenAPostBalanceStr, 10, 64)
	tokenBPreBalance, _ := strconv.ParseInt(tokenBPreBalanceStr, 10, 64)
	tokenBPostBalance, _ := strconv.ParseInt(tokenBPostBalanceStr, 10, 64)
	tokenADeltaVolume := tokenAPostBalance - tokenAPreBalance
	tokenBDeltaVolume := tokenBPostBalance - tokenBPreBalance
	return tokenADeltaVolume, tokenBDeltaVolume
}

func (tvl *TVL) getTvl() error {
	resp, err := GetRpcClient().GetAccountInfo(
		context.TODO(),
		tvl.TokenA.SwapTokenPublicKey,
	)
	if err != nil {
		return errors.Wrap(err)
	}
	var tokenA token.Account
	// Account{}.Data.GetBinary() returns the *decoded* binary data
	// regardless the original encoding (it can handle them all).
	err = bin.NewBinDecoder(resp.Value.Data.GetBinary()).Decode(&tokenA)
	if err != nil {
		return errors.Wrap(err)
	}
	tvl.tokenABalance = tokenA.Amount
	resp, err = GetRpcClient().GetAccountInfo(
		context.TODO(),
		tvl.TokenB.SwapTokenPublicKey,
	)
	if err != nil {
		return errors.Wrap(err)
	}
	var tokenB token.Account
	// Account{}.Data.GetBinary() returns the *decoded* binary data
	// regardless the original encoding (it can handle them all).
	err = bin.NewBinDecoder(resp.Value.Data.GetBinary()).Decode(&tokenB)
	if err != nil {
		return errors.Wrap(err)
	}
	tvl.tokenBBalance = tokenB.Amount

	return nil
}

func SwapConfigList() []*SwapConfig {
	return swapConfigList
}

// getInstructionLen 获取第一个Instruction data长度
func getInstructionLen(instructions []solana.CompiledInstruction) uint64 {
	for _, v := range instructions {
		dataLen := len(v.Data)
		if dataLen > 0 {
			return uint64(dataLen)
		}
	}
	return 0
}
