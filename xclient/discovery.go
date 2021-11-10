package xclient

import (
	"errors"
	"math"
	"math/rand"
	"sync"
	"time"
)

type SelectMode int

const (
	RandomSelect     SelectMode = iota //随机选择
	RoundRobinSelect                   //轮询选择
)

//服务发现接口,包含服务发现的最基本需求
type Discovery interface {
	Refresh() error
	Update(server []string) error
	Get(mode SelectMode) (string, error)
	GetAll() ([]string, error)
}

//无需注册中心的多服务发现端
type MutiServerDiscovery struct {
	r       *rand.Rand   //随机数实例
	mu      sync.RWMutex //读写锁
	servers []string     //服务器列表
	index   int          //轮询索引
}

//新建MutiserverDiscovery
func NewMuilServerDiscovery(servers []string) *MutiServerDiscovery {
	d := &MutiServerDiscovery{
		servers: servers,
		r:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	d.index = d.r.Intn(math.MaxInt32 - 1)
	return d
}

var _ Discovery = (*MutiServerDiscovery)(nil) //判断接口实现与否

func (d *MutiServerDiscovery) Refresh() error {
	return nil
}

//更新服务器列表
func (d *MutiServerDiscovery) Update(servers []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.servers = servers
	return nil
}

//通过selectmode获取服务器
func (d *MutiServerDiscovery) Get(mode SelectMode) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	n := len(d.servers)
	if n == 0 {
		return "", errors.New("rpc discovery : not available servers")
	}
	switch mode {
	case RandomSelect:
		return d.servers[d.r.Intn(n)], nil
	case RoundRobinSelect:
		s := d.servers[d.index%n]
		d.index = (d.index + 1) % n
		return s, nil
	default:
		return "", errors.New("rpc discovery not supported select mode")
	}
}

//返回列表中的所有服务器
func (d *MutiServerDiscovery) GetAll() ([]string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	servers := make([]string, len(d.servers), len(d.servers))
	copy(servers, d.servers)
	return servers, nil
}
