/**
 * @Time : 2022/5/11 4:09 下午
 * @Author : frankj
 * @Email : frankxjkuang@gmail.com
 * @Description : --
 * @Revise : --
 */

package store

import (
	"encoding/gob"
	"io"
	"log"
	"os"
	"sync"
)

const recordQueueMaxLength = 100 // chan通道的最大缓存值
var urlStore *URLStore
var once sync.Once

// 映射kv记录
type record struct {
	Key, URL string
}

type URLStore struct {
	urls map[string]string // 短长链的映射关系
	sync.RWMutex
	record chan record // 映射kv记录
}

// NewURLStore 初始化URLStore
func NewURLStore(filename string) *URLStore {
	once.Do(func() {
		urlStore = new(URLStore)
		urlStore.urls = make(map[string]string)
		urlStore.record = make(chan record, recordQueueMaxLength)
	})
	//err := urlStore
	return urlStore
}

// Get 使用短链获取长链
func (s *URLStore) Get(key string) string {
	s.RLock()
	defer s.RUnlock()
	return s.urls[key]
}

// Set 设置短、长链的映射关系
func (s *URLStore) Set(key, url string) bool {
	s.Lock()
	defer s.Unlock()
	_, ok := s.urls[key]
	// 已经存在
	if ok {
		return false
	}
	s.urls[key] = url
	return true
}

// Delete 删除短链
func (s *URLStore) Delete(key string) {
	// TODO：删除会导致重新生成key重复，暂时不用
	s.Lock()
	defer s.Unlock()
	delete(s.urls, key)
}

// Count 获取映射Store的大小
func (s *URLStore) Count() int {
	s.RLock()
	defer s.RUnlock()
	return len(s.urls)
}

// Put 新增短、长链的映射关系
func (s *URLStore) Put(url string) string {
	for {
		key := genKey(s.Count())
		if s.Set(key, url) {
			s.record <- record{key, url}
			return key
		}
	}
	// shouldn’t get here
	panic("put failed")
}

// GetUrls 获取短长链映射
func (s *URLStore) GetUrls() map[string]string {
	return s.urls
}

// 加载文件缓存数据
func (s *URLStore) load(fileName string) error {
	f, err := os.Open(fileName)
	if err != nil {
		log.Printf("load open filename[%s] err[%v]", fileName, err)
		return err
	}
	defer f.Close()
	b := gob.NewDecoder(f)
	for err == nil {
		r := record{}
		err = b.Decode(&r)
		if err != nil {
			s.Set(r.Key, r.URL)
		}
	}
	// 数据读完
	if err == io.EOF {
		return nil
	}
	return err
}

// 异步存储数据
func (s *URLStore) saveLoop(fileName string) {
	f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal("saveLoop OpenFile[%s] err[%v]", fileName, err)
		return
	}
	defer f.Close()
	e := gob.NewEncoder(f)
	for {
		// 从通道里拉数据
		r := s.record
		if err = e.Encode(r); err != nil {
			log.Printf("saveLoop saving to URLStore err: %v \n", err)
		}
	}
}