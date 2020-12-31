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

	// fileCursor tracks the position of chunk[i] within the file.
	var fileCursor int64

	for i := file.StartChunk; i <= file.EndChunk; i++ {
		chunkHash := snapshot.ChunkHashes[i]
		chunkLength := int64(snapshot.ChunkLengths[i])

		// File boundaries don't necessarily align with chunk boundaries, so the
		// first and last chunk might include bytes that aren't part of the file.
		// We call the relevant portion the "file chunk".
		var fileChunkStart, fileChunkEnd int64 = 0, chunkLength
		if i == file.StartChunk {
			fileChunkStart = int64(file.StartOffset)
		}
		if i == file.EndChunk {
			fileChunkEnd = int64(file.EndOffset)
		}
		fileChunkLength := fileChunkEnd - fileChunkStart

		if fileCursor+fileChunkLength <= offset {
			// All this chunk's data precedes the requested offset, so skip it and
			// move to the next chunk.
			fileCursor += fileChunkLength
			continue
		}

		// We should provide some bytes from this chunk. Fetch the chunk and
		// extract the relevant bytes.
		// (This code is similar to duplicacy's RetrieveFile.)
		lastChunk, lastChunkHash := self.chunkDownloader.GetLastDownloadedChunk()
		var chunk *duplicacy.Chunk
		if lastChunkHash == chunkHash {
			chunk = lastChunk
		} else {
			logger.
				WithField("chunk", i).
				WithField("chunkHash", hex.EncodeToString([]byte(chunkHash))).
				Debug("downloading new chunk")
			chunkIndex := self.chunkDownloader.AddChunk(chunkHash)
			chunk = self.chunkDownloader.WaitForChunk(chunkIndex)
		}

		fileChunk := chunk.GetBytes()[fileChunkStart:fileChunkEnd]

		start := offset - fileCursor
		if start < 0 {
			start = 0
		}
		end := start + buffLength - int64(n)
		if end > fileChunkLength {
			end = fileChunkLength
		}
		n += copy(buff[n:], fileChunk[start:end])
		logger.
			WithField("chunk", i).
			WithField("n", n).
			WithField("start", start).
			WithField("end", end).
			Debug("copied bytes")

		fileCursor += fileChunkLength

		if int64(n) >= buffLength {
			// Buffer is full, so we can stop here.
			break
		}
	}

	return
}
