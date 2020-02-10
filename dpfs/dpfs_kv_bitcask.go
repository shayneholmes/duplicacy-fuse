package dpfs

import (
	duplicacy "github.com/gilbertchen/duplicacy/src"
	"github.com/prologic/bitcask"
	log "github.com/sirupsen/logrus"
)

type bitcaskKv struct {
	db *bitcask.Bitcask
}

func NewBitcaskKv(path string) (kv *bitcaskKv, err error) {
	db, err := bitcask.Open(path, []bitcask.Option{
		bitcask.WithMaxKeySize(1024),
	}...)
	if err != nil {
		log.WithError(err).Debug()
		return nil, err
	}

	return &bitcaskKv{db: db}, nil
}

func (kv *bitcaskKv) Close() error {
	return kv.db.Close()
}

func (kv *bitcaskKv) Delete(key []byte) error {
	return kv.db.Delete(key)
}

func (kv *bitcaskKv) Has(key []byte) bool {
	return kv.db.Has(key)
}

func (kv *bitcaskKv) Get(key []byte) (*duplicacy.Entry, error) {
	value, err := kv.db.Get(key)
	if err != nil {
		return &duplicacy.Entry{}, err
	}

	return decode(value)
}

func (kv *bitcaskKv) Put(key []byte, entry *duplicacy.Entry) error {
	value, err := encode(entry)
	if err != nil {
		return err
	}
	return kv.db.Put(key, value)
}

func (kv *bitcaskKv) Scan(prefix []byte, f func(key []byte) error) error {
	return kv.db.Scan(prefix, f)
}
