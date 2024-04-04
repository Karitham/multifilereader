# MultiFileReader

MultiFileReader is a Go package that provides a way to read the contents of multiple files sequentially as if they were a single file. It manages a set of file paths and allows seeking to specific positions within the combined file.

Thanks to @zhayes for helping me debug this.

## Features

Reads multiple files sequentially as a single file
Supports seeking to specific positions within the combined file
Implements io.Reader, io.Seeker, and io.Closer interfaces
Compatible with fs.FS file systems
Installation
To install the package, use the following command:

```sh
go get github.com/your/package/path/mfr
```

## Usage

Here's an example of how to use the MultiFileReader:

```go
package main

import (
    "fmt"
    "io"
    "os"

    "github.com/your/package/path/mfr"
)

func main() {
    // Create a new MultiFileReader
    dir := os.DirFS("path/to/directory")
    files := []string{"file1.txt", "file2.txt", "file3.txt"}
    reader := mfr.NewMFR(dir, files)

    // Read the contents of the files
    buffer := make([]byte, 1024)
    for {
        n, err := reader.Read(buffer)
        if err == io.EOF {
            break
        }
        if err != nil {
            fmt.Println("Error:", err)
            return
        }
        fmt.Print(string(buffer[:n]))
    }

    // Seek to a specific position
    offset := int64(100)
    newOffset, err := reader.Seek(offset, io.SeekStart)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Println("New offset:", newOffset)

    // Close the reader
    err = reader.Close()
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
}
```

## API

```go
NewMFR(dir fs.FS, files []string) *MultiFileReader
```

Creates a new MultiFileReader from the given directory and file paths. The dir parameter should be an fs.FS file system, and the files parameter should be a slice of file paths relative to the directory.

The order of files passed in signify the order in which the files will be read.

The MFR implements the io.Reader, io.Seeker, and io.Closer interfaces.
