package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
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

	gallery, err := t.getGallery(ctx, args)
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
		for i, v := range gallery {
			if strings.EqualFold(v.Mint, args.Query) || strings.EqualFold(v.Owner, args.Query) {
				returnGalleries = append(returnGalleries, gallery[i])
			}
		}
	}

	return errors.Wrap(returnFunc(returnGalleries, args, reply))
}

func (t *MarketService) filterByQuery(ctx context.Context, args *iface.GetGalleryReq, reply *iface.GetGalleryResp, filterKey string) error {
	names, cla, err := t.redisClient.ZScan(ctx, domain.GetSortedGalleryKey(), 0, filterKey, -1).Result()
	if cla == 0 {
		return nil
	}

	galleries, err := t.redisClient.MGet(ctx, names...).Result()
	if err != nil {
		return errors.Wrap(err)
	}

	// 不是public account,是名字
	newGalleries := make([]*sol.Gallery, 0, len(galleries))
	for _, v := range galleries {

		vStr, _ := v.(string)

		gallery := &sol.Gallery{}
		_ = json.Unmarshal([]byte(vStr), gallery)

		newGalleries = append(newGalleries, gallery)
	}

	return errors.Wrap(returnFunc(newGalleries, args, reply))
}

func returnFunc(gallery []*sol.Gallery, args *iface.GetGalleryReq, reply *iface.GetGalleryResp) error {
	if args.ISPositive {
		sort.Slice(gallery, func(i, j int) bool {
			if strings.Compare(gallery[i].Name, gallery[j].Name) < 0 {
				return false
			}
			return true
		})
	}

	if int64(len(gallery)) < args.Offset {
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
		key[i] = fmt.Sprintf(domain.GalleryPrefix + ":name:" + key[i])
	}
	return key
}

func (t *MarketService) getGallery(ctx context.Context, args *iface.GetGalleryReq) ([]*sol.Gallery, error) {
	var err error
	pipe := t.redisClient.TxPipeline()
	fil := make([]string, 0, 2)

	valueOf := reflect.ValueOf(*args)
	typeOf := reflect.TypeOf(args)
	for i := 0; i < valueOf.NumField(); i++ {

		fieldV := valueOf.Field(i)

		if fieldV.String() == "<[]string Value>" {
			sliceValue := make([]string, 0)
			for j := 0; j < fieldV.Len(); j++ {
				sliceValue = append(sliceValue, fieldV.Index(j).String())
			}

			if len(sliceValue) == 0 {
				continue
			}

			tt := typeOf.Elem().Field(i)
			finalName := tt.Tag.Get("json")
			fil = append(fil, domain.GalleryPrefix+":temp:"+finalName)
			_, err = pipe.SUnionStore(ctx, domain.GalleryPrefix+":temp:"+finalName, getGalleryAttributeKey(tt.Tag.Get("json"), sliceValue)...).Result()
			if err != nil {
				return nil, errors.Wrap(err)
			}
		}

	}

	//if len(args.CoffeeMembership) != 0 {
	//	fil = append(fil, domain.GalleryPrefix+":temp:Coffee Membership")
	//	_, err = pipe.SUnionStore(ctx, domain.GalleryPrefix+":temp:Coffee Membership", getGalleryAttributeKey("Coffee Membership", args.CoffeeMembership)...).Result()
	//	if err != nil {
	//		return nil, errors.Wrap(err)
	//	}
	//} else if len(args.Body) != 0 {
	//	fil = append(fil, domain.GalleryPrefix+":temp:body")
	//	_, err = pipe.SUnionStore(ctx, domain.GalleryPrefix+":temp:body", getGalleryAttributeKey("Body", args.Body)...).Result()
	//	if err != nil {
	//		return nil, errors.Wrap(err)
	//	}
	//}

	if len(fil) != 0 {
		_, err = pipe.SInterStore(ctx, domain.GalleryPrefix+":temp:settle", fil...).Result()
		if err != nil {
			return nil, errors.Wrap(err)
		}

		store := &redis.ZStore{
			Keys:    []string{domain.GalleryPrefix + ":temp:settle", domain.GetSortedGalleryKey()},
			Weights: []float64{0, 1},
		}
		_, err = pipe.ZInterStore(ctx, domain.GalleryPrefix+":temp:final", store).Result()
		if err != nil {
			return nil, errors.Wrap(err)
		}
	} else {
		store := &redis.ZStore{
			Keys:    []string{domain.GetSortedGalleryKey()},
			Weights: []float64{1},
		}
		_, err = pipe.ZInterStore(ctx, domain.GalleryPrefix+":temp:final", store).Result()
		if err != nil {
			return nil, errors.Wrap(err)
		}
	}

	result := pipe.ZRange(ctx, domain.GalleryPrefix+":temp:final", 0, -1)

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