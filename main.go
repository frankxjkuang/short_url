/**
 * @Time : 2022/5/11 5:34 下午
 * @Author : frankj
 * @Email : frankxjkuang@gmail.com
 * @Description : --
 * @Revise : --
 */

package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"short_url/store"
	"syscall"
	"time"
)

var (
	port = flag.String("port", ":9009", "http listen port")
	file = flag.String("file", "store.json", "data store filename")
	host = flag.String("host", "127.0.0.1", "hostname")
)

//var urlStore = store.NewURLStore("store.gob")
var urlStore *store.URLStore

type myHandleFunc func(http.ResponseWriter, *http.Request)

func init() {
	flag.Parse()
	urlStore = store.NewURLStore(*file)
}

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

func main() {
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

	httpSvr := http.Server{
		Addr:              *host + *port,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("ListenAndServe: %s%s", *host, *port)
		err := httpSvr.ListenAndServe()
		if err == http.ErrServerClosed {
			err = nil
		}
		if err != nil {
			log.Printf("Http serve close: %v \n", err)
		}
	}()

	Watch(func() error {
		ctx, cancel := context.WithTimeout(context.Background(),5*time.Second)
		defer cancel()
		return httpSvr.Shutdown(ctx)
	}, func() error {
		urlStore.Close()
		return nil
	})
}

func Watch(fns ...func() error) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)

	// 等待信号
	s := <-ch
	close(ch)
	log.Printf("Catch signal[%s], start shutdown func \n", s.String())
	for i := range fns {
		if err := fns[i](); err != nil {
			log.Println(err)
		}
	}
	log.Println("Serve exit.")
}

// curl POST '127.0.0.1:9009/put' --form 'url="hello"'
func Put(w http.ResponseWriter, r *http.Request) {
	// @TODO 测试优雅退出
	time.Sleep(3 * time.Second)

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
}

func Delete(w http.ResponseWriter, r *http.Request) {
	key := r.FormValue("key")
	if key == "" {
		w.Header().Set("Content-Type", "text/html")
		return
	}
	urlStore.Delete(key)
	w.Header().Set("Content-Type", "application/json")
	// 重定向刷新回去
	http.Redirect(w, r, fmt.Sprintf("%s%s", *host, *port), http.StatusTemporaryRedirect)
}

func Index(w http.ResponseWriter, r *http.Request) {
	//解析指定文件生成模板对象
	tmpl, err := template.ParseFiles("./html/index.html")
	if err != nil {
		log.Fatal(err)
	}
	//利用给定数据渲染模板，并将结果写入
	err = tmpl.Execute(w, urlStore.GetUrls())
	if err != nil {
		log.Fatal(err)
	}
}

func Redirect(w http.ResponseWriter, r *http.Request) {
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
