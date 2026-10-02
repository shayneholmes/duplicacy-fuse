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

type EntriesBatch interface {
	PutEntries(key []byte, entry []*duplicacy.Entry) (int, error)
}

type DpfsKvStore interface {
	Close() error
	Delete(key []byte) error
	Get(key []byte) ([]byte, error)
	GetString(key []byte) (string, error)
	GetEntries(key []byte) ([]*duplicacy.Entry, error)
	GetSnapshot(key []byte) (*duplicacy.Snapshot, error)
	Has(key []byte) bool
	Put(key, value []byte) error
	PutString(key []byte, value string) error
	CreateEntryBatch() EntriesBatch
	WriteEntriesBatch(b EntriesBatch) error
	PutSnapshot(key []byte, entry *duplicacy.Snapshot) error
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

// key generates a unique key for this directory
func key(snapshotid string, revision int, filePath string) []byte {
	// Follow the format SNAPSHOT:REV:PATH.
	return []byte(fmt.Sprintf("%s:%d:%s",
		snapshotid,
		revision,
		strings.TrimSuffix(filePath, "/"),
	))
}

func encodeSnapshot(snapshot *duplicacy.Snapshot) (output []byte, err error) {
	var buf bytes.Buffer

	enc := gob.NewEncoder(&buf)
	err = enc.Encode(snapshot)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func decodeSnapshot(input []byte) (snapshot *duplicacy.Snapshot, err error) {
	buf := bytes.NewBuffer(input)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(&snapshot)
	if err != nil {
		return &duplicacy.Snapshot{}, err
	}

	return snapshot, nil
}

func encodeEntries(entries []*duplicacy.Entry) (output []byte, err error) {
	var buf bytes.Buffer

	enc := gob.NewEncoder(&buf)
	err = enc.Encode(entries)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func decodeEntries(input []byte) (entries []*duplicacy.Entry, err error) {
	buf := bytes.NewBuffer(input)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(&entries)
	if err != nil {
		return nil, err
	}

	return entries, nil
}
