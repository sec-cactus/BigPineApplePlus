# BigPineApplePlus
大菠萝，起名灵感源于游戏中的M249，主要功能是：通过监听流量，发送双向阻断包，达到阻断链接、封禁IP的目的。

其实通过防火墙、IPS包括WAF都可以形成达到类似的效果，但是都会遇到性能瓶颈，即封禁的IP地址/IP地址段数量较多时，会大量消耗设备性能，造成网络延迟，从而影响业务。</br>
从旁路发送阻断包的方式，则完全不涉及链路设备的性能问题。只要发包速度足够快，就可以达到良好的封禁效果。更重要的是，这种方式具备很好的横向扩展能力。当流量压力较大时，可以通过tap设备将流量分给多个旁路阻断设备，实现扩展。</br>
在网络安全攻防演练活动中，与IP黑名单联动使用，可以形成攻击动态封禁能力。

本项目是BigPineApple项目的改进版，原项目链接：https://github.com/sec-cactus/BigPineApple

主要改进内容：

1. 使用redis作为读写配置，包括监听网卡、IP黑白名单等等；使BigPineApplePlus作为旁路阻断模块，可以与其他模块集成，如威胁情报、SOC、IDPS等等。
2. 对代码进行了优化，包括线程安全和代码精简等。

# 部署方式

## 网卡

旁路阻断基于流量镜像，因此主机需要两块网卡，一个接受镜像流量，另一个发送阻断包。

## 主机环境

本项目基于CentOS7.6，需要安装go语言环境、libpcap和redis。

### 安装go环境

参考：https://www.cnblogs.com/yunfan1024/p/11362947.html
    
```
wget https://dl.google.com/go/go1.13.4.linux-amd64.tar.gz
tar -C /usr/local -xzvf go1.13.4.linux-amd64.tar.gz 
vi /etc/profile
export PATH=$PATH:/usr/local/go/bin
source /etc/profile
go version
go env
```

### 安装libpcap

参考：http://linux.it.net.cn/CentOS/course/2016/0504/21303.html
    
```
yum -y install gcc-c++
yum -y install flex 
yum -y install bison
wget http://www.tcpdump.org/release/libpcap-1.9.1.tar.gz
make
make install
```

安装libpcap如果遇到：“error while loading shared libraries: libpcap.so.1: cannot open shared object file: No such file or directory”，则:
1. 执行`whereis libpcap.so.1`查找位置
2. 在/etc/ld.so.conf中最后一行添加该位置（如：`/usr/local/lib`）
3. `sudo ldconfig`
参考：https://blog.csdn.net/LFGxiaogang/article/details/73287152

### 安装redis

```
yum install -y epel-release
yum install -y redis
```

## 编译BigPineApplePlus

```
git clone https://github.com/sec-cactus/BigPineApple.git
cd BigPineApple
go get -d -v
go build
```

## 配置

### 配置redis

需要通过redis配置启动参数及IP黑白名单，参见redis_script.txt

`HMSET sysconfig mirrornetworkdevice "ens37" mgtnetworkdevice "ens33" srcmac "00:0c:29:f0:1a:f6" dstmac "d4:a1:48:96:6a:3c" `

```
mirrornetworkdevice: 监听网卡
mgtnetworkdevice： 管理网卡，即发送阻断包网卡
srcmac： 源MAC地址，即管理网卡MAC地址
dstmac： 目的MAC地址，即管理网卡所在网段的网关MAC地址
```
### 配置监听网卡

需要将监听网卡配置为混杂模式

`ifconfig ens37 promisc`

    - 配置redis连接：修改redisconf.txt，配置redis连接的地址和端口。redisconf.txt将作为应用启动的配置文件。

## 启动

`./BigPineApplePlus -c redisconf.txt`

# 基本原理

## 阻断原理

* 建立TCP连接时，双方需进行三次握手。当双方完成握手、准备传输应用数据时，对双方发送reset包，可以起到阻断的效果。

* 在握手过程中，双方对序列号分别计数，因此想要发送有效的reset包，必须基于TCP协议准确计算双方的序列号。

* 如果只成功发送单一方向的阻断包，则另一方向会产生大量重传数据包。为了达到良好的阻断效果，至少在第二次握手后才具备发送双向阻断包的条件。

* 同时，还要考虑发送速率，阻断包必须先于正常数据包达到。考虑到监听后还需要进行IP地址比对和阻断包构造等流程，而且源地址在收到第二次握手后会立即发送第三次握手和应用数据两个数据包，如果选择在第二次握手后发送阻断包，则反向阻断包必然晚于应用数据包。此时TCP序列号受应用数据包长度影响，已经发生变化。因此，在双方完成握手并发送第一个应用数据包时进行阻断，是比较理想的时机。

## 性能设计

* 本项目主要针对大量IP封禁场景。如果使用传统比对字符串的方式，则会大大增加内存占用并降低比对效率。本项目参考“bitmap”模式，将全部IPV4地址映射到一段连续的内存空间，每个地址对应一个bit，因此全部IPV4地址占用空间可以控制在512MB以内。在比对过程中直接进行内存比对，具有较好的效率。

* 本项目目前使用了libpcap作为监听流量数据包的方式。如果遇到带宽瓶颈、丢包严重，则通过tap设备将流量均分给多个旁路设备。

* 本项目使用了redis进行配置读取，主要考虑便于在接收威胁情报、SOC、IDPS等外部模块的IP黑白名单配置时便于高速读写，从而实现IP动态封禁。

## 运行设计

* 运行过程中，每1分钟自动从redis更新一次IP黑白名单配置。

* 运行过程中，每1分钟自动记录IP黑名单命中情况，写入本地log.txt文件。

* 第二次TCP握手后，在内存中记录本次连接的反向数据包的TCP序列号。如100毫秒内未能完成阻断包发送，则删除此记录，避免占用过多内存。此机制目的在于对丢包场景提供一定的容错能力。

# 改进方向

本项目扔存在许多不足，欢迎指正。除了编码、性能进一步优化外，主要有两个改进方向：

* 支持IPV6

* 支持pfring、DPDK等高性能网卡驱动。



