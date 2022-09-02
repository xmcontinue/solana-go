package process

//// SwapV1 同步更新swap_counts表和user_swap_counts表
//type SwapV1 struct {
//	ID                int64
//	LastTransactionID int64
//	SwapAccount       string
//	SwapRecords       []*parse.SwapRecord
//	tx                *parse.Tx
//	BlockDate         *time.Time
//	spec              string
//	swapR             []parse.SwapRecordIface
//}

//
//// ParserDate 按照区块时间顺序解析
//func (s *SwapV1) ParserDate() error {
//	for {
//		swapCount, err := model.QuerySwapCount(context.TODO(), model.SwapAddressFilter(s.SwapAccount))
//		if err != nil && !errors.Is(err, errors.RecordNotFound) {
//			return errors.Wrap(err)
//		}
//
//		if swapCount != nil {
//			s.ID = swapCount.LastSwapTransactionID
//		}
//
//		filters := []model.Filter{
//			model.NewFilter("id > ?", s.ID),
//			model.NewFilter("id <= ?", s.LastTransactionID),
//			model.SwapAddressFilter(s.SwapAccount),
//			model.OrderFilter("id asc"),
//		}
//
//		swapTransactions, err := model.QuerySwapTransactions(context.TODO(), 100, 0, filters...)
//		if err != nil {
//			logger.Error("get single transaction err", logger.Errorv(err))
//			return errors.Wrap(err)
//		}
//
//		if len(swapTransactions) == 0 {
//			logger.Info(fmt.Sprintf("parse swap, swap address: %s , current id is %d, target id is %d", s.SwapAccount, s.ID, s.LastTransactionID))
//			break
//		}
//
//		for _, transaction := range swapTransactions {
//			s.ID = transaction.ID
//			tx := parse.NewTx(transaction.TxData)
//			err = tx.ParseTxToSwap()
//			if err != nil {
//				if errors.Is(err, errors.RecordNotFound) {
//					continue
//				}
//				logger.Error("sync transaction id err", logger.Errorv(err))
//				return errors.Wrap(err)
//			}
//
//			s.SwapRecords = tx.SwapRecords
//			s.BlockDate = transaction.BlockTime
//
//			if err = s.WriteToDB(transaction.TokenAUSD, transaction.TokenBUSD); err != nil {
//				logger.Error("write to db error:", logger.Errorv(err))
//				return errors.Wrap(err)
//			}
//		}
//
//		// 更新处理数据的位置
//		if err = model.UpdateSwapCountBySwapAccount(context.TODO(), s.SwapAccount, map[string]interface{}{"last_swap_transaction_id": s.ID}); err != nil {
//			return errors.Wrap(err)
//		}
//
//		logger.Info(fmt.Sprintf("parse swap, swap address: %s , current id is %d, target id is %d", s.SwapAccount, s.ID, s.LastTransactionID))
//
//	}
//
//	return nil
//}
//
//func (s *SwapV1) WriteToDB1(tokenAUSD, tokenBUSD decimal.Decimal) error {
//	defer func() {
//		if err := recover(); err != nil {
//			fmt.Println("Recovered in f", err)
//		}
//	}()
//
//	var err error
//	trans := func(ctx context.Context) error {
//		for _, swapRecord := range s.SwapRecords {
//			// 仅当前swapAccount  可以插入
//			if swapRecord.SwapConfig.SwapAccount != s.SwapAccount {
//				continue
//			}
//
//			if err = s.updateSwapCount(ctx, swapRecord); err != nil {
//				return errors.Wrap(err)
//			}
//
//			var (
//				tokenAVolume      decimal.Decimal
//				tokenBVolume      decimal.Decimal
//				tokenAQuoteVolume decimal.Decimal
//				tokenBQuoteVolume decimal.Decimal
//			)
//			if swapRecord.Direction == 0 {
//				tokenAVolume = swapRecord.TokenCount.TokenAVolume
//				tokenBQuoteVolume = swapRecord.TokenCount.TokenBVolume
//			} else {
//				tokenBVolume = swapRecord.TokenCount.TokenBVolume
//				tokenAQuoteVolume = swapRecord.TokenCount.TokenAVolume
//			}
//
//			swapCountKLine := &domain.SwapCountKLine{
//				LastSwapTransactionID: s.ID,
//				SwapAddress:           swapRecord.SwapConfig.SwapAccount,
//				TokenAAddress:         swapRecord.SwapConfig.TokenA.SwapTokenAccount,
//				TokenBAddress:         swapRecord.SwapConfig.TokenB.SwapTokenAccount,
//				TokenAVolume:          tokenAVolume,
//				TokenBVolume:          tokenBVolume,
//				TokenAQuoteVolume:     tokenAQuoteVolume,
//				TokenBQuoteVolume:     tokenBQuoteVolume,
//				TokenABalance:         swapRecord.TokenCount.TokenABalance,
//				TokenBBalance:         swapRecord.TokenCount.TokenBBalance,
//				DateType:              domain.DateMin,
//				Open:                  swapRecord.Price,
//				High:                  swapRecord.Price,
//				Low:                   swapRecord.Price,
//				Avg:                   swapRecord.Price,
//				Settle:                swapRecord.Price,
//				Date:                  s.BlockDate,
//				TxNum:                 1,
//				TokenAUSD:             tokenAUSD,
//				TokenBUSD:             tokenBUSD,
//				TokenASymbol:          swapRecord.SwapConfig.TokenA.Symbol,
//				TokenBSymbol:          swapRecord.SwapConfig.TokenB.Symbol,
//				TvlInUsd:              swapRecord.TokenCount.TokenABalance.Mul(tokenAUSD).Add(swapRecord.TokenCount.TokenBBalance.Mul(tokenBUSD)),
//				VolInUsd:              tokenAVolume.Mul(tokenAUSD).Abs().Add(tokenBVolume.Mul(tokenBUSD)).Abs(),
//			}
//
//			newKline := kline.NewKline(s.BlockDate)
//			for _, t := range newKline.Types {
//				swapCountKLine.DateType = t.DateType
//				// 获取价格
//				tokenAPrice, tokenBPrice, err := PriceToSwapKLineHandle(ctx, swapCountKLine)
//				if err != nil {
//					return errors.Wrap(err)
//				}
//				swapCountKLine.TokenAUSDForContract = tokenAPrice
//				swapCountKLine.TokenBUSDForContract = tokenBPrice
//
//				if err = UpdateSwapCountKline(ctx, swapCountKLine, t); err != nil {
//					return errors.Wrap(err)
//				}
//			}
//		}
//		return nil
//	}
//
//	if err = model.Transaction(context.TODO(), trans); err != nil {
//		logger.Error("transaction error", logger.Errorv(err))
//		return errors.Wrap(err)
//	}
//	return nil
//}

//
//func (m *KLineTyp) updateKline(ctx context.Context, swapCountKLine *domain.SwapCountKLine) error {
//	swapCountKLine.Date = m.GetDate()
//	swapCountKLine.DateType = m.DateType
//	currentSwapCountKLine, err := model.QuerySwapCountKLine(ctx,
//		model.NewFilter("swap_address = ?", swapCountKLine.SwapAddress),
//		model.NewFilter("date = ?", swapCountKLine.Date),
//		model.NewFilter("date_type = ?", swapCountKLine.DateType))
//
//	if err != nil && !errors.Is(err, errors.RecordNotFound) {
//		return errors.Wrap(err)
//	}
//
//	if currentSwapCountKLine != nil {
//		if currentSwapCountKLine.High.GreaterThan(swapCountKLine.High) {
//			swapCountKLine.High = currentSwapCountKLine.High
//		}
//		if currentSwapCountKLine.Low.LessThan(swapCountKLine.Low) {
//			swapCountKLine.Low = currentSwapCountKLine.Low
//		}
//	}
//
//	if m.DateType != domain.DateMin {
//		avg, err := m.calculateAvg(ctx, swapCountKLine.SwapAddress)
//		if err != nil {
//			return errors.Wrap(err)
//		}
//
//		swapCountKLine.Avg = avg
//	}
//
//	_, err = model.UpsertSwapCountKLine(ctx, swapCountKLine, decimal.Zero, decimal.Zero, nil)
//	if err != nil {
//		return errors.Wrap(err)
//	}
//
//	return nil
//}
//
//// calculateAvg 按照上一个周期计算平均值，month除外（按照天计算）
//func (m *KLineTyp) calculateAvg(ctx context.Context, swapAccount string) (decimal.Decimal, error) {
//	type interTime struct {
//		Date time.Time
//		avg  *decimal.Decimal
//	}
//
//	var (
//		count     = int32(0)
//		sum       = decimal.Zero
//		beginTime time.Time
//		endTime   time.Time
//	)
//
//	avgList := make([]*interTime, m.Interval, m.Interval)
//	if m.DateType != domain.DateMin {
//		beginTime = *m.GetDate()
//		if m.DateType == domain.DateMon {
//			lastDateTime := m.Date.AddDate(0, 1, -m.Date.Day())
//			endTime = time.Date(lastDateTime.Year(), lastDateTime.Month(), lastDateTime.Day(), 0, 0, 0, 0, m.Date.Location()).Add(time.Hour * 24)
//		} else {
//			endTime = beginTime.Add(m.InnerTimeInterval * time.Duration(m.Interval))
//		}
//
//		swapCountKLines, err := model.QuerySwapCountKLines(ctx, m.Interval, 0,
//			model.NewFilter("date_type = ?", m.BeforeIntervalDateType),
//			model.SwapAddressFilter(swapAccount),
//			model.NewFilter("date < ?", endTime),
//			model.OrderFilter("date desc"),
//		)
//
//		if err != nil {
//			return decimal.Zero, errors.Wrap(err)
//		}
//
//		for index := range avgList {
//			avgList[index] = &interTime{
//				Date: m.GetDate().Add(m.InnerTimeInterval * time.Duration(index)),
//			}
//		}
//
//		// 减少for 循环
//		swapCountKLineMap := make(map[int64]*domain.SwapCountKLine, len(swapCountKLines))
//		for index := range swapCountKLines {
//			swapCountKLineMap[swapCountKLines[index].Date.Unix()] = swapCountKLines[index]
//		}
//
//		// 找到第一个数据
//		lastAvg := &domain.SwapCountKLine{}
//		for index := range swapCountKLines {
//			if swapCountKLines[len(swapCountKLines)-index-1].Date.After(avgList[0].Date) {
//				break
//			}
//			lastAvg = swapCountKLines[len(swapCountKLines)-index-1]
//		}
//
//		for index, avg := range avgList {
//			lastSwapCountKLine, ok := swapCountKLineMap[avg.Date.Unix()]
//			if ok {
//				lastAvg = lastSwapCountKLine
//				avgList[index].avg = &lastSwapCountKLine.Avg
//			} else {
//				avgList[index].avg = &lastAvg.Settle // 上一个周期的结束值用作空缺周期的平均值
//			}
//		}
//
//		// calculate avg
//		for _, v := range avgList {
//			if !v.avg.IsZero() {
//				sum = sum.Add(*v.avg)
//				count++
//			}
//		}
//	} else {
//		return decimal.Zero, errors.New("error date_type")
//	}
//
//	return sum.Div(decimal.NewFromInt32(count)), nil
//}

//func (s *SwapV1) updateSwapCount(ctx context.Context, swapRecord *parse.SwapRecord) error {
//	swapCount := &domain.SwapCount{
//		LastSwapTransactionID: s.ID,
//		SwapAddress:           swapRecord.SwapConfig.SwapAccount,
//		TokenAAddress:         swapRecord.SwapConfig.TokenA.SwapTokenAccount,
//		TokenBAddress:         swapRecord.SwapConfig.TokenB.SwapTokenAccount,
//		TokenABalance:         swapRecord.TokenCount.TokenABalance,
//		TokenBBalance:         swapRecord.TokenCount.TokenBBalance,
//	}
//
//	_, err := model.UpsertSwapCount(ctx, swapCount)
//	if err != nil {
//		return errors.Wrap(err)
//	}
//
//	return nil
//}

//func updateSwapCount(ctx context.Context, swapCount *domain.SwapCount) error {
//	_, err := model.UpsertSwapCount(ctx, swapCount)
//	if err != nil {
//		return errors.Wrap(err)
//	}
//
//	return nil
//}

//
//func (s *SwapV1) WriteToDB(tokenAUSD, tokenBUSD decimal.Decimal) error {
//	defer func() {
//		if err := recover(); err != nil {
//			fmt.Println("Recovered in f", err)
//		}
//	}()
//
//	var err error
//	trans := func(ctx context.Context) error {
//		for _, swapRecord := range s.swapR {
//			// 仅当前swapAccount  可以插入
//			if swapRecord.GetSwapConfig().SwapAccount != s.SwapAccount {
//				continue
//			}
//
//			if err = updateSwapCount(ctx, &domain.SwapCount{
//				LastSwapTransactionID: s.ID,
//				SwapAddress:           swapRecord.GetSwapConfig().SwapAccount,
//				TokenAAddress:         swapRecord.GetSwapConfig().TokenA.SwapTokenAccount,
//				TokenBAddress:         swapRecord.GetSwapConfig().TokenB.SwapTokenAccount,
//				TokenABalance:         swapRecord.GetTokenABalance(),
//				TokenBBalance:         swapRecord.GetTokenBBalance(),
//			}); err != nil {
//				return errors.Wrap(err)
//			}
//
//			var (
//				tokenAVolume      decimal.Decimal
//				tokenBVolume      decimal.Decimal
//				tokenAQuoteVolume decimal.Decimal
//				tokenBQuoteVolume decimal.Decimal
//			)
//			if swapRecord.GetDirection() == 0 {
//				tokenAVolume = swapRecord.GetTokenAVolume()
//				tokenBQuoteVolume = swapRecord.GetTokenBVolume()
//			} else {
//				tokenBVolume = swapRecord.GetTokenBVolume()
//				tokenAQuoteVolume = swapRecord.GetTokenAVolume()
//			}
//
//			swapCountKLine := &domain.SwapCountKLine{
//				LastSwapTransactionID: s.ID,
//				SwapAddress:           swapRecord.GetSwapConfig().SwapAccount,
//				TokenAAddress:         swapRecord.GetSwapConfig().TokenA.SwapTokenAccount,
//				TokenBAddress:         swapRecord.GetSwapConfig().TokenB.SwapTokenAccount,
//				TokenAVolume:          tokenAVolume,
//				TokenBVolume:          tokenBVolume,
//				TokenAQuoteVolume:     tokenAQuoteVolume,
//				TokenBQuoteVolume:     tokenBQuoteVolume,
//				TokenABalance:         swapRecord.GetTokenABalance(),
//				TokenBBalance:         swapRecord.GetTokenBBalance(),
//				DateType:              domain.DateMin,
//				Open:                  swapRecord.GetPrice(),
//				High:                  swapRecord.GetPrice(),
//				Low:                   swapRecord.GetPrice(),
//				Avg:                   swapRecord.GetPrice(),
//				Settle:                swapRecord.GetPrice(),
//				Date:                  s.BlockDate,
//				TxNum:                 1,
//				TokenAUSD:             tokenAUSD,
//				TokenBUSD:             tokenBUSD,
//				TokenASymbol:          swapRecord.GetSwapConfig().TokenA.Symbol,
//				TokenBSymbol:          swapRecord.GetSwapConfig().TokenB.Symbol,
//				TvlInUsd:              swapRecord.GetTokenABalance().Mul(tokenAUSD).Add(swapRecord.GetTokenBBalance().Mul(tokenBUSD)),
//				VolInUsd:              tokenAVolume.Mul(tokenAUSD).Abs().Add(tokenBVolume.Mul(tokenBUSD)).Abs(),
//			}
//
//			newKline := kline.NewKline(s.BlockDate)
//			for _, t := range newKline.Types {
//				swapCountKLine.DateType = t.DateType
//				// 获取价格
//				tokenAPrice, tokenBPrice, err := PriceToSwapKLineHandle(ctx, swapCountKLine)
//				if err != nil {
//					return errors.Wrap(err)
//				}
//				swapCountKLine.TokenAUSDForContract = tokenAPrice
//				swapCountKLine.TokenBUSDForContract = tokenBPrice
//
//				if err = UpdateSwapCountKline(ctx, swapCountKLine, t); err != nil {
//					return errors.Wrap(err)
//				}
//			}
//		}
//		return nil
//	}
//
//	if err = model.Transaction(context.TODO(), trans); err != nil {
//		logger.Error("transaction error", logger.Errorv(err))
//		return errors.Wrap(err)
//	}
//	return nil
//}
