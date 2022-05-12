/**
 * @Time : 2022/5/11 5:34 下午
 * @Author : frankj
 * @Email : frankxjkuang@gmail.com
 * @Description : --
 * @Revise : --
 */

package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"short_url/store"
)

var urlStore = store.NewURLStore()

type myHandleFunc func(http.ResponseWriter, *http.Request)

func logPanics(m myHandleFunc) myHandleFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[%v] caught panic: %v", request.RemoteAddr, r)
			}
		}()
		m(writer, request)
	}
}

func main()  {
	// 首页
	http.HandleFunc("/index", Index)

	// 重定向
	http.HandleFunc("/", Redirect)

	// 新增
	http.HandleFunc("/put", logPanics(Put))

	// 获取
	// curl POST '127.0.0.1:9009/get' --form 'url="hello"'
	http.HandleFunc("/get", logPanics(Get))

	http.HandleFunc("/delete", logPanics(Delete))

	if err := http.ListenAndServe(":9009", nil); err != nil {
		// fmt.Printf("http ListenAndServe err[%v]", err)
		panic(err)
	}
}

// curl POST '127.0.0.1:9009/put' --form 'url="hello"'
func Put(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue("url")
	if url == "" {
		fmt.Fprintf(w, "url can not be ''")
		return
	}
	key := urlStore.Put(url)
	fmt.Fprintf(w, "Successed shortURL(short: long) is %s: %v \n", key, url)
}

// curl POST '127.0.0.1:9009/get' --form 'key="hello"'
func Get(w http.ResponseWriter, r *http.Request) {
	key := r.FormValue("key")

	if key == "" {
		fmt.Fprintf(w, "key can not be ''")
		return
	}
	longURL := urlStore.Get(key)

	fmt.Fprintf(w, "Successed shortURL(short: long) is %s: %v \n", key, longURL)
	//key := r.FormValue("key")
	//result := map[string]string{}
	//w.Header().Set("Content-Type","application/json")
	//if key == "" {
	//	result["msg"] = "key is ''"
	//} else {
	//	longURL := urlStore.Get(key)
	//	result[key] = longURL
	//}
	//
	//ret, err := json.Marshal(result)
	//if err != nil {
	//	result["error"] = err.Error()
	//}
	//w.Write(ret)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	key := r.FormValue("key")
	//result := map[string]string{"hello": "world"}
	if key == "" {
		w.Header().Set("Content-Type", "text/html")
		//fmt.Fprint(w, AddForm)
		return
	}
	urlStore.Delete(key)
	w.Header().Set("Content-Type","application/json")
	//ret, _ := json.Marshal(result)
	//w.Write(ret)
	// 重定向刷新回去
	http.Redirect(w, r, "127.0.0.1:9009", http.StatusTemporaryRedirect)
}

func Index(w http.ResponseWriter, r *http.Request) {
	//解析指定文件生成模板对象
	tmpl, err := template.ParseFiles("./html/index.html")
	if err != nil {
		log.Fatal(err)
	}
	//urls := map[string]string{
	//	"0": "http:1",
	//	"1": "http:1",
	//	"2": "http:1",
	//}
	//利用给定数据渲染模板，并将结果写入
	err = tmpl.Execute(w, urlStore.GetUrls())
	if err != nil {
		log.Fatal(err)
	}
}

func Redirect(w http.ResponseWriter, r *http.Request)  {
	key := r.URL.Path[1:]
	if key == "" {
		return
	}
	url := urlStore.Get(key)
	if url == "" {
		http.NotFound(w, r)
		return
	}
	// http.StatusFound       302 临时性重定向
	// StatusMovedPermanently 301 永久性的移动
	http.Redirect(w, r, url, http.StatusFound)
}