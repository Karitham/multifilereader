package mfr

import (
	"bytes"
	"crypto/sha256"
	"io"
	"os"
	"path"
	"path/filepath"
	"slices"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntriesReader_Read(t *testing.T) {
	d, e, err := writeTestDir(t, map[string]string{
		"file1.txt": "file1 content",
		"file2.txt": "file2 content",
		"file3.txt": "file3 content",
	})
	require.NoError(t, err)

	// Create an entriesReader with the test files
	r := New(os.DirFS(d), e)

	// Test reading from the entriesReader
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)

	// Assert the content read from the entriesReader
	expected := "file1 contentfile2 contentfile3 content"
	assert.Equal(t, expected, buf.String())

	// Test reading from an empty entriesReader
	r = New(os.DirFS(d), []string{})
	n, err := r.Read(make([]byte, 10))
	assert.Equal(t, 0, n)
	assert.Equal(t, io.EOF, err)
}

func TestEntriesReaderReadOne(t *testing.T) {
	d, e, err := writeTestDir(t, map[string]string{
		"file1.txt": "file1 content",
	})
	require.NoError(t, err)

	r := New(os.DirFS(d), e)

	// Read from the entriesReader
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)

	// Assert the content read from the entriesReader matches the file content
	assert.Equal(t, "file1 content", buf.String())

	// Read the file content directly
	fileContent, err := os.ReadFile(filepath.Join(d, "file1.txt"))
	require.NoError(t, err)

	// Assert the checksums match
	assert.Equal(t, sha256.Sum256([]byte("file1 content")), sha256.Sum256(fileContent))
}

func TestEntriesReader_Seek(t *testing.T) {
	d, e, err := writeTestDir(t, map[string]string{
		"file1.txt": "file1 content",
		"file2.txt": "file2 content",
	})
	require.NoError(t, err)

	r := New(os.DirFS(d), e)

	offset, err := r.Seek(5, io.SeekStart)
	require.NoError(t, err)
	assert.Equal(t, int64(5), offset)

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	assert.Equal(t, " contentfile2 content", buf.String())

	_, err = r.Seek(0, io.SeekCurrent)
	require.NoError(t, err)

	_, err = r.Seek(0, io.SeekEnd)
	require.NoError(t, err)

	_, err = r.Seek(0, 3)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid whence")
}

func TestEntriesReader_SeekMultipleFiles(t *testing.T) {
	d, e, err := writeTestDir(t, map[string]string{
		"file1.txt": "file1 content uwu",
		"file2.txt": "file2 content",
	})
	require.NoError(t, err)

	r := New(os.DirFS(d), e)

	offset, err := r.Seek(5, io.SeekStart)
	require.NoError(t, err)
	assert.Equal(t, int64(5), offset)

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	assert.Equal(t, " content uwufile2 content", buf.String())

	offset, err = r.Seek(0, io.SeekStart)
	require.NoError(t, err)
	assert.Equal(t, int64(0), offset)

	buf.Reset()
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	assert.Equal(t, "file1 content uwufile2 content", buf.String())
}

func TestEntriesReader_SeekInvalidOffset(t *testing.T) {
	d, e, err := writeTestDir(t, map[string]string{
		"file1.txt": "file1 content",
	})
	require.NoError(t, err)

	r := New(os.DirFS(d), e)
	offset, err := r.Seek(-1, io.SeekStart)
	assert.Equal(t, int64(0), offset)
	assert.Error(t, err)
}

func TestEntriesReader(t *testing.T) {
	d, e, err := writeTestDir(t, map[string]string{
		"file1.txt": "file1 content ",
		"file2.txt": "file2 content",
	})
	require.NoError(t, err)

	r := New(os.DirFS(d), e)

	err = iotest.TestReader(r, []byte("file1 content file2 content"))
	require.NoError(t, err)
}

func TestEntriesReader_ReadAtFilesBoundary(t *testing.T) {
	d, e, err := writeTestDir(t, map[string]string{
		"file1.txt": "file1 content",
		"file2.txt": "file2 content",
	})
	require.NoError(t, err)

	r := New(os.DirFS(d), e)

	r.Seek(11, io.SeekStart)

	// Read at the boundary of file1 and file2
	buf := make([]byte, 5)
	n, err := r.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, "ntfil", string(buf))
}

func TestEntriesReader_SeekAndReadAtFilesBoundary(t *testing.T) {
	d, e, err := writeTestDir(t, map[string]string{
		"file1.txt": "file1 content",
		"file2.txt": "file2 content",
	})
	require.NoError(t, err)

	r := New(os.DirFS(d), e)

	// Seek to the end of file1
	offset, err := r.Seek(13, io.SeekStart)
	require.NoError(t, err)
	assert.Equal(t, int64(13), offset)

	// Read from the boundary of file1 and file2
	buf := make([]byte, 5)
	n, err := r.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, "file2", string(buf))
}

func TestEntriesReader_SeekAndReadAtOffset(t *testing.T) {
	d, e, err := writeTestDir(t, map[string]string{
		"file1.txt": "file1 content",
		"file2.txt": "file2 content",
	})
	require.NoError(t, err)

	r := New(os.DirFS(d), e)

	// Seek to offset 18
	offset, err := r.Seek(18, io.SeekStart)
	require.NoError(t, err)
	assert.Equal(t, int64(18), offset)

	// Read from offset 18
	buf := make([]byte, 8)
	n, err := r.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, 8, n)
	assert.Equal(t, " content", string(buf))
}

func writeTestDir(t *testing.T, files map[string]string) (string, []string, error) {
	dir := t.TempDir()
	t.Helper()

	written := make([]string, 0, len(files))
	for name, content := range files {
		filePath := path.Join(dir, name)
		if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
			return "", nil, err
		}

		written = append(written, name)
	}

	// sort
	slices.Sort(written)

	return dir, written, nil
}
