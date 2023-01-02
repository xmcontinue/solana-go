package market

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"git.cplus.link/go/akit/errors"

	"git.cplus.link/crema/backend/chain/sol"
	"git.cplus.link/crema/backend/pkg/domain"
)

var (
	mintAccountType map[string]GallerySet   // gallery类型到mint地址的映射关系
	mintAccountData map[string]*sol.Gallery // mint 地址 到具体数据的映射
	mu              sync.RWMutex
)

var (
	galleryNameKeys   []string
	galleryTypeKeys   []string
	gallerySortedKeys []string
)

func init() {
	galleryNameKeys = make([]string, 0, 11000)
	galleryTypeKeys = make([]string, 0, 256)
	gallerySortedKeys = make([]string, 0, 10)
}

func syncGalleryCache() error {
	keys, err := redisClient.Keys(context.Background(), domain.GetGalleryPrefix()+"*").Result()
	if err != nil {
		return errors.Wrap(err)
	}

	galleryNameKeys = galleryNameKeys[:0]
	galleryTypeKeys = galleryTypeKeys[:0]
	gallerySortedKeys = gallerySortedKeys[:0]
	for _, v := range keys {
		if strings.HasPrefix(v, domain.GetGalleryPrefix()+":name:") {
			galleryNameKeys = append(galleryNameKeys, v)
		} else if strings.Contains(v, ":temp:") {
			continue
		} else if strings.Contains(v, ":sorted") {
			gallerySortedKeys = append(gallerySortedKeys, v)
		} else {
			galleryTypeKeys = append(galleryTypeKeys, v)
		}
	}

	galleryNames, err := redisClient.MGet(context.Background(), galleryNameKeys...).Result()
	if err != nil {
		return errors.Wrap(err)
	}

	mintAccountDataTemp := make(map[string]*sol.Gallery, len(galleryNames))
	for _, v := range galleryNames {
		gallery := &sol.Gallery{}
		err = json.Unmarshal([]byte(v.(string)), gallery)
		if err != nil {
			return errors.Wrap(err)
		}
		mintAccountDataTemp[gallery.Mint] = gallery
	}

	mintAccountTypeTemp := make(map[string]GallerySet)
	for _, v := range galleryTypeKeys {
		addrs, err := redisClient.SMembersMap(context.Background(), v).Result()
		if err != nil {
			return errors.Wrap(err)
		}

		addr := GallerySet{}
		for k, v := range addrs {
			addr[k] = v
		}

		mintAccountTypeTemp[v] = addr
	}

	mu.Lock()
	defer mu.Unlock()
	mintAccountType = mintAccountTypeTemp
	mintAccountData = mintAccountDataTemp

	return nil
}

func GetGalleryCache() (map[string]GallerySet, map[string]*sol.Gallery) {
	mu.RLock()
	defer mu.RUnlock()

	return mintAccountType, mintAccountData
}

type GallerySet map[string]struct{}

func (g *GallerySet) Union(g1 GallerySet) GallerySet {
	for k, v := range g1 {
		(*g)[k] = v
	}

	return *g
}

// Intersect 交集
func (g *GallerySet) Intersect(g1 GallerySet) GallerySet {

	var min int
	if len(*g) > len(g1) {
		min = len(g1)
	} else {
		min = len(*g)
	}

	rg := make(GallerySet, min)

	for k, v := range *g {
		_, ok := g1[k]
		if ok {
			rg[k] = v
		}
	}

	return rg
}
