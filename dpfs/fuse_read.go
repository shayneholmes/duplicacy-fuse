package dpfs

import (
	"encoding/hex"

	duplicacy "github.com/gilbertchen/duplicacy/src"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

// Read satisfies the Read implementation from fuse.FileSystemInterface
func (self *Dpfs) Read(path string, buff []byte, offset int64, fh uint64) (n int) {
	logger := log.WithFields(
		log.Fields{
			"op":     "Read",
			"path":   path,
			"offset": offset,
			"fh":     fh,
			"id":     uuid.NewV4().String(),
		})

	buffLength := int64(len(buff))

	// Check cache

	info := self.newpathInfo(path)

	file, err := self.findFile(info.snapshotid, info.revision, info.filepath)
	if err != nil {
		logger.WithError(err).Debug()
		return 0
	}

	if file.Size == 0 {
		return 0
	}

	manager, err := self.createBackupManager(info.snapshotid)
	if err != nil {
		logger.WithError(err).Debug()
		return 0
	}

	snapshot, err := self.downloadSnapshot(manager, info.snapshotid, info.revision, nil, false)
	if err != nil {
		logger.WithError(err).Debug()
		return 0
	}

	var fileCursor int64 // position of the current chunk within the file

	for i := file.StartChunk; i <= file.EndChunk; i++ {
		chunkHash := snapshot.ChunkHashes[i]
		chunkLength := int64(snapshot.ChunkLengths[i])

		// File boundaries don't necessarily align with chunk boundaries, so the
		// file portion of the chunk may not include the whole chunk, on the first
		// and last chunk of the file.
		var portionStart, portionEnd int64 = 0, chunkLength
		if i == file.StartChunk {
			portionStart = int64(file.StartOffset)
		}
		if i == file.EndChunk {
			portionEnd = int64(file.EndOffset)
		}
		portionSize := portionEnd - portionStart

		if offset > fileCursor+portionSize {
			// This chunk only contains data that precedes the requested offset, so
			// skip it and move to the next chunk.
			fileCursor += portionSize
			continue
		}

		// We want at least some bytes from this chunk. Fetch the chunk and extract the relevant bytes.
		// (This code is similar to RetrieveFile in duplicacy_snapshotmanager.)
		lastChunk, lastChunkHash := self.chunkDownloader.GetLastDownloadedChunk()
		var chunk *duplicacy.Chunk
		if lastChunkHash == chunkHash {
			chunk = lastChunk
		} else {
			logger.
				WithField("chunk", i).
				WithField("chunkHash", hex.EncodeToString([]byte(chunkHash))).
				Debug("downloading new chunk")
			i := self.chunkDownloader.AddChunk(chunkHash)
			chunk = self.chunkDownloader.WaitForChunk(i)
		}

		filePortion := chunk.GetBytes()[portionStart:portionEnd]

		start := offset - fileCursor
		if start < 0 {
			start = 0
		}
		end := start + buffLength - int64(n)
		if end > portionSize {
			end = portionSize
		}
		n += copy(buff[n:], filePortion[start:end])
		logger.
			WithField("chunk", i).
			WithField("len", chunkLength).
			WithField("len(file)", portionSize).
			WithField("n", n).
			WithField("start", start).
			WithField("end", end).
			Debug("copied bytes")

		fileCursor += portionSize

		if int64(n) >= buffLength {
			// Buffer is filled, so we can stop here.
			break
		}
	}

	return
}
