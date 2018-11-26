package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	rr "transfer-http/roundrobin"

	"github.com/pquerna/ffjson/ffjson"
)

var RR rr.RR
var ConfData map[string]interface{}

func init() {
	RR = rr.NewWeightedRR(rr.RR_NGINX)
	GetProxyConf()
}

type handle struct {
	addrs []string
}

func (this *handle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	addr := RR.Next().(string)
	conf := ConfData[addr].(map[string]interface{})
	remote, err := url.Parse("http://" + addr + "/" + conf["path"].(string))
	if err != nil {
		panic(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.ServeHTTP(w, r)
}

/* 读取proxy配置 */
func GetProxyConf() {
	if f, err := os.Open("./proxy.json"); err == nil {
		defer f.Close()
		if v, err := ioutil.ReadAll(f); err == nil {
			ffjson.Unmarshal(v, &ConfData)
		} else {
			log.Println(err.Error())
		}
	} else {
		log.Println(err.Error())
	}
}

/* 开始代理 */
func startServer() {

	//被代理的服务器host和port
	h := &handle{}
	//h.addrs = []string{"localhost:8080"}
	for k, _ := range ConfData {
		h.addrs = append(h.addrs, k)
	}

	w := 1
	for _, e := range h.addrs {
		RR.Add(e, w)
		w++
	}
	err := http.ListenAndServe(":80", h)
	if err != nil {
		log.Fatalln("ListenAndServe: ", err)
	}
}

func main() {
	startServer()
}
