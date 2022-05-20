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
	"net/rpc"
	"os"
	"os/signal"
	"short_url/store"
	"syscall"
	"time"
)

var (
	// 启动端口
	port = flag.String("port", ":8080", "http listen port")
	// 启动文件-持久化存储映射kv
	file = flag.String("file", "store.json", "data store filename")
	// host地址
	host = flag.String("host", "127.0.0.1", "hostname")
	// 是否开启rpc
	rpcEnabled = flag.Bool("rpc", false, "enable rpc service")
	// 以master服务启动服务
	masterAddr = flag.String("master", "", "rpc master address")
)

var urlStore store.Store

type myHandleFunc func(http.ResponseWriter, *http.Request)

func init() {
	flag.Parse()
	//urlStore = store.NewURLStore(*file)

	if *masterAddr != "" {
		urlStore = store.NewProxyStore(*masterAddr)
	} else {
		urlStore = store.NewURLStore(*file)
	}
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
	http.HandleFunc("/get", logPanics(Get))

	//http.HandleFunc("/delete", logPanics(Delete))

	httpSvr := http.Server{
		Addr:              *host + *port,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// 开启 http 服务
	go func() {
		log.Printf("Starting ListenAndServe: %s%s", *host, *port)
		err := httpSvr.ListenAndServe()
		if err == http.ErrServerClosed {
			err = nil
		}
		if err != nil {
			log.Printf("Http serve close: %v \n", err)
		}
	}()

	// 开启 rpc 服务
	if *rpcEnabled {
		rpc.RegisterName("Store", urlStore)
		rpc.HandleHTTP()
	}

	Watch(func() error {
		ctx, cancel := context.WithTimeout(context.Background(),5*time.Second)
		defer cancel()
		return httpSvr.Shutdown(ctx)
	}, func() error {
		//urlStore.Close()
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

// Put 新增：curl -X POST '127.0.0.1:8080/put' --form 'url=https://www.kxjun.top'
func Put(w http.ResponseWriter, r *http.Request) {
	var key string
	// 测试优雅退出
	// time.Sleep(3 * time.Second)
	url := r.FormValue("url")
	if url == "" {
		fmt.Fprintf(w, "url can not be ''")
		return
	}
	if err := urlStore.Put(&url, &key); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Successed shortURL(short: long) is %s: %v \n", key, url)
}

// Get 获取：curl -X POST '127.0.0.1:8080/get' --form 'key=1'
func Get(w http.ResponseWriter, r *http.Request) {
	var (
		key = r.FormValue("key")
		url string
	)

	if key == "" {
		fmt.Fprintf(w, "key can not be ''")
		return
	}
	if err := urlStore.Get(&key, &url); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Successed shortURL(short: long) is %s: %v \n", key, url)
}

//func Delete(w http.ResponseWriter, r *http.Request) {
//	key := r.FormValue("key")
//	if key == "" {
//		w.Header().Set("Content-Type", "text/html")
//		return
//	}
//	urlStore.Delete(key)
//	w.Header().Set("Content-Type", "application/json")
//	// 重定向刷新回去
//	http.Redirect(w, r, fmt.Sprintf("%s%s", *host, *port), http.StatusTemporaryRedirect)
//}

func Index(w http.ResponseWriter, r *http.Request) {
	//解析指定文件生成模板对象
	tmpl, err := template.ParseFiles("./html/index.html")
	if err != nil {
		log.Fatal(err)
	}
	//利用给定数据渲染模板，并将结果写入
	//err = tmpl.Execute(w, urlStore.GetUrls())
	err = tmpl.Execute(w, map[string]string{"1": "2"})
	if err != nil {
		log.Fatal(err)
	}
}

func Redirect(w http.ResponseWriter, r *http.Request) {
	var url string
	key := r.URL.Path[1:]
	if key == "" {
		return
	}
	if err := urlStore.Get(&key, &url); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if url == "" {
		http.NotFound(w, r)
		return
	}
	// http.StatusFound       302 临时性重定向
	// StatusMovedPermanently 301 永久性的移动
	http.Redirect(w, r, url, http.StatusFound)
}
