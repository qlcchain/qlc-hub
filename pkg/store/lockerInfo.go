package store

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/qlcchain/qlc-hub/pkg/db"
	"github.com/qlcchain/qlc-hub/pkg/types"
)

func getLockerInfoKey(hash string) ([]byte, error) {
	key := make([]byte, 0)
	hByte, err := hex.DecodeString(hash)
	if err != nil {
		return nil, fmt.Errorf("key decode: %s", err)
	}
	key = append(key, byte(KeyPrefixLockerInfo))
	key = append(key, hByte...)
	return key, nil
}

func (s *Store) AddLockerInfo(info *types.LockerInfo) error {
	k, err := getLockerInfoKey(info.RHash)
	if err != nil {
		s.logger.Errorf("getLockerInfoKey: %s [%s]", err, info.RHash)
		return err
	}

	info.StartTime = time.Now().Unix()
	info.LastModifyTime = time.Now().Unix()

	v, err := info.Serialize()
	if err != nil {
		s.logger.Errorf("Serialize: %s [%s]", err, info.RHash)
		return err
	}
	_, err = s.store.Get(k)
	if err == nil {
		s.logger.Errorf("LockerInfoExists [%s]", info.RHash)
		return ErrLockerInfoExists
	} else if err != db.KeyNotFound {
		s.logger.Errorf("store get: %s, [%s]", err, info.RHash)
		return err
	}
	err = s.store.Put(k, v)
	if err != nil {
		s.logger.Errorf("store Put: %s, [%s]", err, info.RHash)
		return err
	}
	return nil
}

func (s *Store) GetLockerInfo(hash string) (*types.LockerInfo, error) {
	k, err := getLockerInfoKey(hash)
	if err != nil {
		return nil, err
	}

	info := new(types.LockerInfo)
	val, err := s.store.Get(k)
	if err != nil {
		if err == db.KeyNotFound {
			return nil, ErrLockerInfoNotFound
		}
		return nil, err
	}

	if err := info.Deserialize(val); err != nil {
		return nil, err
	}
	return info, nil
}

func (s *Store) GetLockerInfos(fn func(info *types.LockerInfo) error) error {
	prefix := []byte{byte(KeyPrefixLockerInfo)}
	err := s.store.Iterator(prefix, nil, func(key []byte, val []byte) error {
		info := new(types.LockerInfo)
		if err := info.Deserialize(val); err != nil {
			return fmt.Errorf("deserialize info error: %s", err)
		}
		if err := fn(info); err != nil {
			return fmt.Errorf("process Info error: %s", err)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) CountLockerInfos() (uint64, error) {
	prefix := []byte{byte(KeyPrefixLockerInfo)}
	return s.store.Count(prefix)
}

func (l *Store) UpdateLockerInfo(info *types.LockerInfo) error {
	k, err := getLockerInfoKey(info.RHash)
	if err != nil {
		l.logger.Errorf("getLockerInfoKey: %s  [%s]", err, info.RHash)
		return err
	}

	info.LastModifyTime = time.Now().Unix()
	v, err := info.Serialize()
	if err != nil {
		l.logger.Errorf("info Serialize: %s  [%s]", err, info.RHash)
		return err
	}

	_, err = l.store.Get(k)
	if err != nil {
		l.logger.Errorf("info get: %s  [%s]", err, info.RHash)
		if err == db.KeyNotFound {
			return ErrLockerInfoNotFound
		}
		return err
	}
	err = l.store.Put(k, v)
	if err != nil {
		l.logger.Errorf("store Put: %s [%s]", err, info.RHash)
		return err
	}
	return nil
}
