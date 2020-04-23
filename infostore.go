package infostore

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"

	"github.com/jakobvarmose/dc/crypto"
)

var (
	ErrorNotFound = errors.New("Not Found")
)

type InfoStore struct {
	dir string

	listeners   []Listener
	listenersLk sync.Mutex
	readers     []*Reader
	readersLk   sync.Mutex
}

func NewInfoStore(dir string) (*InfoStore, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	is := &InfoStore{
		dir: dir,
	}
	return is, nil
}

func (is *InfoStore) Get(page []byte, creator []byte, number int64) (*crypto.Signed, error) {
	pageHex := fmt.Sprintf("%x", page)
	if pageHex == "" {
		pageHex = "a"
	}
	creatorHex := fmt.Sprintf("%x", creator)
	numberDec := fmt.Sprintf("%d", number)

	data, err := ioutil.ReadFile(path.Join(is.dir, pageHex, creatorHex, numberDec))
	if err != nil {
		return nil, err
	}
	return crypto.UnmarshalSigned(data)
}

func (is *InfoStore) Put(signed *crypto.Signed) error {
	info, err := UnmarshalInfo(signed.Data)
	if err != nil {
		return err
	}
	page := info.Page
	creator := signed.Pubkey
	number := info.Number
	data := signed.Marshal()
	pageHex := fmt.Sprintf("%x", page)
	if pageHex == "" {
		pageHex = "a"
	}
	creatorHex := fmt.Sprintf("%x", creator)
	numberDec := fmt.Sprintf("%d", number)

	if err := os.MkdirAll(path.Join(is.dir, pageHex, creatorHex), 0755); err != nil {
		return err
	}
	oldData, err := ioutil.ReadFile(path.Join(is.dir, pageHex, creatorHex, numberDec))
	if err == nil {
		oldSigned, err := crypto.UnmarshalSigned(oldData)
		if err == nil {
			oldInfo, err := UnmarshalInfo(oldSigned.Data)
			if err == nil {
				if oldInfo.Revision > info.Revision {
					return nil
				}
				if oldInfo.Revision == info.Revision && string(oldInfo.Hash) >= string(info.Hash) {
					return nil
				}
			}
		}
	}
	if err := ioutil.WriteFile(path.Join(is.dir, pageHex, creatorHex, numberDec), data, 0644); err != nil {
		return err
	}

	is.listenersLk.Lock()
	defer is.listenersLk.Unlock()
	for _, listener := range is.listeners {
		listener.InfoAdded(page, creator, number)
	}

	return nil
}

func (is *InfoStore) Delete(page []byte, creator []byte, number int64) error {
	pageHex := fmt.Sprintf("%x", page)
	if pageHex == "" {
		pageHex = "a"
	}
	creatorHex := fmt.Sprintf("%x", creator)
	numberDec := fmt.Sprintf("%d", number)

	if err := os.Remove(path.Join(is.dir, pageHex, creatorHex, numberDec)); err != nil {
		return err
	}
	_ = os.Remove(path.Join(is.dir, pageHex, creatorHex))
	_ = os.Remove(path.Join(is.dir, pageHex))

	is.listenersLk.Lock()
	defer is.listenersLk.Unlock()
	for _, listener := range is.listeners {
		listener.InfoDeleted(page, creator, number)
	}

	return nil
}

func (is *InfoStore) Subject(page []byte) *Reader {
	//TODO what to put in the reader
	ch := make(chan *Info)
	go func() {
		pageHex := fmt.Sprintf("%x", page)
		dirs, err := ioutil.ReadDir(path.Join(is.dir, pageHex))
		if err != nil {
			//TODO handle error
		}
		for _, dir := range dirs {
			files, err := ioutil.ReadDir(path.Join(is.dir, pageHex, dir.Name()))
			if err != nil {
				//TODO handle error
			}
			for _, file := range files {
				data, err := ioutil.ReadFile(path.Join(is.dir, pageHex, dir.Name(), file.Name()))
				if err != nil {
					//TODO handle error
				}
				info, err := UnmarshalInfo(data)
				if err != nil {
					//TODO handle error
				}
				ch <- info
			}
		}
		//FIXME the same info may be received twice
	}()
	r := &Reader{
		infos: is,
		Ch:    ch,
	}
	return r
}

type Reader struct {
	infos *InfoStore
	Ch    chan *Info
}

func (is *InfoStore) add(info *Info) {
	is.readersLk.Lock()
	defer is.readersLk.Unlock()
	for _, reader := range is.readers {
		reader.Ch <- info
	}
}

func (r *Reader) Close() {
	go func() {
		ok := true
		for ok {
			_, ok = <-r.Ch
		}
	}()
	r.infos.readersLk.Lock()
	defer r.infos.readersLk.Unlock()
	close(r.Ch)
	for i, reader := range r.infos.readers {
		if r == reader {
			r.infos.readers = append(r.infos.readers[:i], r.infos.readers[i+1:]...)
			break
		}
	}
}

func (r *Reader) Next(ctx context.Context) *Info {
	select {
	case <-ctx.Done():
		return nil
	case info := <-r.Ch:
		return info
	}
}

type Listener interface {
	InfoAdded(page []byte, creator []byte, number int64)
	InfoDeleted(page []byte, creator []byte, number int64)
}

func (is *InfoStore) Listen(l Listener) {
	is.listenersLk.Lock()
	defer is.listenersLk.Unlock()
	is.listeners = append(is.listeners, l)
}

func (is *InfoStore) Unlisten(l Listener) {
	is.listenersLk.Lock()
	defer is.listenersLk.Unlock()
	for i, listener := range is.listeners {
		if l == listener {
			is.listeners[i] = is.listeners[len(is.listeners)-1]
			is.listeners = is.listeners[:len(is.listeners)-1]
			break
		}
	}
}
