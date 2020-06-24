package unique_fs

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"
)

var (
	jsonRefFile = "db/db.json"
)

// FileRefObject a DB model that contains the details of the file.
// the details seen by the frontend or user
type FileRefObject struct {
	ID        int       `json:"id"`
	FileHash  string    `json:"file_hash"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

//DB acts as a database store for our example file
type DB struct {
	store   map[int]FileRefObject
	counter int
}

func NewDB() *DB {
	//load from saved json file
	f, err := os.Open(jsonRefFile)
	m := map[int]FileRefObject{}
	if err == nil {
		_ = json.NewDecoder(f).Decode(&m)
	}

	return &DB{store: m, counter: 0}
}

// saveAsJSON persistence
func (d DB) saveAsJSON() {
	f, err := os.Create(jsonRefFile)
	if err != nil {
		log.Println(err)
		return
	}

	err = json.NewEncoder(f).Encode(d.store)
	if err != nil {
		log.Println(err)
	}
}

func (d *DB) Save(f *FileRefObject) error {
	f.ID = d.counter
	//update id counter
	defer func() {
		d.saveAsJSON()
		d.counter++
	}()

	d.store[f.ID] = *f

	return nil
}

func (d *DB) Get(id int) (FileRefObject, error) {

	ref, ok := d.store[id]
	if !ok {
		return FileRefObject{}, errors.New("not found")
	}

	return ref, nil
}

func (d *DB) Delete(id int) error {
	defer d.saveAsJSON()
	delete(d.store, id)
	return nil
}
