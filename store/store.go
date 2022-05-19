/**
 * @Time : 2022/5/19 7:06 下午
 * @Author : frankj
 * @Email : frankxjkuang@gmail.com
 * @Description : --
 * @Revise : --
 */

package store

type Store interface {
	Put(url, key *string) error
	Get(key, url *string) error
}