/**
 * @Time : 2022/5/11 4:09 下午
 * @Author : frankj
 * @Email : frankxjkuang@gmail.com
 * @Description : --
 * @Revise : --
 */

package store

import "sync"

var urlStore *URLStore
var once sync.Once

type URLStore struct {
	urls map[string]string // 短长链的映射关系
	sync.RWMutex
}

// NewURLStore 初始化URLStore
func NewURLStore() *URLStore {
	once.Do(func() {
		urlStore = new(URLStore)
		urlStore.urls = make(map[string]string)
	})
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
			return key
		}
	}
	// shouldn’t get here
	return ""
}

// GetUrls 获取短长链映射
func (s *URLStore) GetUrls() map[string]string {
	return s.urls
}



