package process

import (
	"context"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"

	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/domain"
	"git.cplus.link/crema/backend/pkg/kline"
)

// swap count kline 数据迁移

func migrateSwapCountKline() error {
	beginID := int64(0)

	for {
		filters := []model.Filter{
			model.NewFilter("is_migrate = false"),
			model.NewFilter("id > ?", beginID),
			model.OrderFilter("id asc"),
			model.DateTypeFilter(domain.DateMin),
		}
		// 只在原始表查询数据
		swapCountKLines, err := model.QuerySwapCountKLinesInBaseTable(context.Background(), 100, 0, filters...)
		if err != nil {
			return errors.Wrap(err)
		}

		if len(swapCountKLines) == 0 {
			break
		}

		syncMigrateID := swapCountKLines[len(swapCountKLines)-1].ID

		for _, swapCountKLine := range swapCountKLines {
			trans := func(ctx context.Context) error {
				if swapCountKLine == nil {
					return nil
				}

				if err = model.UpdateSwapCountKLinesInBaseTable(ctx,
					map[string]interface{}{
						"is_migrate": true,
					},
					model.IDFilter(swapCountKLine.ID),
					model.NewFilter("is_migrate = ?", false),
				); err != nil {
					return errors.Wrap(err)
				}

				// 去掉主键，由数据库自动生成
				swapCountKLine.ID = 0

				newKline := kline.NewKline(swapCountKLine.Date)
				for _, t := range newKline.Types {
					swapCountKLine.DateType = t.DateType
					swapCountKLine.Date = t.Date
					// 获取价格
					tokenAPrice, tokenBPrice, err := priceToSwapKLineHandle(ctx, swapCountKLine)
					if err != nil {
						return errors.Wrap(err)
					}
					swapCountKLine.TokenAUSDForContract = tokenAPrice
					swapCountKLine.TokenBUSDForContract = tokenBPrice

					if t.DateType == domain.DateMin {
						swapCountKLine.VolInUsdForContract = swapCountKLine.TokenAVolume.Mul(swapCountKLine.TokenAUSDForContract).Abs().Add(swapCountKLine.TokenBVolume.Mul(swapCountKLine.TokenBUSDForContract)).Abs()
					}

					if err = updateSwapCountKline(ctx, swapCountKLine, t); err != nil {
						return errors.Wrap(err)
					}
				}

				return nil
			}

			if err = model.Transaction(context.TODO(), trans); err != nil {
				logger.Error("transaction error", logger.Errorv(err))
				return errors.Wrap(err)
			}
		}

		beginID = syncMigrateID

	}

	return nil
}
