package files

import (
    "io"
    "io/fs"
)

var sourceFileSystem FileSystem
type FileSystem interface {
    fs.ReadFileFS
    fs.ReadDirFS
}
func GetFileSystem() FileSystem {
    return sourceFileSystem
}

// SetFileSystem overrides where assets are read from. The build-tagged init
// functions pick a directory relative to the working directory, which is the
// package directory when running tests. Tests point this at the real asset
// tree instead.
func SetFileSystem(fileSystem FileSystem) {
    sourceFileSystem = fileSystem
}
func Open(filename string) (io.ReadCloser, error) {
    f, err := sourceFileSystem.Open(filename)
    readCloser, ok := f.(io.ReadCloser)
    if !ok {
        return nil, fs.ErrInvalid
    }
    return readCloser, err
}

func ReadFile(filename string) ([]byte, error) {
    return sourceFileSystem.ReadFile(filename)
}
