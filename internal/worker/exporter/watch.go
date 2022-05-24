package exporter

import (
	"encoding/json"
	"time"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	"github.com/go-resty/resty/v2"

	"git.cplus.link/crema/backend/pkg/iface"
	"git.cplus.link/crema/backend/pkg/prometheus"
)

type Api struct {
	Path   string
	Method string
}

var (
	client   = resty.New()
	retryNum = 3
)

var cremaApis = map[string]Api{
	"swap_count": {
		"/v1/swap/count",
		"GET",
	},
	"tvl_vol": {
		"/v1/histogram?date_type=1min&typ=tvl&limit=100",
		"GET",
	},
	"tvl": {
		"/v1/token/tvl?symbol=USDC",
		"GET",
	},
	"token_config": {
		"/token/config",
		"GET",
	},
	"exchange": {
		"/price",
		"GET",
	},
}

// WatchCremaHttpStatus 监听crema http接口状态
func WatchCremaHttpStatus() error {

	for i := 0; i < retryNum; i++ {
		if getStatus() {
			// 正常，通知push gateway
			sendHttpStatusMsgToPushGateway(1)
			return nil
		}
	}
	// 异常，通知gateway
	sendHttpStatusMsgToPushGateway(0)
	return nil
}

// WatchExchangeSyncTime 监听crema 汇率服务价格更新时间
func WatchExchangeSyncTime() error {
	resp, err := client.R().
		SetQueryParams(map[string]string{}).
		Get(job.CremaHost + cremaApis["exchange"].Path)
	if err != nil {
		return errors.Wrap(err)
	}

	var raw struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Time string `json:"time"`
		} `json:"data"`
	}
	err = json.Unmarshal(resp.Body(), &raw)
	if err != nil {
		return errors.Wrap(err)
	}

	if raw.Code != "OK" || raw.Msg != "Success" {
		return errors.Wrap(err)
	}

	syncTime, err := time.Parse("2006-01-02 15:04:05", raw.Data.Time)
	now, err := time.Parse("2006-01-02 15:04:05", time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		return errors.Wrap(err)
	}
	t := now.Unix() - syncTime.Unix()
	if t < 0 {
		return errors.New("time unusual")
	}
	sendExchangeSyncTimeMsgToPushGateway(float64(t))

	return nil
}

func sendHttpStatusMsgToPushGateway(status float64) {
	log := &iface.LogReq{
		LogName:  "api_status",
		LogValue: status,
		LogHelp:  "all http api status",
		JobName:  "api_status",
		Tags: map[string]string{
			"project": "crema",
		},
	}
	err := prometheus.ExamplePusherPush(log)
	if err != nil {
		logger.Error("send msg to push_gateway failed!", logger.Errorv(err))
	}
}

func sendExchangeSyncTimeMsgToPushGateway(time float64) {
	log := &iface.LogReq{
		LogName:  "exchange_sync_time",
		LogValue: time,
		LogHelp:  "crema exchange sync time lag difference (s)",
		JobName:  "exchange_sync_time",
		Tags: map[string]string{
			"project": "crema",
		},
	}
	err := prometheus.ExamplePusherPush(log)
	if err != nil {
		logger.Error("send msg to push_gateway failed!", logger.Errorv(err))
	}
}

func getStatus() bool {
	for _, v := range cremaApis {
		resp, err := client.R().
			SetQueryParams(map[string]string{}).
			Get(job.CremaHost + v.Path)
		if err != nil {
			return false
		}

		var raw struct {
			Code string `json:"code"`
			Msg  string `json:"msg"`
		}
		err = json.Unmarshal(resp.Body(), &raw)
		if err != nil {
			return false
		}

		if raw.Code != "OK" || raw.Msg != "Success" {
			return false
		}
	}
	return true
}
