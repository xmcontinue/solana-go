package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/transport/rpcx"
	ag_solanago "github.com/gagliardetto/solana-go"
	"github.com/go-redis/redis/v8"

	"git.cplus.link/crema/backend/chain/sol"
	"git.cplus.link/crema/backend/pkg/domain"
	"git.cplus.link/crema/backend/pkg/iface"
)

// GetGallery ...
// 先走属性筛选一次，然后内存过滤query 字段
func (t *MarketService) GetGallery(ctx context.Context, args *iface.GetGalleryReq, reply *iface.GetGalleryResp) error {
	defer rpcx.Recover(ctx)
	if err := validate(args); err != nil {
		return errors.Wrapf(errors.ParameterError, "validate:%v", err)
	}
	gallery, err := t.getGalleryV2(ctx, args)
	//gallery, err := t.getGallery(ctx, args)
	if err != nil {
		return errors.Wrap(err)
	}

	if gallery == nil {
		return nil
	}

	if args.Query == "" {
		return errors.Wrap(returnFunc(gallery, args, reply))
	}

	returnGalleries := make([]*sol.Gallery, 0, len(gallery))
	_, err = ag_solanago.PublicKeyFromBase58(args.Query)
	if err != nil {
		// 不是public account,是名字
		for i, v := range gallery {
			if strings.Contains(v.Name, args.Query) {
				returnGalleries = append(returnGalleries, gallery[i])
			}
		}
	} else {
		// 是public account
		_, err := t.redisClient.Get(ctx, domain.GetAllGalleryKey(args.Query)).Result()
		if err != nil {
			// 是wallet account
			if !t.redisClient.ErrIsNil(err) {
				return errors.Wrap(err)
			}
			mints, err := sol.GetMintsByTokenOwner(ctx, args.Query)
			if err != nil {
				return errors.Wrap(err)
			}

			for _, v := range gallery {
				for _, mint := range mints {
					if v.Mint == mint {
						v.Owner = args.Query // 因为同步时间间隔问题，此处将owner替换为实际钱包地址
						returnGalleries = append(returnGalleries, v)
					}
				}
			}

		} else {
			for _, v := range gallery {
				if v.Mint == args.Query {
					returnGalleries = append(returnGalleries, v)
					break
				}
			}
		}
	}

	return errors.Wrap(returnFunc(returnGalleries, args, reply))
}

func returnFunc(gallery []*sol.Gallery, args *iface.GetGalleryReq, reply *iface.GetGalleryResp) error {
	if len(gallery) == 0 {
		reply.List = nil
	} else if int64(len(gallery)) < args.Offset {
		reply.List = gallery
	} else if int64(len(gallery)) > args.Offset && int64(len(gallery)) < args.Offset+args.Limit {
		reply.List = gallery[args.Offset:]
	} else {
		reply.List = gallery[args.Offset : args.Offset+args.Limit]
	}

	reply.Total = int64(len(gallery))
	reply.Limit = args.Limit
	reply.Offset = args.Offset
	return nil
}

func getGalleryAttributeKey(pre string, key []string) []string {
	for i := range key {
		key[i] = domain.GetGalleryAttributeKey(pre + ":" + key[i])
	}
	return key
}

func getGalleryNameKey(key []string) []string {
	for i := range key {
		key[i] = fmt.Sprintf(domain.GetGalleryPrefix() + ":name:" + key[i])
	}
	return key
}

func (t *MarketService) getGallery(ctx context.Context, args *iface.GetGalleryReq) ([]*sol.Gallery, error) {
	var err error
	pipe := t.redisClient.TxPipeline()
	fil := make([]string, 0, 2)

	valueOf := reflect.ValueOf(args.GalleryType)
	typeOf := reflect.TypeOf(&args.GalleryType)
	for i := 0; i < valueOf.NumField(); i++ {

		fieldV := valueOf.Field(i)

		if fieldV.String() == "<[]string Value>" {
			sliceValue := make([]string, 0)
			for j := 0; j < fieldV.Len(); j++ {
				sliceValue = append(sliceValue, strings.Replace(fieldV.Index(j).String(), " ", "", -1))
			}

			if len(sliceValue) == 0 {
				continue
			}

			tt := typeOf.Elem().Field(i)
			finalName := tt.Tag.Get("yaml")
			fil = append(fil, domain.GetGalleryPrefix()+":temp:"+finalName)
			_, err = pipe.SUnionStore(ctx, domain.GetGalleryPrefix()+":temp:"+finalName, getGalleryAttributeKey(finalName, sliceValue)...).Result()
			if err != nil {
				return nil, errors.Wrap(err)
			}
		}
	}

	if len(fil) != 0 {
		_, err = pipe.SInterStore(ctx, domain.GetGalleryPrefix()+":temp:settle", fil...).Result()
		if err != nil {
			return nil, errors.Wrap(err)
		}

		store := &redis.ZStore{
			Keys:    []string{domain.GetGalleryPrefix() + ":temp:settle", domain.GetSortedGalleryKey()},
			Weights: []float64{0, 1},
		}
		_, err = pipe.ZInterStore(ctx, domain.GetGalleryPrefix()+":temp:final", store).Result()
		if err != nil {
			return nil, errors.Wrap(err)
		}
	} else {
		store := &redis.ZStore{
			Keys:    []string{domain.GetSortedGalleryKey()},
			Weights: []float64{1},
		}
		_, err = pipe.ZInterStore(ctx, domain.GetGalleryPrefix()+":temp:final", store).Result()
		if err != nil {
			return nil, errors.Wrap(err)
		}
	}

	result := &redis.StringSliceCmd{}
	if args.ISPositive {
		result = pipe.ZRange(ctx, domain.GetGalleryPrefix()+":temp:final", 0, -1)
	} else {
		result = pipe.ZRevRange(ctx, domain.GetGalleryPrefix()+":temp:final", 0, -1)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	if result.Err() != nil {
		return nil, errors.Wrap(err)
	}

	if len(result.Val()) == 0 {
		return nil, nil
	}

	galleries, err := t.redisClient.MGet(ctx, getGalleryNameKey(result.Val())...).Result()
	if err != nil {
		return nil, errors.Wrap(err)
	}

	newGalleries := make([]*sol.Gallery, 0, len(galleries))
	for _, v := range galleries {
		vStr, _ := v.(string)

		gallery := &sol.Gallery{}
		_ = json.Unmarshal([]byte(vStr), gallery)

		newGalleries = append(newGalleries, gallery)
	}
	return newGalleries, nil
}

func (t *MarketService) GetGalleryType(ctx context.Context, _ *iface.NilReq, reply *iface.GetGalleryTypeResp) error {
	defer rpcx.Recover(ctx)

	reply.GalleryType = galleryType

	return nil
}
