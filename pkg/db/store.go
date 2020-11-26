package db

import (
	"fmt"
	"time"

	"github.com/qlcchain/qlc-hub/pkg/types"
	"github.com/qlcchain/qlc-hub/pkg/util"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDB(url string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(url), &gorm.Config{
		Logger: logger.Discard,
	})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&types.SwapInfo{})
	return db, nil
}

func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page == 0 {
			page = 1
		}
		//if pageSize == 0 {
		//	pageSize = 20
		//}
		//switch {
		//case pageSize > 100:
		//	pageSize = 100
		//case pageSize <= 0:
		//	pageSize = 10
		//}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

func InsertSwapInfo(db *gorm.DB, record *types.SwapInfo) error {
	record.LastModifyTime = time.Now().Unix()
	record.EthTxHash = util.AddHashPrefix(record.EthTxHash)
	record.NeoTxHash = util.AddHashPrefix(record.NeoTxHash)
	return db.Create(record).Error
}

func UpdateSwapInfo(db *gorm.DB, record *types.SwapInfo) error {
	record.LastModifyTime = time.Now().Unix()
	record.EthTxHash = util.AddHashPrefix(record.EthTxHash)
	record.NeoTxHash = util.AddHashPrefix(record.NeoTxHash)
	return db.Save(record).Error
}

func GetSwapInfos(db *gorm.DB, page, pageSize int) ([]*types.SwapInfo, error) {
	var result []*types.SwapInfo
	if err := db.Scopes(Paginate(page, pageSize)).Order("last_modify_time DESC").Find(&result).Error; err == nil {
		return result, nil
	} else {
		return nil, err
	}
}

func GetSwapInfoByTxHash(db *gorm.DB, hash string, action types.ChainType) (*types.SwapInfo, error) {
	var result types.SwapInfo
	hash = util.AddHashPrefix(hash)
	if action == types.ETH {
		if err := db.Where("eth_tx_hash = ?", hash).First(&result).Error; err == nil {
			return &result, nil
		} else {
			return nil, err
		}
	} else {
		if err := db.Where("neo_tx_hash = ?", hash).First(&result).Error; err == nil {
			return &result, nil
		} else {
			return nil, err
		}
	}
}

func GetSwapInfosByState(db *gorm.DB, page, pageSize int, state types.SwapState) ([]*types.SwapInfo, error) {
	var result []*types.SwapInfo
	if err := db.Where("state = ?", state).Scopes(Paginate(page, pageSize)).Order("last_modify_time DESC").Find(&result).Error; err == nil {
		return result, nil
	} else {
		return nil, err
	}
}

func GetSwapInfosByAddr(db *gorm.DB, page, pageSize int, addr string, action types.ChainType) ([]*types.SwapInfo, error) {
	var result []*types.SwapInfo
	if action == types.ETH {
		if len(addr) == 40 {
			addr = fmt.Sprintf("0x%s", addr)
		}
		if err := db.Where("eth_user_addr = ?", addr).Scopes(Paginate(page, pageSize)).Order("last_modify_time DESC").Find(&result).Error; err == nil {
			return result, nil
		} else {
			return nil, err
		}
	} else {
		if err := db.Where("neo_user_addr = ?", addr).Scopes(Paginate(page, pageSize)).Order("last_modify_time DESC").Find(&result).Error; err == nil {
			return result, nil
		} else {
			return nil, err
		}
	}
}
