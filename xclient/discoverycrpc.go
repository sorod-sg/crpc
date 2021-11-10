package xclient

import (
	"log"
	"net/http"
	"strings"
	"time"
)

//存在注册中心的多服务发现端
type RegisetyDiscovery struct {
	*MutiServerDiscovery
	registry   string //注册中心
	timeout    time.Duration
	lastUpdate time.Time //最后更新时间
}

const defaultUpdateTimeout = time.Second * 10

func NewRegisteyDiscovery(registryAddr string, timeout time.Duration) *RegisetyDiscovery {
	if timeout == 0 {
		timeout = defaultUpdateTimeout
	}
	d := &RegisetyDiscovery{
		MutiServerDiscovery: NewMuilServerDiscovery(make([]string, 0)),
		registry:            registryAddr,
		timeout:             timeout,
	}
	return d
}

//更新服务器列表
func (d *RegisetyDiscovery) Update(servers []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.servers = servers
	d.lastUpdate = time.Now()
	return nil
}

//刷新保证服务没有过期
func (d *RegisetyDiscovery) Refresh() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.lastUpdate.Add(d.timeout).After(time.Now()) {
		return nil
	}
	log.Println("rpc registry: refresh servers from registry", d.registry)
	resp, err := http.Get(d.registry)
	if err != nil {
		log.Println("rpc registry refresh err:", err)
		return err
	}
	servers := strings.Split(resp.Header.Get("X-crpc-Servers"), ",")
	d.servers = make([]string, 0, len(servers))
	for _, server := range servers {
		if strings.TrimSpace(server) != "" {
			d.servers = append(d.servers, strings.TrimSpace(server))
		}
	}
	d.lastUpdate = time.Now()
	return nil
}

//通过SelectMode获取服务器
func (d *RegisetyDiscovery) Get(mode SelectMode) (string, error) {
	if err := d.Refresh(); err != nil {
		return "", err
	}
	return d.MutiServerDiscovery.Get(mode)
}

//获取所有服务器
func (d *RegisetyDiscovery) GetAll() ([]string, error) {
	if err := d.Refresh(); err != nil {
		return nil, err
	}
	return d.MutiServerDiscovery.GetAll()
}
