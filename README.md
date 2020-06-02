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
1. 旁路阻断基于流量镜像，因此主机需要两块网卡，一个接受镜像流量，另一个发送阻断包。
2. 本项目基于CentOS7.6，需要安装go语言环境、libpcap和redis。

安装go环境，参考：https://www.cnblogs.com/yunfan1024/p/11362947.html
```
wget https://dl.google.com/go/go1.13.4.linux-amd64.tar.gz
tar -C /usr/local -xzvf go1.13.4.linux-amd64.tar.gz 
vi /etc/profile
export PATH=$PATH:/usr/local/go/bin
source /etc/profile
go version
go env
```

安装libpcap，参考：http://linux.it.net.cn/CentOS/course/2016/0504/21303.html
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

安装redis：
```
yum install -y epel-release
yum install -y redis
```




