package dpfs

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"strings"

	duplicacy "github.com/gilbertchen/duplicacy/src"
	log "github.com/sirupsen/logrus"
)

type DpfsKvStore interface {
	Close() error
	Delete(key []byte) error
	Get(key []byte) ([]byte, error)
	GetString(key []byte) (string, error)
	GetEntry(key []byte) (*duplicacy.Entry, error)
	Has(key []byte) bool
	Put(key, value []byte) error
	PutString(key []byte, value string) error
	PutEntry(key []byte, entry *duplicacy.Entry) error
	Scan(prefix []byte, f func(key []byte) error) error
}

func NewDpfsKv(url string) (kv DpfsKvStore, err error) {
	scheme := strings.Split(url, "://")[0]
	path := url[len(scheme)+3:]
	log.WithField("scheme", scheme).WithField("path", path).Debug()
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("error creating cache dir: %w", err)
	}
	if scheme == "bitcask" {
		return NewBitcaskKv(path)
	}

	return nil, fmt.Errorf("unsupported kv store")
}

func key(snapshotid string, revision int, path string) []byte {
	return []byte(fmt.Sprintf("%s:%d:%s", snapshotid, revision, strings.TrimSuffix(path, "/")))
}

func encode(entry *duplicacy.Entry) (output []byte, err error) {
	var buf bytes.Buffer

	enc := gob.NewEncoder(&buf)
	err = enc.Encode(entry)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func decode(input []byte) (entry *duplicacy.Entry, err error) {
	buf := bytes.NewBuffer(input)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(&entry)
	if err != nil {
		return &duplicacy.Entry{}, err
	}

	return entry, nil
}
