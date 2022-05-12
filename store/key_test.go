/**
 * @Time : 2022/5/11 5:07 下午
 * @Author : frankj
 * @Email : kuangxj@dustess.com
 * @Description : --
 * @Revise : --
 */

package store

import "testing"

func TestGenKey(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Logf("%d genKey is %s", i, genKey(i))
	}
}