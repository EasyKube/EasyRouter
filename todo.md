>使用命令行添加删除路由,便于跨平台
>使用各操作系统平台的功能构造路由器
>指定节点名称 --node　 用于替换hostname
>指定节点ip  --address 用于替换从自动获取的address


查询服务状态
sc query RemoteAccess

SERVICE_NAME: RemoteAccess
        TYPE               : 20  WIN32_SHARE_PROCESS
        STATE              : 1  STOPPED
        WIN32_EXIT_CODE    : 1077  (0x435)
        SERVICE_EXIT_CODE  : 0  (0x0)
        CHECKPOINT         : 0x0
        WAIT_HINT          : 0x0

sc config RemoteAccess

获取节点
etcdctl ls /registry/minions/wsc2016
查询etcd注册的信息
http://192.168.0.233:2379/v2/keys/

