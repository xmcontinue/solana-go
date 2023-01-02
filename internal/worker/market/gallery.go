package market

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"git.cplus.link/go/akit/errors"
	"github.com/deckarep/golang-set"

	"git.cplus.link/crema/backend/chain/sol"
	"git.cplus.link/crema/backend/pkg/domain"
)

var (
	mintAccountType map[string]mapset.Set   // gallery类型到mint地址的映射关系
	mintAccountData map[string]*sol.Gallery // mint 地址 到具体数据的映射
	mu              sync.RWMutex
)

func syncGalleryCache() error {
	//redisClient.MGet(context.Background(), domain.GetGalleryPrefix())
	keys, err := redisClient.Keys(context.Background(), domain.GetGalleryPrefix()+"*").Result()
	if err != nil {
		return errors.Wrap(err)
	}

	galleryNameKeys := make([]string, 0, len(keys))
	galleryTypeKeys := make([]string, 0, 256)
	gallerySortedKeys := make([]string, 0, 10)
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
		_ = json.Unmarshal([]byte(v.(string)), gallery)
		mintAccountDataTemp[gallery.Mint] = gallery
	}

	mintAccountTypeTemp := make(map[string]mapset.Set)
	for _, v := range galleryTypeKeys {
		addrs, err := redisClient.SMembersMap(context.Background(), v).Result()
		if err != nil {
			return errors.Wrap(err)
		}

		addr := mapset.NewSet()
		for k := range addrs {
			addr.Add(k)
		}

		mintAccountTypeTemp[v] = addr
	}

	mu.Lock()
	defer mu.Unlock()
	mintAccountType = mintAccountTypeTemp
	mintAccountData = mintAccountDataTemp

	return nil
}

func GetGalleryCache() (map[string]mapset.Set, map[string]*sol.Gallery) {
	mu.RLock()
	defer mu.RUnlock()
	return mintAccountType, mintAccountData
}
