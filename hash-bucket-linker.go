package unique_fs

import (
	"encoding/json"
	"errors"
	"log"
	"os"
)

var (
	jsonHashFile = "db/hash-storage.json"
)

// HashStorageLinker describes the interface for managing hash storage linkers
type HashStorageLinker interface {
	//add key and value set to DB
	Add(key, value string) error
	// get the value of a key
	Get(key string) (string, error)
	//removes a key
	Remove(key string) error
}

// this is meant to be implemented using a KV-Store, but I'm making use of maps.
type exampleHash struct {
	store map[string]string
}

func (e exampleHash) saveHashAsJSON() {
	f, err := os.Create(jsonHashFile)
	if err != nil {
		log.Println(err)
		return
	}

	err = json.NewEncoder(f).Encode(e.store)
	if err != nil {
		log.Println(err)
	}
}

func NewExampleHash() *exampleHash {
	//load from saved json file
	f, err := os.Open(jsonHashFile)
	m := map[string]string{}
	if err == nil {
		_ = json.NewDecoder(f).Decode(&m)
	}

	return &exampleHash{store: m}
}

func (e exampleHash) Add(key, value string) error {
	defer e.saveHashAsJSON()
	e.store[key] = value
	return nil
}

func (e exampleHash) Get(key string) (string, error) {
	v, ok := e.store[key]
	if !ok {
		return "", errors.New("not found")
	}

	return v, nil
}

func (e exampleHash) Remove(key string) error {
	defer e.saveHashAsJSON()

	delete(e.store, key)
	return nil
}
