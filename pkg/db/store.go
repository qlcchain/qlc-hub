package db

import (
	"fmt"
	"strings"
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
	db.AutoMigrate(&types.SwapPending{})
	db.AutoMigrate(&types.QGasSwapInfo{})
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
	record.EthUserAddr = stringToLower(record.EthUserAddr)
	return db.Create(record).Error
}

func UpdateSwapInfo(db *gorm.DB, record *types.SwapInfo) error {
	record.LastModifyTime = time.Now().Unix()
	record.EthTxHash = util.AddHashPrefix(record.EthTxHash)
	record.NeoTxHash = util.AddHashPrefix(record.NeoTxHash)
	record.EthUserAddr = stringToLower(record.EthUserAddr)
	return db.Save(record).Error
}

func GetSwapInfos(db *gorm.DB, chain string, page, pageSize int) ([]*types.SwapInfo, error) {
	var result []*types.SwapInfo
	chainType := types.StringToChainType(chain)
	if chain == "" {
		if err := db.Scopes(Paginate(page, pageSize)).Order("last_modify_time DESC").Find(&result).Error; err == nil {
			return result, nil
		} else {
			return nil, err
		}
	} else {
		if chainType == types.ETH {
			if err := db.Where("chain != ?", types.BSC).Scopes(Paginate(page, pageSize)).Order("last_modify_time DESC").Find(&result).Error; err == nil {
				return result, nil
			} else {
				return nil, err
			}
		} else {
			if err := db.Where("chain = ?", types.BSC).Scopes(Paginate(page, pageSize)).Order("last_modify_time DESC").Find(&result).Error; err == nil {
				return result, nil
			} else {
				return nil, err
			}
		}
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

func GetSwapInfosByAddr(db *gorm.DB, page, pageSize int, addr string, chain string, isEthAddr bool) ([]*types.SwapInfo, error) {
	var result []*types.SwapInfo
	chainType := types.StringToChainType(chain)
	if chainType == types.ETH || chainType == types.BSC || chain == "" {
		if isEthAddr {
			addr = stringToLower(addr)
			if len(addr) == 40 {
				addr = fmt.Sprintf("0x%s", addr)
			}
			if chain == "" {
				if err := db.Where("eth_user_addr = ?", addr).Scopes(Paginate(page, pageSize)).Order("last_modify_time DESC").Find(&result).Error; err == nil {
					return result, nil
				} else {
					return nil, err
				}
			} else {
				if chainType == types.BSC {
					if err := db.Where("eth_user_addr = ?", addr).Where("chain = ?", types.BSC).Scopes(Paginate(page, pageSize)).Order("last_modify_time DESC").Find(&result).Error; err == nil {
						return result, nil
					} else {
						return nil, err
					}
				} else {
					if err := db.Where("eth_user_addr = ?", addr).Where("chain != ?", types.BSC).Scopes(Paginate(page, pageSize)).Order("last_modify_time DESC").Find(&result).Error; err == nil {
						return result, nil
					} else {
						return nil, err
					}
				}
			}
		} else {
			if chain == "" {
				if err := db.Where("neo_user_addr = ?", addr).Scopes(Paginate(page, pageSize)).Order("last_modify_time DESC").Find(&result).Error; err == nil {
					return result, nil
				} else {
					return nil, err
				}
			} else {
				if chainType == types.BSC {
					if err := db.Where("neo_user_addr = ?", addr).Where("chain == ?", types.BSC).Scopes(Paginate(page, pageSize)).Order("last_modify_time DESC").Find(&result).Error; err == nil {
						return result, nil
					} else {
						return nil, err
					}
				} else {
					if err := db.Where("neo_user_addr = ?", addr).Where("chain != ?", types.BSC).Scopes(Paginate(page, pageSize)).Order("last_modify_time DESC").Find(&result).Error; err == nil {
						return result, nil
					} else {
						return nil, err
					}
				}
			}
		}
	}
	return nil, nil
}

func stringToLower(str string) string {
	if str == "" {
		return ""
	} else {
		return strings.ToLower(str)
	}
}

func InsertSwapPending(db *gorm.DB, record *types.SwapPending) error {
	record.LastModifyTime = time.Now().Unix()
	record.EthTxHash = util.AddHashPrefix(record.EthTxHash)
	record.NeoTxHash = util.AddHashPrefix(record.NeoTxHash)
	return db.Create(record).Error
}

func GetSwapPendings(db *gorm.DB, page, pageSize int) ([]*types.SwapPending, error) {
	var result []*types.SwapPending
	if err := db.Scopes(Paginate(page, pageSize)).Order("last_modify_time DESC").Find(&result).Error; err == nil {
		return result, nil
	} else {
		return nil, err
	}
}

func GetSwapPendingByTxEthHash(db *gorm.DB, hash string) (*types.SwapPending, error) {
	var result types.SwapPending
	hash = util.AddHashPrefix(hash)
	if err := db.Where("eth_tx_hash = ?", hash).First(&result).Error; err == nil {
		return &result, nil
	} else {
		return nil, err
	}
}

func GetSwapPendingByTxNeoHash(db *gorm.DB, hash string) (*types.SwapPending, error) {
	var result types.SwapPending
	hash = util.AddHashPrefix(hash)
	if err := db.Where("neo_tx_hash = ?", hash).First(&result).Error; err == nil {
		return &result, nil
	} else {
		return nil, err
	}
}

func DeleteSwapPending(db *gorm.DB, record *types.SwapPending) error {
	return db.Delete(record).Error
}

// qgas swap

func InsertQGasSwapInfo(db *gorm.DB, record *types.QGasSwapInfo) error {
	record.LastModifyTime = time.Now().Unix()
	record.CrossChainTxHash = util.AddHashPrefix(record.CrossChainTxHash)
	record.QlcSendTxHash = util.RemoveHexPrefix(record.QlcSendTxHash)
	record.QlcRewardTxHash = util.RemoveHexPrefix(record.QlcRewardTxHash)
	record.CrossChainUserAddr = stringToLower(record.CrossChainUserAddr)
	return db.Create(record).Error
}

func UpdateQGasSwapInfo(db *gorm.DB, record *types.QGasSwapInfo) error {
	record.LastModifyTime = time.Now().Unix()
	record.CrossChainTxHash = util.AddHashPrefix(record.CrossChainTxHash)
	record.CrossChainUserAddr = stringToLower(record.CrossChainUserAddr)
	record.QlcSendTxHash = util.RemoveHexPrefix(record.QlcSendTxHash)
	record.QlcRewardTxHash = util.RemoveHexPrefix(record.QlcRewardTxHash)
	return db.Save(record).Error
}

func GetQGasSwapInfos(db *gorm.DB, chain string, page, pageSize int) ([]*types.QGasSwapInfo, error) {
	var result []*types.QGasSwapInfo
	chainType := types.StringToChainType(chain)
	if chain == "" {
		if err := db.Scopes(Paginate(page, pageSize)).Order("last_modify_time DESC").Find(&result).Error; err == nil {
			return result, nil
		} else {
			return nil, err
		}
	} else {
		if err := db.Where("chain = ?", chainType).Scopes(Paginate(page, pageSize)).Order("last_modify_time DESC").Find(&result).Error; err == nil {
			return result, nil
		} else {
			return nil, err
		}
	}
}

func GetQGasSwapInfoByUserTxHash(db *gorm.DB, hash string) (*types.QGasSwapInfo, error) {
	var result types.QGasSwapInfo
	if err := db.Where("user_tx_hash = ?", hash).First(&result).Error; err == nil {
		return &result, nil
	} else {
		return nil, err
	}
}

func GetQGasSwapInfoByUniqueID(db *gorm.DB, hash string, action types.QGasSwapType) (*types.QGasSwapInfo, error) {
	var result types.QGasSwapInfo
	if action == types.QGasWithdraw {
		hash = util.AddHashPrefix(hash)
		if err := db.Where("cross_chain_tx_hash = ?", hash).First(&result).Error; err == nil {
			return &result, nil
		} else {
			return nil, err
		}
	} else {
		hash = util.RemoveHexPrefix(hash)
		if err := db.Where("qlc_send_tx_hash = ?", hash).First(&result).Error; err == nil {
			return &result, nil
		} else {
			return nil, err
		}
	}
}

func GetQGasSwapInfosByState(db *gorm.DB, page, pageSize int, state types.QGasSwapState) ([]*types.QGasSwapInfo, error) {
	var result []*types.QGasSwapInfo
	if err := db.Where("state = ?", state).Scopes(Paginate(page, pageSize)).Order("last_modify_time DESC").Find(&result).Error; err == nil {
		return result, nil
	} else {
		return nil, err
	}
}

func GetQGasSwapInfosByUserAddr(db *gorm.DB, page, pageSize int, addr string, chain string, isEthUser bool) ([]*types.QGasSwapInfo, error) {
	var result []*types.QGasSwapInfo
	chainType := types.StringToChainType(chain)
	if isEthUser {
		addr = stringToLower(addr)
		if len(addr) == 40 {
			addr = fmt.Sprintf("0x%s", addr)
		}
		if chain == "" {
			if err := db.Where("cross_chain_user_addr = ?", addr).Scopes(Paginate(page, pageSize)).Order("last_modify_time DESC").Find(&result).Error; err == nil {
				return result, nil
			} else {
				return nil, err
			}
		} else {
			if err := db.Where("cross_chain_user_addr = ?", addr).Where("chain = ?", chainType).Scopes(Paginate(page, pageSize)).Order("last_modify_time DESC").Find(&result).Error; err == nil {
				return result, nil
			} else {
				return nil, err
			}
		}
	} else {
		if chain == "" {
			if err := db.Where("qlc_user_addr = ?", addr).Scopes(Paginate(page, pageSize)).Order("last_modify_time DESC").Find(&result).Error; err == nil {
				return result, nil
			} else {
				return nil, err
			}
		} else {
			if err := db.Where("qlc_user_addr = ?", addr).Where("chain = ?", chainType).Scopes(Paginate(page, pageSize)).Order("last_modify_time DESC").Find(&result).Error; err == nil {
				return result, nil
			} else {
				return nil, err
			}
		}
	}
}
