package process

import (
	"context"

	"git.cplus.link/go/akit/errors"
	"github.com/go-redis/redis/v8"

	"git.cplus.link/crema/backend/pkg/domain"
)

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
