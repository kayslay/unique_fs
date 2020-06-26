package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/kayslay/unique_fs"
	"io"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	mx sync.RWMutex
)

func main() {

	hh := httpHandler{
		file: unique_fs.ExampleFs{},
		hash: unique_fs.NewExampleHash(),
		db:   unique_fs.NewDB(),
	}

	r := chi.NewRouter()
	r.Get("/file-ref/{id}", hh.get)
	r.Get("/file/{hash}", hh.getFile)
	r.Post("/upload", hh.uploader)
	r.Get("/file_exists/{hash}", hh.hashExists)

	log.Println("server running 2000")

	err := http.ListenAndServe(":2000", r)
	log.Println(err)
}

type uploadStruct struct {
	Body string
	Path string
	Hash string
}

type httpHandler struct {
	file unique_fs.FileInterface
	hash unique_fs.HashStorageLinker
	db   *unique_fs.DB
}

//uploader handles request to upload a file. string value passed to body field represents
// a file.
func (h httpHandler) uploader(w http.ResponseWriter, r *http.Request) {
	mx.Lock()
	defer mx.Unlock()
	var data uploadStruct

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var hash = data.Hash

	//if the hash is passed. check if the hash exists
	if hash != "" {
		//	hash is passed. check if the hash exists
		_, err := h.hash.Get(hash)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

	} else if data.Body != "" {

		hash = hasher(data.Body)
		fileStorePath := hash[:7] + ".txt" // random file name
		_, err := h.hash.Get(hash)
		if err != nil {
			// body was passed so create a new file
			//	create the file
			err := h.file.Create(fileStorePath, strings.NewReader(data.Body))
			if err != nil {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return
			}

			//	create hash store for path
			err = h.hash.Add(hash, fileStorePath)
			if err != nil {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return
			}
		}
	} else {
		http.Error(w, "body or hash field must contain a value", http.StatusBadRequest)
		return
	}

	fr := unique_fs.FileRefObject{
		FileHash:  hash,
		Name:      path.Base(data.Path),
		Type:      "txt",
		CreatedAt: time.Now(),
	}

	err = h.db.Save(&fr)

	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	render.JSON(w, r, fr)

}

func (h httpHandler) get(w http.ResponseWriter, r *http.Request) {
	mx.RLock()
	defer mx.RUnlock()

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ref, err := h.db.Get(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	render.JSON(w, r, ref)
}

func (h httpHandler) getFile(w http.ResponseWriter, r *http.Request) {
	hash := chi.URLParam(r, "hash")
	//	get the file path
	filePath, err := h.hash.Get(hash)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	//	get file on storage
	f, err := h.file.Get(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	defer f.Close()
	w.Header().Set("Content-Type", "text/plain")
	io.Copy(w, f)
}

func (h httpHandler) hashExists(w http.ResponseWriter, r *http.Request) {
	mx.RLock()
	defer mx.RUnlock()

	_, err := h.hash.Get(chi.URLParam(r, "hash"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Write([]byte("ok"))
}

func hasher(val string) string {
	h := md5.New()
	io.WriteString(h, val)
	return fmt.Sprintf("%x", h.Sum(nil))
}
