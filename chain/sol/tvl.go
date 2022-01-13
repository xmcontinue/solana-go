package sol

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"git.cplus.link/go/akit/config"
	"git.cplus.link/go/akit/errors"
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
	client           *rpc.Client
	tokenABalance    uint64
	tokenBBalance    uint64
	util             *solana.Signature
	PublicKey
}

type Address struct {
	TokenAPoolAddress string `json:"tokenAPoolAddress" mapstructure:"tokenAPoolAddress"`
	TokenBPoolAddress string `json:"tokenBPoolAddress" mapstructure:"tokenBPoolAddress"`
	TokenSwapAddress  string `json:"tokenSwapAddress" mapstructure:"tokenSwapAddress"`
}

type PublicKey struct {
	TokenAPoolAddress solana.PublicKey
	TokenBPoolAddress solana.PublicKey
	TokenSwapAddress  solana.PublicKey
}

var (
	once       sync.Once
	addresses  []Address
	publicKeys []PublicKey
	chainNet   string
)

func Init(config *config.Config) error {
	var rErr error
	once.Do(func() {
		err := config.UnmarshalKey("tvl_address", &addresses)
		if err != nil {
			rErr = errors.Wrap(err)
			return
		}
		err = config.UnmarshalKey("chain_net", &chainNet)
		if err != nil {
			rErr = errors.Wrap(err)
			return
		}
		// 加载配置
		for _, v := range addresses {
			publicKeys = append(publicKeys, PublicKey{
				solana.MustPublicKeyFromBase58(v.TokenAPoolAddress),
				solana.MustPublicKeyFromBase58(v.TokenBPoolAddress),
				solana.MustPublicKeyFromBase58(v.TokenSwapAddress),
			})
		}
	})
	return rErr
}

func NewTVL(publicKey PublicKey) *TVL {

	// TODO 若重启时是否由数据库中读取last transaction至 tvl中
	net := rpc.MainNetBeta_RPC
	if chainNet == "dev" {
		net = rpc.DevNet_RPC
	}

	return &TVL{
		transactionCache: make(map[string]*rpc.TransactionWithMeta),
		signatureList:    make([]*rpc.TransactionSignature, 0),
		tokenAVolume:     0,
		tokenBVolume:     0,
		client:           rpc.New(net),
		PublicKey:        publicKey,
	}
}

func (tvl *TVL) work() error {
	if len(tvl.signatureList) != 0 {
		tvl.util = &tvl.signatureList[0].Signature
	} else {
		tvl.util = nil
	}
	tvl.pullLastSignature()
	tvl.removeOldSignature()
	tvl.calculate()
	err := tvl.getTvl()

	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}
func (tvl *TVL) Start() error {
	err := tvl.work()

	if err != nil {
		return errors.Wrap(err)
	}

	// 存入数据库
	transactionsByte, _ := json.Marshal(tvl.transactionCache)
	signaturesByte, _ := json.Marshal(tvl.signatureList)

	swapPairCount := &domain.SwapPairCount{
		TokenAVolume:      tvl.tokenAVolume,
		TokenBVolume:      tvl.tokenBVolume,
		TokenABalance:     tvl.tokenABalance,
		TokenBBalance:     tvl.tokenBBalance,
		TokenAPoolAddress: tvl.TokenAPoolAddress.String(),
		TokenBPoolAddress: tvl.TokenBPoolAddress.String(),
		TokenSwapAddress:  tvl.TokenSwapAddress.String(),
		LastTransaction:   string(transactionsByte),
		Signature:         string(signaturesByte),
	}

	err = model.CreateSwapPairCount(context.Background(), swapPairCount)
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}

func (tvl *TVL) pullLastSignature() {
	limit := 10
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
		out, err := tvl.client.GetSignaturesForAddressWithOpts(
			context.TODO(),
			tvl.TokenSwapAddress,
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
	finalResult := make([]*rpc.TransactionSignature, 0)
	for _, value := range pullResult {
		out, err := tvl.client.GetConfirmedTransaction(
			context.TODO(),
			value.Signature,
		)
		if err != nil || out.Meta.Err != nil {
			continue
		}
		key := value.Signature.String()
		if len(out.Transaction.Message.Instructions[0].Data) != 17 {
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
	for _, meta := range tvl.transactionCache {
		tokenAVolumeTmp, tokenBVolumeTmp := tvl.getSwapVolume(meta, tvl.TokenAPoolAddress, tvl.TokenBPoolAddress)
		tvl.tokenAVolume = tvl.tokenAVolume + uint64(tokenAVolumeTmp)
		tvl.tokenBVolume = tvl.tokenBVolume + uint64(tokenBVolumeTmp)
	}
}

func (tvl TVL) getSwapVolume(meta *rpc.TransactionWithMeta, tokenAPoolAddress solana.PublicKey, tokenBPoolAddress solana.PublicKey) (int, int) {
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

	tokenAPreBalance, _ := strconv.Atoi(tokenAPreBalanceStr)
	tokenAPostBalance, _ := strconv.Atoi(tokenAPostBalanceStr)
	tokenBPreBalance, _ := strconv.Atoi(tokenBPreBalanceStr)
	tokenBPostBalance, _ := strconv.Atoi(tokenBPostBalanceStr)
	tokenADeltaVolume := tokenAPostBalance - tokenAPreBalance
	tokenBDeltaVolume := tokenBPostBalance - tokenBPreBalance
	return tokenADeltaVolume, tokenBDeltaVolume
}

func (tvl *TVL) getTvl() error {
	resp, err := tvl.client.GetAccountInfo(
		context.TODO(),
		tvl.TokenAPoolAddress,
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
	resp, err = tvl.client.GetAccountInfo(
		context.TODO(),
		tvl.TokenBPoolAddress,
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

func PublicKeys() []PublicKey {
	return publicKeys
}

func Addresses() []Address {
	return addresses
}
