package handler

import (
	"context"
	netHttp "net/http"
	"reflect"
	"sync"

	"git.cplus.link/go/akit/config"
	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/transport/http"
	"git.cplus.link/go/akit/transport/rpcx"
	"github.com/gin-gonic/gin"

	"git.cplus.link/crema/backend/pkg/iface"
	"git.cplus.link/crema/backend/pkg/iface/client"
)

var (
	once           sync.Once
	ExchangeClient *client.ExchangeClient
)

// Init handler初始化
func Init(c *config.Config) error {
	var rErr error
	once.Do(func() {
		var err error
		ExchangeClient, err = newCremaExchangeClient(c)
		if err != nil {
			rErr = errors.Wrap(err)
		}
	})

	return errors.Wrap(rErr)
}

func newCremaExchangeClient(c *config.Config) (*client.ExchangeClient, error) {
	conf, err := c.Service(iface.ExchangeServiceName)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	cli, err := client.NewCremaExchangeClient(context.Background(), conf)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return cli.(*client.ExchangeClient), nil
}

func exchangeClient() *rpcx.Client {
	return ExchangeClient.Client
}

type rpcxClient func() *rpcx.Client
type preCallHandle = func(*gin.Context, string, interface{}) error

func handleFunc(cli rpcxClient, name string, reqArg, respArg interface{}, preCalls ...preCallHandle) func(*gin.Context) {
	//
	reqTyp := reflect.TypeOf(reqArg)
	if reqTyp.Kind() == reflect.Ptr {
		reqTyp = reqTyp.Elem()
	}
	//
	respTyp := reflect.TypeOf(respArg)
	if respTyp.Kind() == reflect.Ptr {
		respTyp = respTyp.Elem()
	}
	return func(c *gin.Context) {
		//
		req := reflect.New(reqTyp).Interface() // Elem().
		resp := reflect.New(respTyp).Interface()

		if c.Request.Method == netHttp.MethodGet {
			if err := c.BindQuery(req); err != nil {
				http.ResponseError(c, errors.Wrapf(errors.ParameterError, "bind query"))
				return
			}
		} else {
			if err := c.BindJSON(req); err != nil {
				http.ResponseError(c, errors.Wrapf(errors.ParameterError, "bind json"))
				return
			}
		}
		for _, f := range preCalls {
			if err := f(c, name, req); err != nil {
				http.ResponseError(c, errors.Wrapf(err, "pre call"))
				return
			}
		}
		if err := cli().Call(c, name, req, resp); err != nil {
			http.ResponseError(c, errors.Wrapf(err, "call"))
			return
		}
		http.ResponseOK(c, resp)
	}
}
