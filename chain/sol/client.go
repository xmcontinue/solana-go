package sol

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"git.cplus.link/go/akit/config"
	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	"git.cplus.link/go/akit/util/decimal"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"

	"git.cplus.link/crema/backend/internal/etcd"
)

const (
	etcdSwapPairsKey = "/crema/swap-pairs"
	chainNetRpcKey   = "chain_net_rpc"
)

type ChainNet struct {
	Address string
	Client  *rpc.Client
	Height  uint64
	Status  uint64
}

var (
	chainNet        *ChainNet
	chainNets       []*ChainNet
	chainNetsConfig []string
	swapConfigList  []*SwapConfig
	swapConfigMap   map[string]*SwapConfig
	once            sync.Once
)

func Init(config *config.Config) error {
	var rErr error
	once.Do(func() {
		stopChan := make(chan struct{})
		resChan, err := etcd.Watch(etcdSwapPairsKey, stopChan)
		if err != nil {
			rErr = errors.Wrap(err)
			return
		}

		go func() {
			for {
				select {
				case res := <-resChan:
					err = json.Unmarshal(res.Value, &swapConfigList)
					if err != nil {
						rErr = errors.Wrap(err)
						return
					}

					swapMap := make(map[string]*SwapConfig, len(swapConfigList))

					// 加载配置
					for _, v := range swapConfigList {
						v.SwapPublicKey = solana.MustPublicKeyFromBase58(v.SwapAccount)
						v.TokenA.SwapTokenPublicKey = solana.MustPublicKeyFromBase58(v.TokenA.SwapTokenAccount)
						v.TokenB.SwapTokenPublicKey = solana.MustPublicKeyFromBase58(v.TokenB.SwapTokenAccount)
						swapMap[v.SwapAccount] = v
					}

					swapConfigMap = swapMap
				}
			}
		}()

		time.Sleep(time.Second) // todo

		err = config.UnmarshalKey(chainNetRpcKey, &chainNetsConfig)
		if err != nil {
			rErr = errors.Wrap(err)
			return
		}

		if len(chainNetsConfig) == 0 {
			rErr = errors.New("chain net rpc address is not found!")
			return
		}

		for _, v := range chainNetsConfig {
			chainNets = append(chainNets, &ChainNet{
				Client:  rpc.New(v),
				Address: v,
			})
		}

		chainNet = chainNets[0]

		go WatchNet()
	})
	return rErr
}

// WatchNet 监测网络
func WatchNet() {
	for {
		WatchBlockHeight()

		CheckNet()

		logger.Info(fmt.Sprintf("chain net block height is %d", chainNet.Height))

		time.Sleep(time.Minute)
	}
}

// CheckNet 检查当前网络
func CheckNet() {
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

// WatchBlockHeight 监测区块高度,若落后则切换
func WatchBlockHeight() {
	// 获取最新区块高度
	for _, v := range chainNets {
		height, err := GetBlockHeightForClient(v.Client)
		if err != nil {
			continue
		}
		v.Height = height
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

// precisionConversion 精度转换
func precisionConversion(num decimal.Decimal, precision int) decimal.Decimal {
	return num.Div(decimal.NewFromFloat(math.Pow10(precision)))
}
