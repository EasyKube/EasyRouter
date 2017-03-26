package main

import (
	goflag "flag"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"strconv"
	"time"

	flag "github.com/spf13/pflag"

	"github.com/cloudflare/cfssl/log"
	"github.com/easykube/easyrouter/hostgw"
	"github.com/easykube/route"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/fields"
	"k8s.io/client-go/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func EasyRouterCli() {

	flags := flag.NewFlagSet("", flag.ExitOnError)
	overrides := &clientcmd.ConfigOverrides{}
	overrideFlags := clientcmd.RecommendedConfigOverrideFlags("")
	clientcmd.BindOverrideFlags(overrides, flags, overrideFlags)

	// Parse flags
	resync := flags.Int("resync", 30, "Resync period in seconds")
	incluster := flags.Bool("in-cluster", false, "If this in run inside a pod")
	profile := flags.Bool("profile", false, "Enable profiling")
	address := flags.String("profile_host", "localhost", "Profiling server host")
	port := flags.Int("profile_port", 9801, "Profiling server port")
	test := flags.Bool("test", false, "Dry-run. To test if the binary is complete")
	//节点相关参数
	etcdServers := flags.String("etcd_servers", "", "etcd服务器地址")
	node := flags.String("node", "", "当前节点的名称")
	publicip := flags.IP("public_ip", nil, "当前节点公共ip")
	eth := flags.String("eth", "eth0", "进行路由的网卡名称")

	flags.AddGoFlagSet(goflag.CommandLine)
	flags.Parse(os.Args)

	if *test {
		return
	}

	// Set up profiling server
	if *profile {
		go func() {
			mux := http.NewServeMux()
			mux.HandleFunc("/debug/pprof/", pprof.Index)
			mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
			mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)

			server := &http.Server{
				Addr:    net.JoinHostPort(*address, strconv.Itoa(*port)),
				Handler: mux,
			}
			log.Fatalf("%+v", server.ListenAndServe())
		}()
	}

	// Create kubeconfig
	var (
		clientConfig *rest.Config
		err          error
	)
	if *incluster {
		clientConfig, err = rest.InClusterConfig()
	} else {

		kubeconfig := clientcmd.NewDefaultClientConfig(*clientcmdapi.NewConfig(), overrides)
		clientConfig, err = kubeconfig.ClientConfig()
	}

	if err != nil {
		log.Fatalf("创建k8s配置出错: %+v", err)
	}

	// Create kubeclient
	k8s, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		log.Fatalf("k8s的api配置无效: %+v", err)
	}

	// Create listwatch instance
	listwatch := cache.NewListWatchFromClient(k8s.Core().RESTClient(), "nodes", api.NamespaceAll, fields.Everything())
	handler := &hostgw.Handler{}
	handler.Node = *node
	handler.PublicIp = publicip
	handler.EtcdEndPoints = []string{*etcdServers}
	if handler.Node == "" {
		handler.Node, err = os.Hostname()
	}

	hostgw.RegNode(handler.EtcdEndPoints, handler.Node, handler.PublicIp.String())
	// Create informer
	_, informer := cache.NewInformer(listwatch, &v1.Node{}, time.Second*(time.Duration)(*resync), handler)

	// Handle signals (optional)
	// Start watching
	log.Infof("EasyRouter正在启动...")
	route.InitRouter(*eth)
	_, err = k8s.Core().Nodes().List(v1.ListOptions{})
	if err != nil {
		log.Fatalf("不能连接k8s")
	}
	informer.Run(wait.NeverStop)
	hostgw.UnRegNode(handler.EtcdEndPoints, handler.Node)
	log.Infof("结束")

}
