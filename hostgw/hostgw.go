package hostgw

import (
	"fmt"
	"net"

	"github.com/cloudflare/cfssl/log"
	"github.com/easykube/route"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/cache"
)

const (
	Namespace = "easyrouter.easykube.com"
)

type Handler struct {
	Node          string   //当前节点的名称
	PublicIp      *net.IP  //当前节点公共ip
	EtcdEndPoints []string //etcd连接串
}

func server2route(s *v1.Node) (*route.Route, error) {
	subnetStr := s.Spec.PodCIDR
	if subnetStr == "" && s.Labels != nil {
		sip, _ := s.Labels[Namespace+"/subnet-ip"]
		smask, _ := s.Labels[Namespace+"/subnet-mask"]
		if sip != "" && smask != "" {
			subnetStr = sip + "/" + smask
		}
	}
	if subnetStr == "" {
		return nil, fmt.Errorf("从节点信息中获取pod子网失败: %+v", s)
	}

	address := ""
	for _, addr := range s.Status.Addresses {
		if addr.Type == v1.NodeInternalIP {
			address = addr.Address
		}
	}

	if address == "" {
		return nil, fmt.Errorf("从节点信息中获取ip失败: %+v", s)
	}

	_, subnet, _ := net.ParseCIDR(subnetStr)

	ip := net.ParseIP(address)

	return &route.Route{
		Gw:  ip,
		Dst: subnet,
	}, nil
}

//var i int = 0

func (h *Handler) OnAdd(o interface{}) {

	s := o.(*v1.Node)
	/**
	datajson, err := json.Marshal(s)
	if err != nil {
		println(err)
	} else {
		i++
		//println(*(*string)(unsafe.Pointer(&datajson)))
	    io.WriteFile("i:/node"+strconv.Itoa(i)+".json", datajson, 0777)
	}**/
	//如果是本机的节点，不用更新路由
	if s.Name == h.Node {
		return
	}
	r, err := server2route(s)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	ip := GetIpByNode(h.EtcdEndPoints, s.Name)
	if ip != "" {
		r.Gw = net.ParseIP(ip)
	}

	log.Infof("添加路由 %+v", r)
	if err = route.RouteAdd(r); err != nil {
		log.Errorf("添加路由失败: %+v\n%+v", r, err)
	}
}

func (h *Handler) OnDelete(o interface{}) {
	s, ok := o.(*v1.Node)
	if !ok {
		tmp, ok := o.(cache.DeletedFinalStateUnknown)
		if !ok {
			log.Infof("未知事件: %+v", o)
			return
		}
		s = tmp.Obj.(*v1.Node)
	}
	r, err := server2route(s)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	ip := GetIpByNode(h.EtcdEndPoints, s.Name)
	if ip != "" {
		r.Gw = net.ParseIP(ip)
	}

	log.Infof("删除路由 %+v", r)
	if err = route.RouteDel(r); err != nil {
		log.Errorf("删除路由失败: %+v\n%+v", r, err)
	}
}

func (h *Handler) OnUpdate(old, new interface{}) {
	oldNode, ok := old.(*v1.Node)
	if !ok {
		return
	}
	newNode, ok := new.(*v1.Node)
	if !ok {
		return
	}

	//如果是本机的节点，不用更新路由
	if oldNode.Name == h.Node {
		return
	}

	oldRoute, err := server2route(oldNode)
	newRoute, err2 := server2route(newNode)

	if err == nil && err2 == nil && route.RouteEquals(oldRoute, newRoute) {
		return
	}

	if err == nil {
		h.OnDelete(old)
	}
	if err2 == nil {
		h.OnAdd(new)
	}
}
