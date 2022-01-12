package auth

import (
	"context"
	"time"

	"git.cplus.link/cassava/common/pkg/jwtx"
	redisV8 "git.cplus.link/go/akit/client/redis/v8"
	"git.cplus.link/go/akit/config"
	"git.cplus.link/go/akit/errors"
	net "git.cplus.link/go/akit/transport"
	"git.cplus.link/go/akit/transport/http"
	"github.com/gin-gonic/gin"
)

// JWTAuth 客户端(web/iOS/android)JWT认证
type JWTAuth struct {
	SigningKey  map[string]*jwtx.SigningKeyConf
	RedisClient *redisV8.Client
}

// Name 命名
func (*JWTAuth) Name() string {
	return "JWTAuth"
}

// NewJWTAuth 创建一个用于
func NewJWTAuth(ctx context.Context, conf *config.Config) (*JWTAuth, error) {
	var (
		signingKeyConf = map[string]*jwtx.SigningKeyConf{}
		redisConf      = redisV8.DefaultRedisConfig()
	)
	if err := conf.UnmarshalKey("SigningKeyConf", &signingKeyConf); err != nil {
		return nil, errors.Wrap(err)
	}
	if err := conf.UnmarshalKey("redis", redisConf); err != nil {
		return nil, errors.Wrap(err)
	}

	client, err := redisV8.NewClient(redisConf)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return &JWTAuth{
		SigningKey:  signingKeyConf,
		RedisClient: client,
	}, nil
}

// NoAuth 登录认证
func (j *JWTAuth) NoAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		var info struct {
			Version  string `header:"ex-ver"       binding:"required"`
			Dev      string `header:"ex-dev"       binding:"required"`
			Ts       int64  `header:"ex-ts"        binding:"required"`
			Language string `header:"ex-language"  binding:"required"`
		}

		if err := c.BindHeader(&info); err != nil {
			http.ResponseError(c, errors.Wrapf(errors.ParameterError, "error:%v", err))
			return
		}

		if err := checkTs(info.Ts); err != nil {
			http.ResponseError(c, errors.Wrapf(errors.Unauthorized, "error:%v", err))
			return
		}
		// 向rpc服务端传递的数据
		traces := []*net.Trace{
			net.NewTrace("remote_addr", c.Request.RemoteAddr),
			net.NewTrace("ver", info.Version),
			net.NewTrace("dev", info.Dev),
			net.NewTrace("ts", info.Ts),
			net.NewTrace("lang", info.Language),
			net.NewTrace("ipaddr", c.ClientIP()),
		}
		http.WithTraces(c, traces...)
	}
}

// checkTs 检查时间戳
func checkTs(ts int64) error {
	curTs := time.Now().UTC().Unix()
	if (curTs-ts/1000000) > 1800 || (curTs-ts/1000000) < -1800 {
		return errors.Errorf("invalid timestamp request ts:%d system ts:%d", ts/1000000, curTs)
	}
	return nil
}

// getSigningKey 获取SigningKey 配置
func (j *JWTAuth) getSigningKey(version string) (*jwtx.SigningKeyConf, error) {
	sconf, ok := j.SigningKey[version]
	if ok {
		return sconf, nil
	}
	sconf, ok = j.SigningKey["default"]
	if ok {
		return sconf, nil
	}
	return nil, errors.Errorf("version:%s session config not found", version)
}
