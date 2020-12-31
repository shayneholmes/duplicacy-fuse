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

	// TODO: Cache snapshot between reads
	snapshot, err := self.downloadSnapshot(manager, info.snapshotid, info.revision, nil, false)
	if err != nil {
		logger.WithError(err).Debug()
		return 0
	}

	skippedBytes := int64(0)

	for i := file.StartChunk; i <= file.EndChunk; i++ {
		chunkLength := int64(snapshot.ChunkLengths[i])
		chunkStart := 0
		if i == file.StartChunk {
			chunkStart = file.StartOffset
		}
		chunkEnd := snapshot.ChunkLengths[i]
		if i == file.EndChunk {
			chunkEnd = file.EndOffset
		}
		fileInChunkLength := int64(chunkEnd - chunkStart)

		if offset > skippedBytes+fileInChunkLength {
			// This chunk has no data we're interested in: Skip it.
			logger.
				WithField("chunk", i).
				WithField("len", chunkLength).
				WithField("len(file)", fileInChunkLength).
				WithField("skippedBytes", skippedBytes).
				Debug("skipping chunk")
			skippedBytes += fileInChunkLength
			continue
		}

		// We want at least some bytes from this chunk. Fetch the bytes that belong to this file.
		// (This code is similar to RetrieveFile in duplicacy_snapshotmanager.)
		hash := snapshot.ChunkHashes[i]
		lastChunk, lastChunkHash := self.chunkDownloader.GetLastDownloadedChunk()
		var chunk *duplicacy.Chunk
		if lastChunkHash != hash {
			logger.
				WithField("chunk", i).
				WithField("hash", hex.EncodeToString([]byte(hash))).
				Debug("downloading new chunk")
			i := self.chunkDownloader.AddChunk(hash)
			chunk = self.chunkDownloader.WaitForChunk(i)
		} else {
			chunk = lastChunk
		}

		fileChunk := chunk.GetBytes()[chunkStart:chunkEnd]

		// We want at least some bytes from this chunk.
		start := offset - int64(skippedBytes)
		if start < 0 {
			start = 0
		}
		end := start + buffLength - int64(n)
		if end > fileInChunkLength {
			end = fileInChunkLength
		}
		n += copy(buff[n:], fileChunk[start:end])
		logger.
			WithField("chunk", i).
			WithField("len", chunkLength).
			WithField("len(file)", fileInChunkLength).
			WithField("n", n).
			WithField("start", start).
			WithField("end", end).
			Debug("copied bytes")

		if int64(n) >= buffLength {
			logger.Debug("filled the buffer")
			break
		}

		skippedBytes += fileInChunkLength
	}

	return
}
