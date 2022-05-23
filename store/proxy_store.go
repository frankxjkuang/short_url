/**
 * @Time : 2022/5/19 6:26 下午
 * @Author : frankj
 * @Email : frankxjkuang@gmail.com
 * @Description : --
 * @Revise : --
 */

package store

import (
	"log"
	"net/rpc"
)

type ProxyStore struct {
	urls   *URLStore // local cache
	client *rpc.Client
}

func NewProxyStore(addr string) *ProxyStore {
	client, err := rpc.DialHTTP("tcp", addr)
	if err != nil {
		log.Printf("Error constructing ProxyStore: %v \n", err)
		panic(err)
	}
	return &ProxyStore{urls: NewURLStore(""), client: client}
}

func (s *ProxyStore) Get(key, url *string) error {
	if err := s.urls.Get(key, url); err == nil {
		return nil
	}
	// rpc call to master:
	if err := s.client.Call("Store.Get", key, url); err != nil {
		return err
	}
	s.urls.Set(key, url) // update local cache
	return nil
}

func (s *ProxyStore) Put(url, key *string) error {
	// rpc call to master:
	if err := s.client.Call("Store.Put", url, key); err != nil {
		return err
	}
	s.urls.Set(key, url) // update local cache
	return nil
}

// GetUrls 获取短长链映射
func (s *ProxyStore) GetUrls() map[string]string {
	return s.urls.urls
}
