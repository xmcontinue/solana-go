package process

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"sync"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	ag_solanago "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/go-redis/redis/v8"

	"git.cplus.link/crema/backend/chain/sol"
	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/domain"
)

func SyncGalleryJob() error {
	logger.Info("sync gallery job ......")

	outs, err := sol.PubMetadataAccount(collectionMint)
	if err != nil {
		return errors.Wrap(err)
	}
	err = parser(outs)
	if err != nil {
		return errors.Wrap(err)
	}
	logger.Info("sync gallery job finished......")
	return nil
}

func parser(outs []*rpc.KeyedAccount) error {
	ctx := context.Background()
	sortGalleryName := make([]*redis.Z, 0, len(outs))
	fullGallery := make(map[string]interface{}, len(outs))
	galleryAttributes := make(map[string][]interface{})
	limitChan := make(chan struct{}, 5)

	wg := &sync.WaitGroup{}
	wg.Add(len(outs))

	for i := range outs {
		limitChan <- struct{}{}
		go makeGalleryValue(wg, outs[i], limitChan, &sortGalleryName, galleryAttributes, fullGallery)
	}

	wg.Wait()

	pipe := redisClient.TxPipeline()

	err := pushGalleryAttributesByPipe(ctx, &pipe, &galleryAttributes)
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
		logger.Error("push into redis zset error", logger.Errorv(err))
		return errors.Wrap(err)
	}

	return nil
}

func makeGalleryValue(wg *sync.WaitGroup, out *rpc.KeyedAccount, limitChan chan struct{}, sortGalleryName *[]*redis.Z, galleryAttributes map[string][]interface{}, fullGallery map[string]interface{}) error {
	defer func() {
		<-limitChan
		wg.Done()
	}()

	metadata, _ := sol.DecodeMetadata(out.Account.Data.GetBinary())

	repos, err := getUriJson(metadata.Data.Uri)
	if err != nil {
		logger.Error("get metadata uri err", logger.Errorv(err))
		return errors.Wrap(err)
	}

	metadataJson := &sol.MetadataJSON{}
	_ = json.Unmarshal(repos, metadataJson)
	score, err := strconv.ParseFloat(strings.Split(metadataJson.Name, "#")[1], 10)
	if err != nil {
		logger.Error("parse nft name error", logger.String("nft name", metadataJson.Name), logger.Errorv(err))
		return errors.Wrap(err)
	}

	owner, err := sol.GetOwnerByMintAccount(metadata.Mint)
	if err != nil {
		logger.Error("get owner by mint account error", logger.Errorv(err))
		return errors.Wrap(err)
	}

	redisKeyMintAccount := metadata.Mint.String()

	gallery := &sol.Gallery{
		Metadata:     metadata,
		MetadataJSON: metadataJson,
		Mint:         metadata.Mint.String(),
		Name:         metadataJson.Name,
		Owner:        owner,
	}

	lock.Lock()
	defer lock.Unlock()

	fullGallery[redisKeyMintAccount] = gallery

	*sortGalleryName = append(*sortGalleryName, &redis.Z{
		Score:  score,
		Member: redisKeyMintAccount,
	})

	if len(*metadataJson.Attributes) != 0 {
		for _, v := range *metadataJson.Attributes {
			if v.Value == "None" {
				continue
			}

			key := strings.Replace(v.TraitType, " ", "", -1) + ":" + strings.Replace(v.Value, " ", "", -1)
			_, ok := galleryAttributes[key]
			if ok {
				galleryAttributes[key] = append(galleryAttributes[key], redisKeyMintAccount)
			} else {
				galleryAttributes[key] = []interface{}{redisKeyMintAccount}
			}
		}
	}

	return nil
}

func getUriJson(uri string) ([]byte, error) {
	if len(uri) == 0 {
		return nil, nil
	}

	metadataJson, err := model.QueryMetadataJson(context.TODO(), model.NewFilter("uri = ?", uri))
	if err == nil {
		return []byte(metadataJson.Data), nil
	}

	repos, err := httpClient.R().Get(uri)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	if repos.StatusCode() != 200 {
		return nil, errors.Wrap(errors.RecordNotFound)
	}

	_ = model.CreateMetadataJson(context.TODO(), &domain.MetadataJsonDate{
		URI:  uri,
		Data: string(repos.Body()),
	})

	return repos.Body(), nil
}

type SubMetadata struct {
	account        string
	collectionMint string
}

func (s *SubMetadata) Sub() error {
	collectionBase58 := ag_solanago.MustPublicKeyFromBase58(s.collectionMint).Bytes()
	filters := []rpc.RPCFilter{
		{
			DataSize: sol.CremaMetadataDataSize,
		},
		{
			Memcmp: &rpc.RPCFilterMemcmp{
				Offset: sol.CremaMetadataCollectionIndex,
				Bytes:  collectionBase58[:],
			},
		},
	}

	sub, err := sol.GetWsClient().ProgramSubscribeWithOpts(ag_solanago.MustPublicKeyFromBase58(s.account), rpc.CommitmentConfirmed, ag_solanago.EncodingBase64, filters)
	if err != nil {
		logger.Error("sub metadata err:", logger.Errorv(err))
		panic(err)
	}

	defer sub.Unsubscribe()

	for {
		recV, err := sub.Recv()
		if err != nil {
			logger.Error("sub metadata err:", logger.Errorv(err))
			panic(err)
		}
		logger.Info("sub metadata", logger.String("public key:", recV.Value.Pubkey.String()))
		go parser([]*rpc.KeyedAccount{&recV.Value})

	}
}
