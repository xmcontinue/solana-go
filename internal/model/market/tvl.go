package model

import (
	"context"
	"database/sql"
	"time"

	dbPool "git.cplus.link/go/akit/client/psql"
	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	"git.cplus.link/go/akit/util/decimal"
	sq "github.com/Masterminds/squirrel"
	"gorm.io/gorm"
	"gorm.io/hints"

	"git.cplus.link/crema/backend/pkg/domain"
)

func QuerySwapTransactions(ctx context.Context, limit, offset int, filter ...Filter) ([]*domain.SwapTransaction, error) {
	var (
		db   = rDB(ctx)
		list []*domain.SwapTransaction

		err error
	)

	if err = db.Model(&domain.SwapTransaction{}).Scopes(filter...).Limit(limit).Offset(offset).Scan(&list).Error; err != nil {
		return nil, errors.Wrap(err)
	}

	return list, nil
}

func QuerySwapTransactionsV2InBaseTable(ctx context.Context, limit, offset int, filter ...Filter) ([]*domain.SwapTransactionV2, error) {
	var (
		db   = rDB(ctx)
		list []*domain.SwapTransactionV2

		err error
	)

	if err = db.Clauses(hints.Comment("select", "nosharding")).Model(&domain.SwapTransactionV2{}).Scopes(filter...).Limit(limit).Offset(offset).Scan(&list).Error; err != nil {
		return nil, errors.Wrap(err)
	}

	return list, nil
}

func CreatedSwapTransactionsV2(ctx context.Context, transactionV2 []*domain.SwapTransactionV2) error {
	if err := wDB(ctx).Create(transactionV2).Error; err != nil {
		if dbPool.IsDuplicateKeyError(err) {
			return errors.Wrap(errors.AlreadyExists)
		}
		return errors.Wrap(err)
	}
	return nil
}

func UpdateSwapTransactionsV2InBaseTable(ctx context.Context, updates map[string]interface{}, filter ...Filter) error {
	db := wDB(ctx).Clauses(hints.Comment("update", "nosharding")).Model(&domain.SwapTransactionV2{}).Scopes(filter...).Updates(updates)
	if err := db.Error; err != nil {
		return errors.Wrap(err)
	}
	return nil
}

type SwapVol struct {
	SwapAddress   string          `json:"swap_address"`
	TokenAVolume  decimal.Decimal `json:"token_a_volume"`
	TokenBVolume  decimal.Decimal `json:"token_b_volume"`
	TokenAAddress string          `json:"token_a_address"`
	TokenBAddress string          `json:"token_b_address"`
	Vol           decimal.Decimal `json:"vol"`
	Num           int             `json:"num"`
}

// SumSwapAccountLast24Vol 最近24小时 swap account 的总交易额
func SumSwapAccountLast24Vol(ctx context.Context, filter ...Filter) ([]*SwapVol, error) {
	var (
		err     error
		db      = rDB(ctx)
		swapVol []*SwapVol
	)

	if err = db.Model(&domain.SwapTransaction{}).Scopes(filter...).
		Select("sum(token_a_volume) as token_a_volume,sum(token_b_volume) as token_b_volume,count(*) as num,swap_address,token_a_address,token_b_address").
		Group("swap_address,token_a_address,token_b_address").Find(&swapVol).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.Wrap(errors.RecordNotFound)
		}
		return nil, errors.Wrap(err)
	}

	return swapVol, nil
}

type UserSwapVol struct {
	UserAddress   string          `json:"user_address"`
	TokenAVolume  decimal.Decimal `json:"token_a_volume"`
	TokenBVolume  decimal.Decimal `json:"token_b_volume"`
	SwapAddress   string          `json:"swap_address"`
	TokenAAddress string          `json:"token_a_address"`
	TokenBAddress string          `json:"token_b_address"`
	Vol           decimal.Decimal `json:"vol"`
	Num           int             `json:"num"`
}

func UpsertSwapCount(ctx context.Context, swapCount *domain.SwapCountSharding) (*domain.SwapCountSharding, error) {
	var (
		after   domain.SwapCountSharding
		now     = time.Now().UTC()
		inserts = map[string]interface{}{
			"last_swap_transaction_id": swapCount.LastSwapTransactionID,
			"swap_address":             swapCount.SwapAddress,
			"token_a_address":          swapCount.TokenAAddress,
			"token_b_address":          swapCount.TokenBAddress,
			"token_a_volume":           swapCount.TokenAVolume.Abs(),
			"token_b_volume":           swapCount.TokenBVolume.Abs(),
			"token_a_balance":          swapCount.TokenABalance,
			"token_b_balance":          swapCount.TokenBBalance,
			"updated_at":               &now,
			"created_at":               &now,
			"tx_num":                   1,
		}
	)

	sqlStem, args, err := sq.Insert("swap_count_shardings").SetMap(inserts).Suffix("ON CONFLICT(swap_address) DO UPDATE SET").
		Suffix("last_swap_transaction_id = ?,", swapCount.LastSwapTransactionID).
		Suffix("token_a_volume = swap_count_shardings.token_a_volume + ?,", swapCount.TokenAVolume.Abs()).
		Suffix("token_b_volume = swap_count_shardings.token_b_volume + ?,", swapCount.TokenBVolume.Abs()).
		Suffix("token_a_balance = ?,", swapCount.TokenABalance).
		Suffix("token_b_balance = ?,", swapCount.TokenBBalance).
		Suffix("tx_num = swap_count_shardings.tx_num + 1").
		Suffix("WHERE ").
		Suffix("swap_count_shardings.last_swap_transaction_id <= ?", swapCount.LastSwapTransactionID).
		Suffix("RETURNING *").
		ToSql()

	if err != nil {
		return nil, errors.Wrap(err)
	}

	res := wDB(ctx).Raw(sqlStem, args...).Scan(&after)
	if err = res.Error; err != nil {
		return nil, errors.Wrap(err)
	}

	return &after, nil
}

func UpdateSwapCount(ctx context.Context, updates map[string]interface{}, filter ...Filter) error {
	if err := wDB(ctx).Model(&domain.SwapCountSharding{}).Scopes(filter...).Updates(updates).Error; err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func UpsertSwapCountKLine(ctx context.Context, swapCount *domain.SwapCountKLine, tokenABalance, tokenBBalance decimal.Decimal, maxBlockTimeWithDateType *time.Time) (*domain.SwapCountKLine, error) {
	var (
		after   domain.SwapCountKLine
		now     = time.Now().UTC()
		avgFmt  string
		inserts = map[string]interface{}{
			"last_swap_transaction_id":      swapCount.LastSwapTransactionID,
			"swap_address":                  swapCount.SwapAddress,
			"token_a_address":               swapCount.TokenAAddress,
			"token_b_address":               swapCount.TokenBAddress,
			"token_a_volume":                swapCount.TokenAVolume.Abs(),
			"token_b_volume":                swapCount.TokenBVolume.Abs(),
			"token_a_quote_volume":          swapCount.TokenAQuoteVolume.Abs(),
			"token_b_quote_volume":          swapCount.TokenBQuoteVolume.Abs(),
			"token_a_balance":               tokenABalance,
			"token_b_balance":               tokenBBalance,
			"token_a_ref_amount":            swapCount.TokenARefAmount.Abs(),
			"token_a_fee_amount":            swapCount.TokenAFeeAmount.Abs(),
			"token_a_protocol_amount":       swapCount.TokenAProtocolAmount.Abs(),
			"token_b_ref_amount":            swapCount.TokenBRefAmount.Abs(),
			"token_b_fee_amount":            swapCount.TokenBFeeAmount.Abs(),
			"token_b_protocol_amount":       swapCount.TokenBProtocolAmount.Abs(),
			"date_type":                     swapCount.DateType,
			"open":                          swapCount.Open,
			"high":                          swapCount.High,
			"low":                           swapCount.Low,
			"settle":                        swapCount.Settle,
			"avg":                           swapCount.Avg,
			"updated_at":                    &now,
			"created_at":                    &now,
			"date":                          swapCount.Date,
			"tx_num":                        swapCount.TxNum,
			"token_a_usd":                   swapCount.TokenAUSD,
			"token_b_usd":                   swapCount.TokenBUSD,
			"token_a_symbol":                swapCount.TokenASymbol,
			"token_b_symbol":                swapCount.TokenBSymbol,
			"tvl_in_usd":                    swapCount.TvlInUsd,
			"vol_in_usd":                    swapCount.VolInUsd,
			"vol_in_usd_for_contract":       swapCount.VolInUsdForContract,
			"token_ausd_for_contract":       swapCount.TokenAUSDForContract,
			"token_busd_for_contract":       swapCount.TokenBUSDForContract,
			"max_block_time_with_date_type": maxBlockTimeWithDateType,
		}
	)
	//fullName := "swap_count_k_lines"
	fullName, err := getTableFullName(after.TableName(), swapCount.SwapAddress)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	logger.Info("测试", logger.String(swapCount.SwapAddress, "11"), logger.String(string(swapCount.DateType), fullName))
	// 除了domain.DateMin 类型，其他的都是根据前一个类型求平均值
	if swapCount.DateType == domain.DateMin {
		avgFmt = "avg = (" + fullName + ".avg * " + fullName + ".tx_num + ? )/(" + fullName + ".tx_num+ 1)"
	} else {
		avgFmt = "avg = ?"
	}

	sqlStem, args, err := sq.Insert("swap_count_k_lines").SetMap(inserts).Suffix("ON CONFLICT(swap_address,date,date_type) DO UPDATE SET").
		Suffix("last_swap_transaction_id = ?,", swapCount.LastSwapTransactionID).
		Suffix("token_a_volume = "+fullName+".token_a_volume + ?,", swapCount.TokenAVolume.Abs()).
		Suffix("token_b_volume = "+fullName+".token_b_volume + ?,", swapCount.TokenBVolume.Abs()).
		Suffix("token_a_quote_volume = "+fullName+".token_a_quote_volume + ?,", swapCount.TokenAQuoteVolume.Abs()).
		Suffix("token_b_quote_volume = "+fullName+".token_b_quote_volume + ?,", swapCount.TokenBQuoteVolume.Abs()).
		Suffix("token_a_balance = ?,", tokenABalance).
		Suffix("token_b_balance = ?,", tokenBBalance).
		Suffix("token_a_ref_amount = "+fullName+".token_a_ref_amount + ?,", swapCount.TokenARefAmount.Abs()).
		Suffix("token_a_fee_amount = "+fullName+".token_a_fee_amount + ?,", swapCount.TokenAFeeAmount.Abs()).
		Suffix("token_a_protocol_amount = "+fullName+".token_a_protocol_amount + ?,", swapCount.TokenAProtocolAmount.Abs()).
		Suffix("token_b_ref_amount = "+fullName+".token_b_ref_amount + ?,", swapCount.TokenBRefAmount.Abs()).
		Suffix("token_b_fee_amount = "+fullName+".token_b_fee_amount + ?,", swapCount.TokenBFeeAmount.Abs()).
		Suffix("token_b_protocol_amount = "+fullName+".token_b_protocol_amount + ?,", swapCount.TokenBProtocolAmount.Abs()).
		Suffix("high = ?,", swapCount.High).
		Suffix("low = ?,", swapCount.Low).
		Suffix("settle = ?,", swapCount.Settle).
		Suffix("token_a_usd = ?,", swapCount.TokenAUSD). // 不求平均值是因为价格本身就是一分钟更新一次，在一分钟内，其值都是相同的，不用求平均值了
		Suffix("token_b_usd = ?,", swapCount.TokenBUSD).
		Suffix("tvl_in_usd = ?,", swapCount.TvlInUsd).
		Suffix("max_block_time_with_date_type = ?,", maxBlockTimeWithDateType).
		Suffix("tx_num = "+fullName+".tx_num + ?,", swapCount.TxNum).
		Suffix("vol_in_usd = "+fullName+".vol_in_usd + ?,", swapCount.VolInUsd).
		Suffix("vol_in_usd_for_contract = "+fullName+".vol_in_usd_for_contract + ?,", swapCount.VolInUsdForContract).
		Suffix(avgFmt, swapCount.Avg).
		Suffix("WHERE ").
		Suffix(""+fullName+".last_swap_transaction_id < ?", swapCount.LastSwapTransactionID).
		Suffix("RETURNING *").
		ToSql()

	if err != nil {
		return nil, errors.Wrap(err)
	}
	logger.Info("测试", logger.String(swapCount.SwapAddress, "12"), logger.String(string(swapCount.DateType), fullName))
	res := wDB(ctx).Raw(sqlStem, args...).Scan(&after)
	if err = res.Error; err != nil {
		return nil, errors.Wrap(err)
	}

	//if res.RowsAffected == 0 {
	//	swapCountKLine, _ := QuerySwapCountKLine(ctx, SwapAddressFilter(swapCount.SwapAddress), NewFilter("date = ?", swapCount.DateType), NewFilter("date_type = ?", swapCount.DateType))
	//	fmt.Println("RowsAffected=0", swapCount.SwapAddress, swapCount.LastSwapTransactionID, swapCountKLine.LastSwapTransactionID, swapCountKLine.Date, swapCountKLine.DateType)
	//} else {
	//
	//	fmt.Println("RowsAffected!=0", swapCount.SwapAddress, swapCount.LastSwapTransactionID, after.LastSwapTransactionID)
	//}

	return &after, nil
}

func UpsertUserSwapCountKLine(ctx context.Context, userSwapCount *domain.UserCountKLine) (*domain.UserCountKLine, error) {
	var (
		after   domain.UserCountKLine
		now     = time.Now().UTC()
		inserts = map[string]interface{}{
			"last_swap_transaction_id":          userSwapCount.LastSwapTransactionID,
			"user_address":                      userSwapCount.UserAddress,
			"swap_address":                      userSwapCount.SwapAddress,
			"token_a_address":                   userSwapCount.TokenAAddress,
			"token_b_address":                   userSwapCount.TokenBAddress,
			"date_type":                         userSwapCount.DateType,
			"token_a_symbol":                    userSwapCount.TokenASymbol,
			"token_b_symbol":                    userSwapCount.TokenBSymbol,
			"user_token_a_volume":               userSwapCount.UserTokenAVolume.Abs(),
			"user_token_b_volume":               userSwapCount.UserTokenBVolume.Abs(),
			"token_a_quote_volume":              userSwapCount.TokenAQuoteVolume,
			"token_b_quote_volume":              userSwapCount.TokenBQuoteVolume,
			"updated_at":                        &now,
			"created_at":                        &now,
			"tx_num":                            1,
			"date":                              userSwapCount.Date,
			"token_a_withdraw_liquidity_volume": userSwapCount.TokenAWithdrawLiquidityVolume,
			"token_b_withdraw_liquidity_volume": userSwapCount.TokenBWithdrawLiquidityVolume,
			"token_a_deposit_liquidity_volume":  userSwapCount.TokenADepositLiquidityVolume,
			"token_b_deposit_liquidity_volume":  userSwapCount.TokenBDepositLiquidityVolume,
			"token_a_claim_volume":              userSwapCount.TokenAClaimVolume,
			"token_b_claim_volume":              userSwapCount.TokenBClaimVolume,
		}
	)

	sqlStem, args, err := sq.Insert("user_count_k_lines").SetMap(inserts).Suffix("ON CONFLICT(user_address,date,date_type,swap_address) DO UPDATE SET").
		Suffix("last_swap_transaction_id = ?,", userSwapCount.LastSwapTransactionID).
		Suffix("user_token_a_volume = user_count_k_lines.user_token_a_volume + ?,", userSwapCount.UserTokenAVolume.Abs()).
		Suffix("user_token_b_volume = user_count_k_lines.user_token_b_volume + ?,", userSwapCount.UserTokenBVolume.Abs()).
		Suffix("token_a_quote_volume = user_count_k_lines.token_a_quote_volume + ?,", userSwapCount.TokenAQuoteVolume).
		Suffix("token_b_quote_volume = user_count_k_lines.token_b_quote_volume + ?,", userSwapCount.TokenBQuoteVolume).
		Suffix("token_a_withdraw_liquidity_volume = user_count_k_lines.token_a_withdraw_liquidity_volume + ?,", userSwapCount.TokenAWithdrawLiquidityVolume).
		Suffix("token_b_withdraw_liquidity_volume = user_count_k_lines.token_b_withdraw_liquidity_volume + ?,", userSwapCount.TokenBWithdrawLiquidityVolume).
		Suffix("token_a_deposit_liquidity_volume = user_count_k_lines.token_a_deposit_liquidity_volume + ?,", userSwapCount.TokenADepositLiquidityVolume).
		Suffix("token_b_deposit_liquidity_volume = user_count_k_lines.token_b_deposit_liquidity_volume + ?,", userSwapCount.TokenBDepositLiquidityVolume).
		Suffix("token_a_claim_volume = user_count_k_lines.token_a_claim_volume + ?,", userSwapCount.TokenAClaimVolume).
		Suffix("token_b_claim_volume = user_count_k_lines.token_b_claim_volume + ?,", userSwapCount.TokenBClaimVolume).
		Suffix("tx_num = user_count_k_lines.tx_num +1").
		Suffix("WHERE ").
		Suffix("user_count_k_lines.last_swap_transaction_id <= ?", userSwapCount.LastSwapTransactionID).
		Suffix("RETURNING *").
		ToSql()

	if err != nil {
		return nil, errors.Wrap(err)
	}

	res := wDB(ctx).Raw(sqlStem, args...).Scan(&after)
	if err = res.Error; err != nil {
		return nil, errors.Wrap(err)
	}

	return &after, nil
}

func QueryUserSwapCountDay(ctx context.Context, limit, offset int, filter ...Filter) ([]*domain.UserCountKLine, int64, error) {
	var (
		db    = rDB(ctx)
		list  []*domain.UserCountKLine
		total int64
		err   error
	)
	if err = db.Model(&domain.UserCountKLine{}).Scopes(filter...).Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(err)
	}

	if total == 0 {
		return list, 0, nil
	}

	if err = db.Scopes(filter...).Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, errors.Wrap(err)
	}

	return list, total, nil

}

func QueryPositions(ctx context.Context, limit, offset int, filter ...Filter) ([]*domain.PositionCountSnapshot, error) {
	var (
		db   = rDB(ctx)
		list []*domain.PositionCountSnapshot
		err  error
	)

	if err = db.Scopes(filter...).Limit(limit).Offset(offset).Order("id asc").Find(&list).Error; err != nil {
		return nil, errors.Wrap(err)
	}

	return list, nil

}

func GetLastMaxTvls(ctx context.Context, filter ...Filter) ([]*domain.SwapCountSharding, error) {
	var ids []int64
	if err := wDB(ctx).Model(&domain.SwapCountSharding{}).Scopes(filter...).Select("max(last_swap_transaction_id)").Group("swap_address").Scan(&ids).Error; err != nil {
		return nil, errors.Wrap(err)
	}

	if ids == nil {
		return nil, nil
	}

	var list []*domain.SwapCountSharding
	wDB(ctx).Model(&domain.SwapCountSharding{}).Select("swap_address").Where("id in ?", ids).Scan(&list)
	return list, nil
}

func QuerySwapCount(ctx context.Context, filter ...Filter) (*domain.SwapCountSharding, error) {
	var swapCount = &domain.SwapCountSharding{}
	if err := rDB(ctx).Model(&domain.SwapCountSharding{}).Scopes(filter...).Take(swapCount).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.Wrap(errors.RecordNotFound)
		}
		return nil, errors.Wrap(err)
	}
	return swapCount, nil

}

func CreateSwapCount(ctx context.Context, swapCount *domain.SwapCountSharding) error {
	if err := wDB(ctx).Create(swapCount).Error; err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func QuerySwapCountMigrate(ctx context.Context, filter ...Filter) (*domain.SwapCountMigrate, error) {
	var count = &domain.SwapCountMigrate{}
	if err := rDB(ctx).Model(&domain.SwapCountMigrate{}).Scopes(filter...).Take(count).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.Wrap(errors.RecordNotFound)
		}
		return nil, errors.Wrap(err)
	}
	return count, nil

}

func CreateSwapCountMigrate(ctx context.Context, swapCount *domain.SwapCountMigrate) error {
	if err := wDB(ctx).Create(swapCount).Error; err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func UpdateSwapCountMigrate(ctx context.Context, updates map[string]interface{}, filter ...Filter) error {
	if err := wDB(ctx).Model(&domain.SwapCountMigrate{}).Scopes(filter...).Updates(updates).Error; err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func UpdateUserCountKLine(ctx context.Context, updates map[string]interface{}, filter ...Filter) error {
	db := wDB(ctx).Model(&domain.UserCountKLine{}).Scopes(filter...).Updates(updates)
	if err := db.Error; err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func GetMaxUserCountKLineID(ctx context.Context, swapAccount string) (int64, error) {
	var max sql.NullInt64
	if err := rDB(ctx).Model(&domain.UserCountKLine{}).Select("last_swap_transaction_id").Scopes(SwapAddressFilter(swapAccount), NewFilter("date_type = 'mon'")).Order("last_swap_transaction_id desc").Take(&max).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.Wrap(errors.RecordNotFound)
		}
		return 0, errors.Wrap(err)
	}
	if max.Valid {
		return max.Int64, nil
	}

	return 0, nil
}
func UpdateSwapCountBySwapAccount(ctx context.Context, swapAccount string, updates map[string]interface{}, filter ...Filter) error {
	if err := wDB(ctx).Model(&domain.SwapCountSharding{}).Scopes(append(filter, SwapAddressFilter(swapAccount))...).Updates(updates).Error; err != nil {
		if dbPool.IsDuplicateKeyError(err) {
			return errors.Wrap(errors.AlreadyExists)
		}
		return errors.Wrap(err)
	}
	return nil
}

func QuerySwapCountKLine(ctx context.Context, filter ...Filter) (*domain.SwapCountKLine, error) {
	var (
		db             = rDB(ctx)
		err            error
		swapCountKLine = &domain.SwapCountKLine{}
	)

	if err = db.Model(&domain.SwapCountKLine{}).Scopes(filter...).Take(swapCountKLine).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.Wrap(errors.RecordNotFound)
		}
		return nil, errors.Wrap(err)
	}

	return swapCountKLine, nil
}

func QuerySwapCountKLines(ctx context.Context, limit, offset int, filter ...Filter) ([]*domain.SwapCountKLine, error) {
	var (
		db             = rDB(ctx)
		err            error
		swapCountKLine []*domain.SwapCountKLine
	)

	if err = db.Model(&domain.SwapCountKLine{}).Scopes(filter...).Limit(limit).Offset(offset).Scan(&swapCountKLine).Error; err != nil {
		return nil, errors.Wrap(err)
	}

	return swapCountKLine, nil
}

func QuerySwapCountKLinesInBaseTable(ctx context.Context, limit, offset int, filter ...Filter) ([]*domain.SwapCountKLine, error) {
	var (
		db             = rDB(ctx)
		err            error
		swapCountKLine []*domain.SwapCountKLine
	)

	if err = db.Clauses(hints.Comment("select", "nosharding")).Model(&domain.SwapCountKLine{}).Scopes(filter...).Limit(limit).Offset(offset).Scan(&swapCountKLine).Error; err != nil {
		return nil, errors.Wrap(err)
	}

	return swapCountKLine, nil
}

func UpdateSwapCountKLinesInBaseTable(ctx context.Context, updates map[string]interface{}, filter ...Filter) error {
	if err := wDB(ctx).Clauses(hints.Comment("update", "nosharding")).Model(&domain.SwapCountKLine{}).Scopes(filter...).Updates(updates).Error; err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func CreateSwapCountKLine(ctx context.Context, kLines []*domain.SwapCountKLine) error {
	if err := wDB(ctx).Create(kLines).Error; err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func DeleteSwapCountKLines(ctx context.Context, filter ...Filter) error {
	res := wDB(ctx).Scopes(filter...).Delete(&domain.SwapCountKLine{})
	if err := res.Error; err != nil {
		return errors.Wrap(err)
	}
	if res.RowsAffected <= 0 {
		return errors.Wrap(errors.RecordNotFound)
	}
	return nil
}

type DateAndPrice struct {
	Tvl  decimal.Decimal
	Date time.Time
}

func SumTvlPriceInUSD(ctx context.Context, limit, offset int, filter ...Filter) ([]*DateAndPrice, error) {
	var (
		db           = rDB(ctx)
		err          error
		dateAndPrice []*DateAndPrice
	)

	if err = db.Model(&domain.SwapCountKLine{}).Select("sum(token_a_usd*token_a_balance) as tvl,date").Scopes(filter...).Group("date").Limit(limit).Offset(offset).Scan(&dateAndPrice).Error; err != nil {
		return nil, errors.Wrap(err)
	}
	return dateAndPrice, nil
}

// SumSwapCountVolForKLines ...
func SumSwapCountVolForKLines(ctx context.Context, filter ...Filter) (*domain.SwapCountKLineVolCount, error) {
	var (
		db                     = rDB(ctx)
		err                    error
		swapCountKLineVolCount = &domain.SwapCountKLineVolCount{}
	)

	if err = db.Model(&domain.SwapCountKLine{}).Select("SUM(token_a_volume) as token_a_volume, SUM(token_b_volume) as token_b_volume," +
		" SUM(token_a_quote_volume) as token_a_quote_volume, SUM(token_b_quote_volume) as token_b_quote_volume," +
		"SUM(token_a_volume*token_ausd_for_contract) as token_a_volume_for_usd, " +
		"SUM(token_b_volume*token_busd_for_contract) as token_b_volume_for_usd, " +
		"SUM(token_a_quote_volume*token_ausd_for_contract) as token_a_quote_volume_for_usd, " +
		"SUM(token_b_quote_volume*token_busd_for_contract) as token_b_quote_volume_for_usd, SUM(tx_num) as tx_num," +
		"SUM(token_a_fee_amount*token_ausd_for_contract + token_b_fee_amount*token_busd_for_contract) as fee_amount," +
		"SUM(token_a_ref_amount*token_ausd_for_contract + token_b_ref_amount*token_busd_for_contract) as ref_amount," +
		"SUM(vol_in_usd_for_contract) as vol_in_usd_for_contract," +
		"SUM(token_a_protocol_amount*token_ausd_for_contract + token_b_protocol_amount*token_busd_for_contract) as protocol_amount," +
		"COUNT(*) as day_num").Scopes(filter...).Scan(&swapCountKLineVolCount).Error; err != nil {
		return nil, errors.Wrap(err)
	}

	return swapCountKLineVolCount, nil
}
