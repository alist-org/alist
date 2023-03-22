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

func setFileKV(key string, value string) bool {
	kv := fileInit()
	err := kv.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(key), []byte(value))
		return err
	})
	if err != nil {
		log.Fatal(err)
		return false
	}
	return true
}

func getFileKV(key string) string {
	kv := fileInit()
	var valCopy []byte
	err := kv.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			valCopy = append([]byte{}, val...)
			return nil
		})
		return err
	})
	if err != nil {
		log.Fatal(err)
		return ""
	}
	return string(valCopy)
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

func setMemKV(key string, value string) bool {
	kv := memInit()
	err := kv.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(key), []byte(value))
		return err
	})
	if err != nil {
		log.Fatal(err)
		return false
	}
	return true
}

func getMemKV(key string) string {
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
