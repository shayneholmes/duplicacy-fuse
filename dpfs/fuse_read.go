package dpfs

import (
	duplicacy "github.com/gilbertchen/duplicacy/src"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

// Read satisfies the Read implementation from fuse.FileSystemInterface
func (self *Dpfs) Read(path string, buff []byte, offset int64, fh uint64) (n int) {
	defer func() {
		if r := recover(); r != nil {
			log.WithField("recover", r).Debug("Recovered")
		}
	}()
	logger := log.WithFields(
		log.Fields{
			"op":     "Read",
			"path":   path,
			"buff":   len(buff),
			"offset": offset,
			"fh":     fh,
			"id":     uuid.NewV4().String(),
		})

	buffLength := int64(len(buff))

	// Check cache

	info, err := self.newpathInfo(path)
	if err != nil {
		logger.WithError(err).Debug()
		return 0
	}

	file, err := self.findFile(info.snapshotid, info.revision, info.filepath)
	if err != nil {
		logger.WithError(err).Debug()
		return 0
	}

	if file.Size == 0 {
		return 0
	}

	manager, err := self.createBackupManager()
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

		c := make(chan *duplicacy.Chunk)
		completionFunc := func(chunk *duplicacy.Chunk) {
			c <- chunk
		}
		self.getDataChunkAsync(chunkHash, i, completionFunc, logger)

		// Speculatively prefetch the next few chunks
		for offset := 1; offset <= 3; offset++ {
			if i+offset > file.EndChunk {
				break
			}
			self.getDataChunkAsync(
				snapshot.ChunkHashes[i+offset],
				i+offset,
				func(chunk *duplicacy.Chunk) {},
				logger.
					WithField("prefetch", offset),
			)
		}

		chunk := <-c
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
			WithField("chunkID", chunk.GetID()).
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

func (self *Dpfs) getDataChunkAsync(chunkHash string, i int, f func(chunk *duplicacy.Chunk), logger *log.Entry) {
	logger = logger.
		WithField("chunkID", self.config.GetChunkIDFromHash(chunkHash)).
		WithField("chunkIndex", i)
	if cachedChunk, ok := self.chunkCache.Get(chunkHash); ok {
		go f(cachedChunk)
		return
	}
	self.chunkDownloadsMu.Lock()
	if c, ok := self.activeChunkDownloads[chunkHash]; ok {
		// This chunk is currently being downloaded. Wait for it.
		self.chunkDownloadsMu.Unlock()
		go func() {
			chunk := <-c
			f(chunk)
		}()
		return
	} else {
		// Fetch the chunk
		c := make(chan *duplicacy.Chunk)
		self.activeChunkDownloads[chunkHash] = c
		self.chunkDownloadsMu.Unlock()

		self.chunkOperator.DownloadAsync(
			chunkHash,
			i,     // chunkIndex
			false, // isMetadata
			func(chunk *duplicacy.Chunk, _ int) {
				logger.Debug("fetched new chunk")

				self.chunkCache.Add(chunkHash, chunk)

				self.chunkDownloadsMu.Lock()
				delete(self.activeChunkDownloads, chunkHash)
				self.chunkDownloadsMu.Unlock()

				// Notify everyone who's listening
				for {
					select {
					case c <- chunk:
						continue
					default:
						close(c)
						f(chunk)
						return
					}
				}
			},
		)
	}
}
