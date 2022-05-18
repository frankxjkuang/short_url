/**
 * @Time : 2022/5/11 4:09 下午
 * @Author : frankj
 * @Email : frankxjkuang@gmail.com
 * @Description : --
 * @Revise : --
 */

package store

import (
	"encoding/json"
	"errors"
	"fmt"
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
func NewURLStore(fileName string) *URLStore {
	once.Do(func() {
		urlStore = new(URLStore)
		urlStore.urls = make(map[string]string)
		urlStore.record = make(chan record, recordQueueMaxLength)
		if err := urlStore.load(fileName); err != nil {
			panic(fmt.Sprintf("NewURLStore fileName[%s] failed err: [%v]", fileName, err))
		}
		go urlStore.saveLoop(fileName)
	})
	return urlStore
}

// Get 使用短链获取长链
func (s *URLStore) Get(key, url *string) error {
	s.RLock()
	defer s.RUnlock()
	if u, ok := s.urls[*key]; ok {
		*url = u
		return nil
	}
	return errors.New("Key does not exist")
}

// Set 设置短、长链的映射关系
func (s *URLStore) Set(key, url *string) error {
	s.Lock()
	defer s.Unlock()
	_, ok := s.urls[*key]
	// 已经存在
	if ok {
		return errors.New("Key already exists")
	}
	s.urls[*key] = *url
	return nil
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
func (s *URLStore) Put(url, key *string) error {
	for {
		*key = genKey(s.Count())
		if err := s.Set(key, url); err == nil {
			s.record <- record{*key, *url}
			break
		}
	}
	return nil
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
	b := json.NewDecoder(f)
	for err == nil {
		r := record{}
		if err = b.Decode(&r); err == nil {
			s.Set(&r.Key, &r.URL)
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
		log.Fatalf("saveLoop OpenFile[%s] err[%v]", fileName, err)
		return
	}
	defer f.Close()
	e := json.NewEncoder(f)
	for {
		// 从通道里拉数据
		r := <-s.record
		if r.Key == "" && r.URL == "" {
			continue
		}
		log.Printf("get record [%s:%s] \n", r.Key, r.URL)
		if err = e.Encode(r); err != nil {
			log.Printf("saveLoop saving to URLStore err: %v \n", err)
		}
	}
}

// Close 关闭chan
func (s *URLStore) Close() {
	close(s.record)
}