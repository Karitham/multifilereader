package mfr

import (
	"errors"
	"io"
	"io/fs"
)

// MultiFileReader is a struct that manages a set of entries (file paths) and provides
// a way to read the contents of those entries sequentially.
type MultiFileReader struct {
	filepaths []string
	dir       fs.FS

	// the offset of the current file in the entries reader
	entriesOffset int

	currentReader io.ReadSeekCloser

	// the total offset of the entries reader
	// sum of all file sizes passed + the current file offset
	totalOffset int64

	// the offset of the current file
	fileOffset int64

	// the size of all files combined
	totalSize int64

	// size of each file
	filesizes []int64
}

// New creates a new MultiFileReader from the given directory and entries.
// fs.FS Opened files must be seekable.
func New(dir fs.FS, files []string) *MultiFileReader {
	r := &MultiFileReader{
		filepaths: files,
		dir:       dir,
	}

	for _, entry := range files {
		info, err := fs.Stat(dir, entry)
		if err != nil {
			panic(err)
		}
		r.filesizes = append(r.filesizes, info.Size())
		r.totalSize += info.Size()
	}

	return r
}

func (mfr *MultiFileReader) Read(p []byte) (int, error) {
	var read int
	for {
		if mfr.currentReader == nil {
			err := mfr.openCurrentFile()
			if err != nil {
				return read, err
			}
		}

		if mfr.fileOffset > 0 {
			n, err := mfr.currentReader.Seek(mfr.fileOffset, io.SeekStart)
			if err != nil {
				return read, err
			}

			if n != mfr.fileOffset {
				return read, errors.New("seek failed")
			}
		}

		n, err := mfr.currentReader.Read(p[read:])
		read += n
		mfr.fileOffset += int64(n)
		mfr.totalOffset += int64(n)
		if err != nil {
			if errors.Is(err, io.EOF) && mfr.entriesOffset >= len(mfr.filepaths) {
				return read, io.EOF
			}

			// we aren't the last file
			if errors.Is(err, io.EOF) {
				mfr.fileOffset = 0
				mfr.currentReader.Close()
				mfr.currentReader = nil
				mfr.entriesOffset++

				continue
			}

			return read, err
		}

		if read == len(p) {
			return read, nil
		}
	}
}

func (mfr *MultiFileReader) Seek(offset int64, whence int) (int64, error) {
	var abs int64
	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = mfr.totalOffset + offset
	case io.SeekEnd:
		abs = mfr.totalSize + offset
	default:
		return 0, errors.New("invalid whence")
	}

	if abs < 0 {
		return 0, errors.New("negative position")
	}

	if mfr.currentReader != nil {
		mfr.currentReader.Close()
		mfr.currentReader = nil
	}

	if abs == mfr.totalSize {
		mfr.totalOffset = mfr.totalSize
		return mfr.totalSize, nil
	}

	if abs > mfr.totalSize {
		mfr.totalOffset = mfr.totalSize
		return 0, io.EOF
	}

	mfr.entriesOffset = 0
	var fileStartOffset int64
	for mfr.entriesOffset < len(mfr.filepaths) {
		fileSize := mfr.filesizes[mfr.entriesOffset]
		if fileStartOffset+fileSize > abs {
			break
		}
		fileStartOffset += fileSize
		mfr.entriesOffset++
	}

	mfr.totalOffset = abs
	mfr.fileOffset = abs - fileStartOffset
	return abs, nil
}

func (mfr *MultiFileReader) Close() error {
	if mfr.currentReader != nil {
		return mfr.currentReader.Close()
	}
	return nil
}

func (mfr *MultiFileReader) openCurrentFile() error {
	if mfr.entriesOffset >= len(mfr.filepaths) {
		return io.EOF
	}

	f, err := mfr.dir.Open(mfr.filepaths[mfr.entriesOffset])
	if err != nil {
		return err
	}

	if mfr.currentReader != nil {
		mfr.currentReader.Close()
	}

	mfr.currentReader = f.(io.ReadSeekCloser)
	return nil
}
