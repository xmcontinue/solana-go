package process

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/go-redis/redis/v8"

	"git.cplus.link/crema/backend/chain/sol"
	"git.cplus.link/crema/backend/pkg/domain"
)

// 1,同步所有数据到本地
// 2,使用redis string 结构存储  「score，nft name」 -> 完整的nft 对象数据
// 3,按照level分成不同的集合，key 为level，值为nft name
// 4，按照属性分为不同的集合，key 为属性，值为nft name
// 5,每个属性有很多值，每个值对应很多nft,同一个属性里面是并集，不同的属性做交集

// 设计，一个有序集合，每个属性对应一个集合

func SyncGalleryJob() error {
	logger.Info("sync gallery job ......")
	ctx := context.Background()
	outs, err := sol.PubMetadataAccount(collectionMint)
	if err != nil {
		return errors.Wrap(err)
	}

	sortGalleryName := make([]*redis.Z, 0, len(outs))
	fullGallery := make(map[string]interface{}, len(outs))
	galleryAttributes := make(map[string][]interface{})
	limitChan := make(chan struct{}, 10)

	for i := range outs {
		limitChan <- struct{}{}
		_ = makeGalleryValue(outs[i], limitChan, &sortGalleryName, galleryAttributes, fullGallery)
	}

	pipe := redisClient.TxPipeline()

	err = pushGalleryAttributesByPipe(ctx, &pipe, &galleryAttributes)
	if err != nil {
		return errors.Wrap(err)
	}

	err = pushSortedGallery(ctx, &pipe, sortGalleryName)
	if err != nil {
		return errors.Wrap(err)
	}

	err = pushAllGallery(ctx, &pipe, &fullGallery)
	if err != nil {
		return errors.Wrap(err)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}

func makeGalleryValue(out *rpc.KeyedAccount, limitChan chan struct{}, sortGalleryName *[]*redis.Z, galleryAttributes map[string][]interface{}, fullGallery map[string]interface{}) error {
	defer func() {
		<-limitChan
	}()

	metadata, _ := sol.DecodeMetadata(out.Account.Data.GetBinary())

	repos, err := httpClient.R().Get(metadata.Data.Uri)
	if err != nil {
		logger.Error("get metadata uri err", logger.Errorv(err))
		return errors.Wrap(err)
	}

	if repos.StatusCode() != 200 {
		return errors.Wrap(errors.RecordNotFound)
	}

	metadataJson := &sol.MetadataJSON{}
	_ = json.Unmarshal(repos.Body(), metadataJson)
	score, err := strconv.ParseFloat(strings.Split(metadataJson.Name, "#")[1], 10)
	if err != nil {
		return errors.Wrap(err)
	}

	owner, err := sol.GetUserAddressForTokenKey(metadata.Mint)
	if err != nil {
		return errors.Wrap(err)
	}

	gallery := &sol.Gallery{
		Metadata:     metadata,
		MetadataJSON: metadataJson,
		Owner:        owner,
		Mint:         metadata.Mint.String(),
		Name:         metadataJson.Name,
	}

	*sortGalleryName = append(*sortGalleryName, &redis.Z{
		Score:  score,
		Member: metadataJson.Name,
	})

	if len(*metadataJson.Attributes) != 0 {
		for _, v := range *metadataJson.Attributes {
			if v.Value == "None" {
				continue
			}

			key := v.TraitType + ":" + v.Value
			_, ok := galleryAttributes[key]
			if ok {
				galleryAttributes[key] = append(galleryAttributes[key], metadataJson.Name)
			} else {
				galleryAttributes[key] = []interface{}{metadataJson.Name}
			}
		}
	}

	fullGallery[metadataJson.Name] = gallery

	return nil
}

func pushSortedGallery(ctx context.Context, pipe *redis.Pipeliner, sortedGallery []*redis.Z) error {
	_, err := (*pipe).ZAdd(ctx, domain.GetSortedGalleryKey(), sortedGallery...).Result()
	if err != nil {
		return errors.Wrap(err)
	}
	return nil

}

func pushAllGallery(ctx context.Context, pipe *redis.Pipeliner, allGallery *map[string]interface{}) error {
	for k, v := range *allGallery {
		_, err := (*pipe).Set(ctx, domain.GetAllGalleryKey(k), v, 0).Result()
		if err != nil {
			return errors.Wrap(err)
		}
	}

	return nil
}

func pushGalleryAttributesByPipe(ctx context.Context, pipe *redis.Pipeliner, attributeMap *map[string][]interface{}) error {
	for attributeValue, attributes := range *attributeMap {
		_, err := pushAttributes(ctx, domain.GetGalleryAttributeKey(attributeValue), pipe, attributes)
		if err != nil {
			return errors.Wrap(err)
		}
	}
	return nil
}

func pushAttributes(ctx context.Context, key string, pipe *redis.Pipeliner, attributes []interface{}) (int64, error) {
	i, err := (*pipe).SAdd(ctx, key, attributes...).Result()
	if err != nil {
		return 0, errors.Wrap(err)
	}
	return i, nil
}
