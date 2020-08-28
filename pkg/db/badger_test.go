package db

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/qlcchain/qlc-hub/pkg/util"
)

func setupTestCase(t *testing.T) (func(t *testing.T), Store) {
	wd, _ := os.Getwd()
	dir := filepath.Join(wd, "store", util.RandomHexString(10))
	_ = util.CreateDirIfNotExist(dir)
	db, err := NewBadgerStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	return func(t *testing.T) {
		//err := l.DBStore.Erase()
		err := db.Close()
		if err != nil {
			t.Fatal(err)
		}
		//CloseLedger()
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
	}, db
}

func TestNewBadgerStore(t *testing.T) {
	teardownTestCase, db := setupTestCase(t)
	defer teardownTestCase(t)

	if err := db.Put([]byte{1, 2, 3}, []byte{4, 5, 6}); err != nil {
		t.Fatal(err)
	}
	r, err := db.Get([]byte{1, 2, 3})
	if err != nil {
		t.Fatal(err)
	}
	prefix := []byte{1}
	if err := db.Drop(prefix); err != nil {
		t.Fatal(err)
	}
	if i, err := db.Count(prefix); i != 0 || err != nil {
		t.Fatal(i, err)
	}
	t.Log(r)
}
