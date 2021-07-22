package dpfs

import (
	duplicacy "github.com/gilbertchen/duplicacy/src"
	"git.mills.io/prologic/bitcask"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
)

type bitcaskKv struct {
	db *bitcask.Bitcask
}

func NewBitcaskKv(path string) (kv *bitcaskKv, err error) {
	db, err := bitcask.Open(path, []bitcask.Option{
		bitcask.WithMaxKeySize(4096),
		bitcask.WithMaxValueSize(8388608),
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

func (kv *bitcaskKv) Get(key []byte) ([]byte, error) {
	return kv.db.Get(key)
}

func (kv *bitcaskKv) GetString(key []byte) (string, error) {
	value, err := kv.db.Get(key)
	if err != nil {
		return "", err
	}

	return cast.ToStringE(value)
}

func (kv *bitcaskKv) GetEntry(key []byte) (*duplicacy.Entry, error) {
	value, err := kv.db.Get(key)
	if err != nil {
		return &duplicacy.Entry{}, err
	}

	return decodeEntry(value)
}

func (kv *bitcaskKv) GetSnapshot(key []byte) (*duplicacy.Snapshot, error) {
	value, err := kv.db.Get(key)
	if err != nil {
		return &duplicacy.Snapshot{}, err
	}

	return decodeSnapshot(value)
}

func (kv *bitcaskKv) Put(key, value []byte) error {
	return kv.db.Put(key, value)
}

func (kv *bitcaskKv) PutString(key []byte, value string) error {
	return kv.db.Put(key, []byte(value))
}

func (kv *bitcaskKv) PutEntry(key []byte, entry *duplicacy.Entry) error {
	value, err := encodeEntry(entry)
	if err != nil {
		return err
	}
	return kv.db.Put(key, value)
}

func (kv *bitcaskKv) PutSnapshot(key []byte, snapshot *duplicacy.Snapshot) error {
	value, err := encodeSnapshot(snapshot)
	if err != nil {
		return err
	}
	return kv.db.Put(key, value)
}

func (kv *bitcaskKv) Scan(prefix []byte, f func(key []byte) error) error {
	return kv.db.Scan(prefix, f)
}
