package handler

import (
	"context"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/transport/rpcx"

	"git.cplus.link/crema/backend/chain/sol"
	"git.cplus.link/crema/backend/internal/worker/market"
	"git.cplus.link/crema/backend/pkg/domain"
	"git.cplus.link/crema/backend/pkg/iface"
)

func (t *MarketService) GetGalleryV2(ctx context.Context, args *iface.GetGalleryReq, reply *iface.GetGalleryResp) error {
	defer rpcx.Recover(ctx)
	if err := validate(args); err != nil {
		return errors.Wrapf(errors.ParameterError, "validate:%v", err)
	}
	return nil
}

func (t *MarketService) getGalleryV2(ctx context.Context, args *iface.GetGalleryReq) ([]*sol.Gallery, error) {
	fil := make([]string, 0, 2)
	innerTypeSet := map[string]market.GallerySet{}
	valueOf := reflect.ValueOf(args.GalleryType)
	typeOf := reflect.TypeOf(&args.GalleryType)

	mintAccountType, mintAccountData := market.GetGalleryCache()
	for i := 0; i < valueOf.NumField(); i++ {
		fieldV := valueOf.Field(i)
		if fieldV.String() == "<[]string Value>" {
			galleryValues := make([]string, 0)
			for j := 0; j < fieldV.Len(); j++ {
				galleryValues = append(galleryValues, strings.Replace(fieldV.Index(j).String(), " ", "", -1))
			}

			if len(galleryValues) == 0 {
				continue
			}

			tt := typeOf.Elem().Field(i)
			attributeType := tt.Tag.Get("yaml")

			galleryAttributeKeys := getGalleryAttributeKey(attributeType, galleryValues)
			galleryMintAccounts := make(market.GallerySet)
			for _, v := range galleryAttributeKeys {
				galleryMintAccounts = galleryMintAccounts.Union(mintAccountType[v]) // 同一个类型执行交集
			}

			galleryAttributeKey := domain.GetGalleryPrefix() + ":temp:" + attributeType
			fil = append(fil, galleryAttributeKey)
			innerTypeSet[galleryAttributeKey] = galleryMintAccounts
		}

	}

	out := make(market.GallerySet)
	if len(fil) != 0 {

		for _, v := range innerTypeSet {
			if len(out) == 0 {
				out = deepCopy(v)
			}
			out = out.Intersect(v)
		}
	} else {
		for _, v := range mintAccountType {
			out = out.Union(v)
		}
	}

	gallerys := make([]*sol.Gallery, 0, len(out))
	for v := range out {
		gallerys = append(gallerys, mintAccountData[v])
	}

	if args.ISPositive {
		sort.Slice(gallerys, func(i, j int) bool {
			scorej, _ := strconv.ParseFloat(strings.Split(gallerys[j].Name, "#")[1], 10)
			scorei, _ := strconv.ParseFloat(strings.Split(gallerys[i].Name, "#")[1], 10)

			if scorej > scorei {
				return true
			}
			return false
		})
	} else {
		sort.Slice(gallerys, func(i, j int) bool {
			scorej, _ := strconv.ParseFloat(strings.Split(gallerys[j].Name, "#")[1], 10)
			scorei, _ := strconv.ParseFloat(strings.Split(gallerys[i].Name, "#")[1], 10)

			if scorej < scorei {
				return true
			}
			return false
		})
	}
	return gallerys, nil
}

func deepCopy(src market.GallerySet) market.GallerySet {
	dest := market.GallerySet{}
	for k, v := range src {
		dest[k] = v
	}
	return dest
}
