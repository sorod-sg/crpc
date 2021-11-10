package registry

import (
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

//服务注册中心
type Registry struct {
	timeout time.Duration //超时时间
	mu      sync.Mutex
	servers map[string]*ServerItem //服务列表
}

type ServerItem struct {
	Addr  string    //地址
	start time.Time //启动时间
}

const (
	defaultPath    = "/_crpc_/registy"
	defaultTimeOut = time.Minute * 5 //默认超时时间
)

//新建注册中心
func New(timeout time.Duration) *Registry {
	return &Registry{
		servers: make(map[string]*ServerItem),
		timeout: timeout,
	}
}

//注册中心添加服务
func (r *Registry) putServer(addr string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s := r.servers[addr]
	if s == nil {
		r.servers[addr] = &ServerItem{
			Addr:  addr,
			start: time.Now(),
		}
	} else {
		s.start = time.Now()
	}
}

//返回所有可用服务
func (r *Registry) aliveServers() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var alive []string
	for addr, s := range r.servers {
		if r.timeout == 0 || s.start.Add(r.timeout).After(time.Now()) {
			alive = append(alive, addr)
		} else {
			delete(r.servers, addr)
		}

	}
	sort.Strings(alive)
	return alive
}

//实现http通信
// GET :返回所有可用的server列表
// POST :添加服务
func (r *Registry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		w.Header().Set("X-crpc-Servers", strings.Join(r.aliveServers(), ","))
	case "Post":
		addr := req.Header.Get("X-crpc-Servers")
		if addr == "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r.putServer(addr)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

//处理http请求
func (r *Registry) HandleHTTP(registryPath string) {
	http.Handle(registryPath, r)
	log.Println("rpc registry path:", registryPath)
}

//默认注册中心
var DefaultRegister = &Registry{
	timeout: defaultTimeOut,
	servers: make(map[string]*ServerItem),
}

//使用默认注册中心处理http请求
func HandleHTTP() {
	DefaultRegister.HandleHTTP(defaultPath)
}

//心跳注册,默认心跳时间为defaultTimeOut - 1
func Heartbeat(register, addr string, duration time.Duration) {
	if duration == 0 {
		duration = defaultTimeOut - time.Duration(1)*time.Minute
	}
	var err error
	err = sendHeartbeat(register, addr)
	go func() {
		t := time.NewTicker(duration)
		for err == nil {
			<-t.C
			err = sendHeartbeat(register, addr)
		}
	}()
}

//发送心跳
func sendHeartbeat(registry, addr string) error {
	log.Println(addr, "send heart beat to regisety", registry)
	httpClient := &http.Client{}
	req, _ := http.NewRequest("post", registry, nil)
	req.Header.Set("X-crpc-Server", addr)
	if _, err := httpClient.Do(req); err != nil {
		log.Println("rpc server :heart beat err:", err)
		return err
	}
	return nil
}
