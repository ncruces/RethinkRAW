package main

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"rethinkraw/internal/util"
	"rethinkraw/osutil"
)

var batches Batches

type Batches struct {
	lock  sync.Mutex
	queue [16]struct {
		id    string
		paths []string
	}
}

func (p *Batches) New(paths []string) string {
	p.lock.Lock()
	defer p.lock.Unlock()

	for i := len(p.queue) - 1; i > 0; i-- {
		p.queue[i] = p.queue[i-1]
	}

	id := util.RandomID()
	p.queue[0].id = id
	p.queue[0].paths = paths
	return id
}

func (p *Batches) Get(id string) []string {
	p.lock.Lock()
	defer p.lock.Unlock()

	for j, t := range p.queue {
		if t.id == id {
			for i := j; i > 0; i-- {
				p.queue[i] = p.queue[i-1]
			}
			p.queue[0] = t
			return t.paths
		}
	}

	return nil
}

type BatchPhoto struct {
	Path string
	Name string
}

func FindPhotos(batch []string) ([]BatchPhoto, error) {
	var photos []BatchPhoto
	for _, path := range batch {
		var prefix string
		if len(batch) > 1 {
			prefix, _ = filepath.Split(path)
		} else {
			prefix = path + string(filepath.Separator)
		}
		err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if osutil.HiddenFile(info) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if info.Mode().IsRegular() {
				if _, ok := extensions[strings.ToUpper(filepath.Ext(path))]; ok {
					var name string
					if strings.HasPrefix(path, prefix) {
						name = path[len(prefix):]
					} else {
						_, name = filepath.Split(path)
					}
					photos = append(photos, BatchPhoto{path, name})
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return photos, nil
}

func BatchProcess(photos []BatchPhoto, proc func(photo BatchPhoto) error) <-chan error {
	const parallelism = 2

	output := make(chan error, parallelism)
	input := make(chan BatchPhoto)
	wait := sync.WaitGroup{}
	wait.Add(parallelism)

	for n := 0; n < parallelism; n++ {
		go func() {
			for photo := range input {
				output <- proc(photo)
			}
			wait.Done()
		}()
	}

	go func() {
		for _, photo := range photos {
			input <- photo
		}
		close(input)
		wait.Wait()
		close(output)
	}()

	return output
}
