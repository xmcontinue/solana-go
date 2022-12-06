package model

import (
	"context"
	"reflect"
	"sync"

	dbpool "git.cplus.link/go/akit/client/psql"
	"git.cplus.link/go/akit/config"
	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/util/gquery"
	"gorm.io/gorm"

	"git.cplus.link/crema/backend/pkg/domain"
)

var (
	configer *config.Config
	dbRPool  *dbpool.PGPool // 读库
	dbWPool  *dbpool.PGPool // 写库
	once     sync.Once
)

type Filter = func(*gorm.DB) *gorm.DB

// NewFilter 创建新查询条件
func NewFilter(query string, args ...interface{}) Filter {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(query, args...)
	}
}

// 链接数据库
func initDB() error {
	var (
		rConf dbpool.DBConfig
		wConf dbpool.DBConfig
	)
	if err := configer.UnmarshalKey("dbs.read", &rConf); err != nil {
		return errors.Wrap(err)
	}
	dbRPool = dbpool.NewPGPool(&rConf).Assert()

	if err := configer.UnmarshalKey("dbs.write", &wConf); err != nil {
		return errors.Wrap(err)
	}
	dbWPool = dbpool.NewPGPool(&wConf).Assert()

	return nil
}

// 初始化/同步数据库表结构
func autoMigrate() error {
	if err := dbWPool.NewConn().AutoMigrate(
		&domain.SwapPairCountSharding{},
		&domain.Tvl{},
		&domain.SwapTransaction{},
		&domain.SwapTransactionV2{},
		&domain.NetRecode{},
		&domain.SwapCountSharding{},
		&domain.SwapCountMigrate{},
		&domain.SwapCountKLine{},
		&domain.UserCountKLine{},
		&domain.SwapUserCount{},
		&domain.TransActionUserCount{},
		&domain.SwapPairBaseSharding{},
		&domain.SwapPairPriceKLine{},
		&domain.SwapTokenPriceKLine{},
		&domain.ActivityHistory{},
		&domain.PositionCountSnapshot{},
		&domain.PositionV2Snapshot{},
		&domain.MetadataJsonDate{},
	); err != nil {
		return errors.Wrap(err)
	}
	return nil
}

// 从ctx获取数据库事务，可以配合dbpool.WithTx使用传入数据库对象，在调用链中传递, 必须保证ctx, 只存在一个数据库对象，否则会覆盖
func rDB(ctx context.Context) *gorm.DB {
	tx := dbpool.GetTx(ctx)
	if tx == nil {
		return dbRPool.NewConn()
	}
	return tx
}

// 从ctx获取数据库事务，可以配合dbpool.WithTx使用传入数据库对象，在调用链中传递, 必须保证ctx, 只存在一个数据库对象，否则会覆盖
func wDB(ctx context.Context) *gorm.DB {
	tx := dbpool.GetTx(ctx)
	if tx == nil {
		return dbWPool.NewConn()
	}
	return tx
}

// WriteDB 返回数据库可写对象
func WriteDB() *gorm.DB {
	return dbWPool.NewConn()
}

// Init model初始化
func Init(config *config.Config) error {
	var rErr error
	once.Do(func() {
		configer = config
		if err := initDB(); err != nil {
			rErr = errors.Wrapf(err, "init db")
			return
		}
		if err := autoMigrate(); err != nil {
			rErr = errors.Wrapf(err, "auto migrate")
		}
	})

	return rErr
}

// Transaction 开启事物处理
func Transaction(ctx context.Context, f func(context.Context) error) error {
	var (
		tx   = WriteDB().Begin()
		wCtx = dbpool.WithTx(ctx, tx)
	)

	err := f(wCtx)
	if err != nil {
		tx.Rollback()

		return errors.Wrap(err)
	}

	err = tx.Commit().Error
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

// IDFilter ID查询条件生成
func IDFilter(id int64) Filter {
	return NewFilter("id = ?", id)
}

// SwapAddressFilter swapAddress查询条件生成
func SwapAddressFilter(swapAddress string) Filter {
	return NewFilter("swap_address = ?", swapAddress)
}

func DateTypeFilter(dateType domain.DateType) Filter {
	return NewFilter("date_type = ?", dateType)
}

// OrderFilter order查询条件生成
func OrderFilter(by string) Filter {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order(by)
	}
}

// SwapTransferFilter 从transactions中查询swap操作( TODO 延续初版判断swap方法，后续更改算法，instruction_len为((8,26,17,12)代表swap交易，9,41,50,52为其他操作)
func SwapTransferFilter() Filter {
	return NewFilter("instruction_len not in ?", []uint64{9, 41, 50, 52})
}

// GQueryOrderFilter gQuery order查询条件生成
func GQueryOrderFilter(args interface{}, by *gquery.GOrderBy) Filter {
	return func(db *gorm.DB) *gorm.DB {
		objT := reflect.TypeOf(args)
		if objT.Kind() == reflect.Ptr && objT.Elem().Kind() == reflect.Struct {
			objT = objT.Elem()
		}
		order, ok := objT.FieldByName("OrderBy")
		if !ok {
			return db
		}
		tag := order.Tag.Get("gquery")
		return by.SetOrderBy(tag, db)
	}
}

// IDDESCFilter ID降序序过滤
func IDDESCFilter() func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order("id desc")
	}
}
