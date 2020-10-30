package mqc

import (
	"fmt"

	"github.com/asaskevich/govalidator"
	"github.com/micro-plat/hydra/conf"
	"github.com/micro-plat/hydra/global"
)

//MainConfName 主配置中的关键配置名
var MainConfName = []string{"status", "sharding"}

//SubConfName 子配置中的关键配置名
var SubConfName = []string{"queue"}

//Server mqc服务配置
type Server struct {
	Status   string `json:"status,omitempty" valid:"in(start|stop)" toml:"status,omitempty"`
	Sharding int    `json:"sharding,omitempty" toml:"sharding,omitempty"`
	Addr     string `json:"addr,omitempty" valid:"required"  toml:"addr,omitempty"`
	Trace    bool   `json:"trace,omitempty" toml:"trace,omitempty"`
	Timeout  int    `json:"timeout,omitzero" toml:"timeout,omitzero"`
}

//New 构建mqc server配置，默认为对等模式
func New(addr string, opts ...Option) *Server {
	if _, _, err := global.ParseProto(addr); err != nil {
		panic(fmt.Errorf("mqc服务器地址配置有误，必须是:proto://addr 格式 %w", err))
	}
	s := &Server{Addr: addr, Status: "start"}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

//GetConf 获取主配置信息
func GetConf(cnf conf.IServerConf) (*Server, error) {
	s := Server{}
	if cnf.GetServerType() != global.MQC {
		return nil, fmt.Errorf("mqc主配置类型错误:%s != mqc", cnf.GetServerType())
	}

	_, err := cnf.GetMainObject(&s)
	if err != nil && err != conf.ErrNoSetting {
		return nil, fmt.Errorf("mqc服务器配置格式有误:%v", err)
	}
	if b, err := govalidator.ValidateStruct(&s); !b {
		return nil, fmt.Errorf("mqc服务器配置数据有误:%v %v", err, s)
	}
	return &s, nil
}
