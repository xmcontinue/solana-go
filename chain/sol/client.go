package sol

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"git.cplus.link/go/akit/config"
	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	"git.cplus.link/go/akit/util/decimal"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/rpcxio/libkv/store"

	"git.cplus.link/crema/backend/chain/sol/parse"
	"git.cplus.link/crema/backend/internal/etcd"
	"git.cplus.link/crema/backend/pkg/domain"
)

const (
	etcdSwapPairsKey = "/crema/swap-pairs"
	chainNetRpcKey   = "chain_net_rpc"
)

var (
	chainNet       *domain.ChainNet
	chainNets      []*domain.ChainNet
	swapConfigList []*domain.SwapConfig
	swapConfigMap  map[string]*domain.SwapConfig
	tokenConfigMap map[string]*domain.Token
	once           sync.Once
	wg             sync.WaitGroup
	isInit         bool
	configLock     sync.Mutex
)

func Init(conf *config.Config) error {
	var err error
	once.Do(func() {
		// 加载swap pairs配置
		stopChan := make(chan struct{})
		resChan, err := etcd.Watch(etcdSwapPairsKey, stopChan)
		if err != nil {
			return
		}

		wg.Add(1)
		go watchSwapPairsConfig(resChan)
		wg.Wait()

		// 加载网络配置
		err = initNet(conf)
		if err != nil {
			return
		}

		// Watch Balance
		wg.Add(1)
		go watchBalance()
		wg.Wait()

		// watchNet 监测网络
		go watchNet()
	})

	return errors.Wrap(err)
}

// watchSwapPairsConfig 监听swap pairs配置变动
func watchSwapPairsConfig(swapConfigChan <-chan *store.KVPair) {
	for {
		select {
		case res := <-swapConfigChan:
			configLock.Lock()

			err := json.Unmarshal(res.Value, &swapConfigList)
			if err != nil {
				logger.Error("swap config unmarshal failed :", logger.Errorv(err))
			}

			swapMap := make(map[string]*domain.SwapConfig, len(swapConfigList))
			tokenMap := make(map[string]*domain.Token, 0)

			// 加载配置
			for _, v := range swapConfigList {
				v.SwapPublicKey = solana.MustPublicKeyFromBase58(v.SwapAccount)
				v.TokenA.SwapTokenPublicKey = solana.MustPublicKeyFromBase58(v.TokenA.SwapTokenAccount)
				v.TokenB.SwapTokenPublicKey = solana.MustPublicKeyFromBase58(v.TokenB.SwapTokenAccount)
				v.TokenA.TokenMintPublicKey = solana.MustPublicKeyFromBase58(v.TokenA.TokenMint)
				v.TokenB.TokenMintPublicKey = solana.MustPublicKeyFromBase58(v.TokenB.TokenMint)
				swapMap[v.SwapAccount] = v
				tokenMap[v.TokenA.SwapTokenAccount] = &v.TokenA
				tokenMap[v.TokenB.SwapTokenAccount] = &v.TokenB
			}

			swapConfigMap = swapMap
			tokenConfigMap = tokenMap
			parse.SetSwapConfig(swapConfigMap)

			if !isInit {
				wg.Done()
			}

			configLock.Unlock()
		}
	}
}

// initNet 初始化网络
func initNet(conf *config.Config) error {
	chainNetsConfig := make([]string, 0)
	err := conf.UnmarshalKey(chainNetRpcKey, &chainNetsConfig)
	if err != nil {
		return errors.Wrap(err)
	}

	if len(chainNetsConfig) == 0 {
		return errors.New("chain net rpc address is not found!")
	}

	for _, v := range chainNetsConfig {
		chainNets = append(chainNets, &domain.ChainNet{
			Client:  rpc.New(v),
			Address: v,
		})
	}

	// 默认使用第一个网络
	chainNet = chainNets[0]

	return nil
}

// watchNet 监测网络
func watchNet() {
	for {
		// watchBlockHeight 监测区块高度,若落后则切换
		watchBlockHeight()
		// checkNet 检查当前网络
		checkNet()

		logger.Info(fmt.Sprintf("chain net block height is %d", chainNet.Height))

		time.Sleep(time.Minute)
	}
}

// checkNet 检查当前网络
func checkNet() {
	// 获取网络组中最高的区块高度
	var maxHeight uint64
	for _, v := range chainNets {
		if v.Height > maxHeight {
			maxHeight = v.Height
		}
	}
	// 判断是否当前区块有落后
	if maxHeight-chainNet.Height > 1000 {
		logger.Info(fmt.Sprintf("chain net rpc(%s) block height too low, max height is %d, now is %d ", chainNet.Address, maxHeight, chainNet.Height))
		// 替换为区块高度正常的rpc
		for _, v := range chainNets {
			if maxHeight-v.Height < 1000 && v.Address != chainNet.Address {
				logger.Info(fmt.Sprintf("chain net rpc has been switched from %s to %s", chainNet.Address, v.Address))
				chainNet = v
			}
		}
	}
}

// watchBlockHeight 监测区块高度,若落后则切换
func watchBlockHeight() {
	// 获取最新区块高度
	for _, v := range chainNets {
		height, err := GetBlockHeightForClient(v.Client)
		if err != nil {
			continue
		}
		v.Height = height
	}
}

// watchBalance 监听钱包余额变动
func watchBalance() {
	for {
		configLock.Lock()

		for _, v := range tokenConfigMap {
			resp, err := GetRpcClient().GetAccountInfo(
				context.TODO(),
				v.SwapTokenPublicKey,
			)
			if err != nil {
				return
			}
			var tokenA token.Account
			err = bin.NewBinDecoder(resp.Value.Data.GetBinary()).Decode(&tokenA)
			if err != nil {
				return
			}
			v.Balance = parse.PrecisionConversion(decimal.NewFromInt(int64(tokenA.Amount)), int(v.Decimal))
		}

		for _, v := range swapConfigList {
			v.TokenA.Balance = GetTokenForTokenAccount(v.TokenA.SwapTokenAccount).Balance
			v.TokenB.Balance = GetTokenForTokenAccount(v.TokenB.SwapTokenAccount).Balance
		}
		if !isInit {
			isInit = true
			wg.Done()
		}

		configLock.Unlock()
		time.Sleep(time.Second * 10)
	}
}

func GetBlockHeightForClient(rpcClient *rpc.Client) (uint64, error) {
	return rpcClient.GetBlockHeight(context.TODO(), rpc.CommitmentMax)
}

func GetRpcClient() *rpc.Client {
	return chainNet.Client
}

func abs(n int64) int64 {
	y := n >> 63
	return (n ^ y) - y
}

// GetTokenForTokenAccount 根据token account获取token配置
func GetTokenForTokenAccount(account string) *domain.Token {
	return tokenConfigMap[account]
}