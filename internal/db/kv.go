package db

import (
	"github.com/dgraph-io/badger/v3"
	log "github.com/sirupsen/logrus"
)

func fileInit() *badger.DB {
	opt := badger.DefaultOptions("/tmp/kv") // Store the data in /tmp/kv
	kv, err := badger.Open(opt)
	if err != nil {
		log.Fatal(err)
	}
	defer kv.Close()
	return kv
}

func memInit() *badger.DB {
	opt := badger.DefaultOptions("").WithInMemory(true)
	kv, err := badger.Open(opt)
	if err != nil {
		log.Fatal(err)
	}
	defer kv.Close()
	return kv
}

func setTokenKV(key string, token string) bool {
	kv := memInit()
	err := kv.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(key), []byte(token))
		return err
	})
	if err != nil {
		log.Fatal(err)
		return false
	}
	return true
}

func getTokenKV(key string) string {
	kv := memInit()
	var valCopy []byte
	err := kv.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		valCopy, err = item.ValueCopy(nil)
		return err
	})
	if err != nil {
		log.Fatal(err)
		return ""
	}
	return string(valCopy)
}
