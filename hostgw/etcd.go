package hostgw

import (
	"context"
	"log"
	"time"

	"github.com/coreos/etcd/client"
)

func RegNode(etcdEndpoints []string, node string, ip string) {
	cfg := client.Config{
		Endpoints:               etcdEndpoints,
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}
	etcdClient, err := client.New(cfg)
	if err != nil {
		log.Fatal("连接etcd出错:", err)
		return
	}

	api := client.NewKeysAPI(etcdClient)
	_, err = api.Set(context.Background(), "/k8s_easy_router/node/"+node+"/ip", ip, &client.SetOptions{})
	if err != nil {
		log.Fatal("etcd设值出错:", err)
		return
	}

}

func UnRegNode(etcdEndpoints []string, node string) {
	cfg := client.Config{
		Endpoints:               etcdEndpoints,
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}
	etcdClient, err := client.New(cfg)
	if err != nil {
		log.Fatal("连接etcd出错:", err)
		return
	}
	api := client.NewKeysAPI(etcdClient)
	_, err = api.Delete(context.Background(), "/k8s_easy_router/node/"+node, &client.DeleteOptions{Recursive: true, Dir: true})

	if err != nil {
		log.Fatal("etcd删除出错:", err)
		return
	}
}

func GetIpByNode(etcdEndpoints []string, node string) string {
	cfg := client.Config{
		Endpoints:               etcdEndpoints,
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}
	etcdClient, err := client.New(cfg)
	if err != nil {
		log.Fatal("连接etcd出错:", err)
		return ""
	}

	api := client.NewKeysAPI(etcdClient)
	r, err := api.Get(context.Background(), "/k8s_easy_router/node/"+node+"/ip", &client.GetOptions{})
	if err != nil {
		log.Fatal("etcd读取出错:", err)
		return ""
	}
	return r.Node.Value

}
