package sol

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
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
	"github.com/gagliardetto/solana-go/rpc/jsonrpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/rpcxio/libkv/store"

	event "git.cplus.link/crema/backend/chain/event/activity"
	"git.cplus.link/crema/backend/chain/sol/parse"
	"git.cplus.link/crema/backend/internal/etcd"
	"git.cplus.link/crema/backend/pkg/crema"
	"git.cplus.link/crema/backend/pkg/domain"
)

const (
	chainNetRpcKey = "chain_net_rpc"
)

var (
	etcdSwapPairsKey   = "/swap-pairs"
	etcdTokenListKey   = "/token-list"
	etcdSwapPairsKeyV2 = "/v2-swap-pairs"
)

var (
	chainNet            *domain.ChainNet
	wsClient            *ws.Client
	wsNet               string
	chainNets           []*domain.ChainNet
	chainNetsConfig     []string
	swapConfigList      []*domain.SwapConfig
	swapConfigListV2    []*domain.SwapConfig
	swapConfigMap       map[string]*domain.SwapConfig
	tokenConfigMap      map[string]*domain.Token
	TokenList           []*domain.TokenInfo
	once                sync.Once
	wg                  sync.WaitGroup
	isInit              bool
	configLock          sync.Mutex
	activityEventParser event.EventParser
)

func Init(conf *config.Config) error {

	var err error
	once.Do(func() {

		// 获取programID
		err = conf.UnmarshalKey("program_address_crema_v2", &ProgramIDV2)
		if err != nil {
			panic(err.Error())
		}

		swapConfigMap = make(map[string]*domain.SwapConfig)
		tokenConfigMap = make(map[string]*domain.Token)

		err = conf.UnmarshalKey("solana_ws_net", &wsNet)
		if err != nil {
			panic(err.Error())
		}

		wsClient = newWSConnect()

		etcdSwapPairsKey = "/" + domain.GetPublicPrefix() + etcdSwapPairsKey
		// 加载swap pairs配置
		stopChan := make(chan struct{})
		resChan, err := etcd.Watch(etcdSwapPairsKey, stopChan)
		if err != nil {
			return
		}

		activityEventParser = event.NewActivityEventParser()

		wg.Add(1)
		go watchSwapPairsConfig(resChan, &swapConfigList, "")
		wg.Wait()

		etcdSwapPairsKeyV2 = "/" + domain.GetPublicPrefix() + etcdSwapPairsKeyV2

		stopChanV2 := make(chan struct{})
		resChanV2, err := etcd.Watch(etcdSwapPairsKeyV2, stopChanV2)
		if err != nil {
			return
		}

		wg.Add(1)
		go watchSwapPairsConfig(resChanV2, &swapConfigListV2, "v2")
		wg.Wait()

		etcdTokenListKey = "/" + domain.GetPublicPrefix() + etcdTokenListKey

		stopTokenListChanV2 := make(chan struct{})
		resTokenListChanV2, err := etcd.Watch(etcdTokenListKey, stopTokenListChanV2)
		if err != nil {
			return
		}

		wg.Add(1)
		go watchTokenListConfig(resTokenListChanV2)
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

// newWSConnect 新建ws连接
func newWSConnect() *ws.Client {
	wsCli, err := ws.Connect(context.Background(), wsNet)
	if err != nil {
		panic(err)
	}
	return wsCli
}

// watchTokenListConfig 监听token list配置变动
func watchTokenListConfig(swapConfigChan <-chan *store.KVPair) {
	for {
		select {
		case res := <-swapConfigChan:
			configLock.Lock()

			var err error

			tokenListTemp := make([]*domain.TokenInfo, 0)
			err = json.Unmarshal(res.Value, &tokenListTemp)
			if err != nil {
				logger.Error("swap config unmarshal failed :", logger.Errorv(err))
				return
			}
			TokenList = tokenListTemp

			if !isInit {
				wg.Done()
			}

			configLock.Unlock()
		}
	}
}

// watchSwapPairsConfig 监听swap pairs配置变动
func watchSwapPairsConfig(swapConfigChan <-chan *store.KVPair, swapConfigList1 *[]*domain.SwapConfig, version string) {
	for {
		select {
		case res := <-swapConfigChan:
			configLock.Lock()

			var err error

			swapConfigListTemp := make([]*domain.SwapConfig, 0)
			if version == "v2" {
				err = json.Unmarshal(res.Value, &swapConfigListTemp)
				if err != nil {
					logger.Error("swap config unmarshal failed :", logger.Errorv(err))
					return
				}
				swapConfigListV2 = swapConfigListTemp

			} else {

				err = json.Unmarshal(res.Value, &swapConfigListTemp)
				if err != nil {
					logger.Error("swap config unmarshal failed :", logger.Errorv(err))
					return
				}
				swapConfigList = swapConfigListTemp

			}

			swapMap := make(map[string]*domain.SwapConfig, len(swapConfigListTemp))
			tokenMap := make(map[string]*domain.Token, 0)

			// 加载配置
			for _, v := range swapConfigListTemp {
				v.SwapPublicKey = solana.MustPublicKeyFromBase58(v.SwapAccount)

				v.TokenA.TokenMintPublicKey = solana.MustPublicKeyFromBase58(v.TokenA.TokenMint)
				v.TokenB.TokenMintPublicKey = solana.MustPublicKeyFromBase58(v.TokenB.TokenMint)

				if version == "v2" {
					tokenATokenAccount, _, _ := solana.FindAssociatedTokenAddress(v.SwapPublicKey, v.TokenA.TokenMintPublicKey)

					v.TokenA.SwapTokenAccount = tokenATokenAccount.String()

					tokenBTokenAccount, _, _ := solana.FindAssociatedTokenAddress(v.SwapPublicKey, v.TokenB.TokenMintPublicKey)

					v.TokenB.SwapTokenAccount = tokenBTokenAccount.String()

					//  如果是v2 ，要严格确保顺序的正确性，否则数据将会出错
					// if strings.Compare(v.TokenA.TokenMint, v.TokenB.TokenMint) > 0 {
					//	temp := v.TokenB
					//	v.TokenB = v.TokenA
					//	v.TokenA = temp
					// }
				}

				v.TokenA.SwapTokenPublicKey = solana.MustPublicKeyFromBase58(v.TokenA.SwapTokenAccount)
				v.TokenB.SwapTokenPublicKey = solana.MustPublicKeyFromBase58(v.TokenB.SwapTokenAccount)

				if v.TokenA.RefundAddress != "" {
					v.TokenA.RefundAddressPublicKey = solana.MustPublicKeyFromBase58(v.TokenA.RefundAddress)
					v.TokenB.RefundAddressPublicKey = solana.MustPublicKeyFromBase58(v.TokenB.RefundAddress)
				}

				swapMap[v.SwapAccount] = v
				tokenMap[v.TokenA.SwapTokenAccount] = &v.TokenA
				tokenMap[v.TokenB.SwapTokenAccount] = &v.TokenB
			}

			// swapConfigMap = swapMap
			for k, v := range swapMap {
				swapConfigMap[k] = v
			}
			// tokenConfigMap = tokenMap
			for k, v := range tokenMap {
				tokenConfigMap[k] = v
			}

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
	err := conf.UnmarshalKey(chainNetRpcKey, &chainNetsConfig)
	if err != nil {
		return errors.Wrap(err)
	}

	if len(chainNetsConfig) == 0 {
		return errors.New("chain net rpc address is not found!")
	}

	for _, v := range chainNetsConfig {
		chainNets = append(chainNets, &domain.ChainNet{
			Client:  NewRPC(v),
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

		time.Sleep(time.Second * 5)
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
	for k, v := range chainNets {
		height, err := GetBlockHeightForClient(v.Client)
		if err != nil {
			chainNets[k].Client = NewRPC(chainNetsConfig[k])
		}
		slot, err := GetBlockSlotForClient(v.Client)
		if err != nil {
			chainNets[k].Client = NewRPC(chainNetsConfig[k])
		} else {
			v.Slot = slot
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
				if !isInit {
					panic(err)
				} else {
					continue
				}
			}
			var tokenA token.Account
			err = bin.NewBinDecoder(resp.Value.Data.GetBinary()).Decode(&tokenA)
			if err != nil {
				panic(err)
			}
			v.Balance = parse.PrecisionConversion(decimal.NewFromInt(int64(tokenA.Amount)), int(v.Decimal))

			if v.RefundAddress != "" {
				refundResp, err := GetRpcClient().GetTokenAccountBalance(context.Background(), v.RefundAddressPublicKey, rpc.CommitmentConfirmed)
				if err != nil {
					panic(err)
				}
				refundAmount, err := decimal.NewFromString(refundResp.Value.Amount)
				v.RefundBalance = parse.PrecisionConversion(refundAmount, int(v.Decimal))
			} else {
				v.RefundBalance = decimal.Decimal{}
			}
		}

		for _, v := range SwapConfigList() {
			v.TokenA.Balance = GetTokenForTokenAccount(v.TokenA.SwapTokenAccount).Balance
			v.TokenB.Balance = GetTokenForTokenAccount(v.TokenB.SwapTokenAccount).Balance
			if v.Version != "v2" {
				continue
			}
			resp, err := GetRpcClient().GetAccountInfo(
				context.TODO(),
				v.SwapPublicKey,
			)
			if err != nil {
				if !isInit {
					panic(err)
				} else {
					continue
				}
			}

			// get info
			var swapV2 SwapAccountV2
			err = bin.NewBinDecoder(resp.Value.Data.GetBinary()[8:]).Decode(&swapV2)
			if err != nil {
				panic(err)
			}
			rewarders := swapV2.RewarderInfos.list()
			// fmt.Println(rewarders)
			rewarderUsd := decimal.Decimal{}

			getTokenInfo := func(key solana.PublicKey) *domain.TokenInfo {
				for _, r := range TokenList {
					if r.Address.Equals(key) {
						return r
					}
				}
				return nil
			}

			for _, f := range rewarders {
				if f.Mint.IsZero() {
					continue
				}
				// v.Mint
				tokenInfo := getTokenInfo(f.Mint)
				if tokenInfo == nil {
					continue
				}
				emissionsPerSecond := parse.PrecisionConversion(f.EmissionsPerSecond.Val(), int(tokenInfo.Decimal))
				// 同步 token price
				tokenPrice, err := crema.GetPriceForSymbol(tokenInfo.Symbol)
				if err != nil {
					continue
				}

				rewarderUsd = rewarderUsd.Add(emissionsPerSecond.Mul(tokenPrice))
			}

			v.RewarderUsd = rewarderUsd
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
	return rpcClient.GetBlockHeight(context.TODO(), rpc.CommitmentFinalized)
}

func GetBlockSlotForClient(rpcClient *rpc.Client) (uint64, error) {
	return rpcClient.GetSlot(context.TODO(), rpc.CommitmentFinalized)
}

func GetRpcClient() *rpc.Client {
	return chainNet.Client
}

func GetRpcSlot() uint64 {
	return chainNet.Slot
}

func GetWsClient() *ws.Client {
	if wsClient == nil {
		return newWSConnect()
	}
	return wsClient
}

func GetAcEventParser() *event.EventParser {
	return &activityEventParser
}

func abs(n int64) int64 {
	y := n >> 63
	return (n ^ y) - y
}

// GetTokenForTokenAccount 根据token account获取token配置
func GetTokenForTokenAccount(account string) *domain.Token {
	return tokenConfigMap[account]
}

// GetTokenShowDecimalForTokenAccount 根据token account获取token show decimal
func GetTokenShowDecimalForTokenAccount(account string) uint8 {
	d := uint8(4)
	t, ok := tokenConfigMap[account]
	if !ok {
		return d
	}
	if t.ShowDecimal == 0 {
		return d
	}

	return t.ShowDecimal
}

func NewHTTPTransport(
	timeout time.Duration,
	maxIdleConnsPerHost int,
	keepAlive time.Duration,
) *http.Transport {
	return &http.Transport{
		IdleConnTimeout:     timeout,
		MaxIdleConnsPerHost: maxIdleConnsPerHost,
		Proxy:               http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: keepAlive,
		}).Dial,
	}
}

// NewHTTP returns a new Client from the provided config.
func NewHTTP(
	timeout time.Duration,
	maxIdleConnsPerHost int,
	keepAlive time.Duration,
) *http.Client {
	tr := NewHTTPTransport(
		timeout,
		maxIdleConnsPerHost,
		keepAlive,
	)

	return &http.Client{
		Timeout:   timeout,
		Transport: tr,
	}
}

// NewRPC creates a new Solana JSON RPC client.
func NewRPC(rpcEndpoint string) *rpc.Client {
	var (
		defaultMaxIdleConnsPerHost = 10
		defaultTimeout             = 500 * time.Second
		defaultKeepAlive           = 180 * time.Second
	)
	opts := &jsonrpc.RPCClientOpts{
		HTTPClient: NewHTTP(
			defaultTimeout,
			defaultMaxIdleConnsPerHost,
			defaultKeepAlive,
		),
	}
	rpcClient := jsonrpc.NewClientWithOpts(rpcEndpoint, opts)
	return rpc.NewWithCustomRPCClient(rpcClient)
}
