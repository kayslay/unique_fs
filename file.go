package unique_fs

import (
	"io"
	"os"
	"path"
	"time"
)

// FileInterface an interface that describes the methods a file storage implementation should have.
// the signature of the methods depend on the storage platform.
type FileInterface interface {
	//create new file
	Create(path string, file io.Reader) error
	//get file
	Get(path string) (io.ReadCloser, error)
	//delete file
	Delete(path string) error
}

var root = "store"

//exampleFs implements file interface. This is just a simple
// implementation using os.File to show my point.
type ExampleFs struct {
}

func (e ExampleFs) Create(p string, r io.Reader) error {
	//sleeping for 3s to make creating files slow
	time.Sleep(time.Second * 3)

	f, err := os.Create(path.Join(root, p))
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, r)

	return err
}

func (e ExampleFs) Get(p string) (io.ReadCloser, error) {

	return os.Open(path.Join(root, p))
}

func (e ExampleFs) Delete(p string) error {

	return os.Remove(path.Join(root, p))
}
