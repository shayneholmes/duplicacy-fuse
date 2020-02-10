package dpfs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_key(t *testing.T) {
	tests := []struct {
		snapshotid string
		revision   int
		path       string
		want       []byte
	}{
		{"id", 3, "/file.txt", []byte("id:3:/file.txt")},
	}
	for _, tt := range tests {
		got := key(tt.snapshotid, tt.revision, tt.path)
		assert.Equal(t, tt.want, got)
	}
}

func Test_encodedecode(t *testing.T) {
	tests := []struct {
		input   DpfsKvStoreEntry
		wantErr bool
	}{
		{DpfsKvStoreEntry{
			Size:  100,
			Time:  100,
			Mode:  0555,
			IsDir: false,
		}, false},
	}
	for _, tt := range tests {
		enc, err := encode(tt.input)
		if tt.wantErr {
			assert.NotNil(t, err)
		} else {
			if assert.Nil(t, err) {
				dec, err := decode(enc)
				if assert.Nil(t, err) {
					assert.Equal(t, tt.input, dec)
				}
			}
		}
	}
}
