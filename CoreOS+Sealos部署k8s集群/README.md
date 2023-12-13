博文地址：https://github.com/wrype/blogs/tree/main/CoreOS%2BSealos%E9%83%A8%E7%BD%B2k8s%E9%9B%86%E7%BE%A4

<!-- TOC -->

- [前期准备](#前期准备)
  - [CoreOS 机器准备](#coreos-机器准备)
    - [**集群机器卸载 docker**](#集群机器卸载-docker)
    - [（可选）卸载云镜像自带的 containerd、runc](#可选卸载云镜像自带的-containerdrunc)
    - [**集群每台机器设置唯一的主机名**](#集群每台机器设置唯一的主机名)
    - [运行 sealos 的机器上做集群机器免密](#运行-sealos-的机器上做集群机器免密)
    - [上传`helm`、`sealos`二进制文件](#上传helmsealos二进制文件)
  - [kubernetes、cilium 镜像修改](#kubernetescilium-镜像修改)
    - [修改 sealos kubernetes 镜像](#修改-sealos-kubernetes-镜像)
    - [修改 sealos cilium 镜像](#修改-sealos-cilium-镜像)
  - [基础镜像列表](#基础镜像列表)
- [k8s 集群部署](#k8s-集群部署)
  - [Clusterfile 讲解](#clusterfile-讲解)
  - [镜像顺序](#镜像顺序)
  - [`/usr` 只读挂载处理](#usr-只读挂载处理)
  - [设置 CIDR](#设置-cidr)
  - [修改 chart values 文件](#修改-chart-values-文件)
  - [查看集群状态](#查看集群状态)
- [参考文档](#参考文档)
- [sealos 常用命令](#sealos-常用命令)
  - [sealos images](#sealos-images)
  - [sealos build --debug -t \<tag\> \<Kubefile\>](#sealos-build---debug--t-tag-kubefile)
  - [sealos pull \<img\>](#sealos-pull-img)
  - [sealos save -o \<tar\> -m --format docker-archive \<imgs\>](#sealos-save--o-tar--m---format-docker-archive-imgs)
  - [sealos load -i \<tar\>](#sealos-load--i-tar)
  - [sealos rmi \<imgs\>](#sealos-rmi-imgs)
  - [sealos add](#sealos-add)
  - [sealos delete](#sealos-delete)
  - [sealos reset](#sealos-reset)

<!-- /TOC -->

# 前期准备

## CoreOS 机器准备

### **集群机器卸载 docker**

CoreOS 的云镜像自带 docker，需要把他卸载掉，运行命令：`rpm-ostree override remove moby-engine`，重启机器

### （可选）卸载云镜像自带的 containerd、runc

运行命令：`rpm-ostree override remove containerd runc`，重启机器

### **集群每台机器设置唯一的主机名**

### 运行 sealos 的机器上做集群机器免密

### 上传`helm`、`sealos`二进制文件

将二进制文件解压放到 master-01 的 `/usr/local/bin` 下面，手动添加 completion：

```bash
cat <<EOF >> ~/.bash_profile
source <(sealos completion bash)
source <(sealctl completion bash)
source <(lvscare completion bash)
source <(helm completion bash)
EOF
```

## kubernetes、cilium 镜像修改

针对 CoreOS 需要对 kubernetes、cilium 镜像做一些修改，修改后把所有镜像都导出到 tar 包：

```bash
sealos save -o k8s1.26.tar -m --format docker-archive labring/kubernetes:v1.26.11-coreos labring/cilium:v1.13.4-coreos labring/metrics-server:v0.6.4 labring/nfs-subdir-external-provisioner:v4.0.18
```

### 修改 sealos kubernetes 镜像

```bash
sealos pull registry.cn-shanghai.aliyuncs.com/labring/kubernetes:v1.26.11
```

sealos rootfs 镜像中的脚本会把一些二进制文件复制到 `/usr/bin/` 下面，
然而 `/usr/bin/` 在 CoreOS 上是只读的，导致部署失败。

这里修改一下相关的脚本和 service 文件，重新打包镜像。修改过的文件在 [custom-k8s](./custom-k8s/) 下面，主要的改动是把 `/usr/bin/` 替换为 `/usr/local/bin/`。

> rootfs 源码仓库 https://github.com/labring-actions/runtime/tree/main/k8s

```bash
sealos build --debug -t labring/kubernetes:v1.26.11-coreos custom-k8s/
```

### 修改 sealos cilium 镜像

```bash
sealos pull registry.cn-shanghai.aliyuncs.com/labring/cilium:v1.13.4
```

sealos cilium 镜像的 Dockerfile CMD 里面把二进制文件复制到 `/usr/bin/` 下面，这里重新打包镜像，把 `/usr/bin/` 替换为 `/usr/local/bin/`。

> 源码仓库 https://github.com/labring-actions/cluster-image/tree/main/applications

```bash
sealos build --debug -t labring/cilium:v1.13.4-coreos custom-cilium/
```

## 基础镜像列表

- labring/kubernetes:v1.26.11
- labring/cilium:v1.13.4
- labring/metrics-server:v0.6.4
- labring/nfs-subdir-external-provisioner:v4.0.18

这里我们可以下载阿里云上的 sealos 镜像，加上仓库地址：`registry.cn-shanghai.aliyuncs.com`

# k8s 集群部署

```bash
# 导入镜像
sealos load -i k8s1.26.tar
# 部署集群，相关的临时文件在 ~/.sealos 下面
sealos apply -f Clusterfile
```

## [Clusterfile](./Clusterfile) 讲解

```yaml
apiVersion: apps.sealos.io/v1beta1
kind: Cluster
metadata:
  name: default
spec:
  hosts:
    - ips:
        - 192.168.56.104
        - 192.168.56.118
        - 192.168.56.119
      roles:
        - master
        - amd64
    - ips:
        - 192.168.56.120
        - 192.168.56.121
      roles:
        - node
        - amd64
  image:
    - labring/kubernetes:v1.26.11-coreos
    - labring/cilium:v1.13.4-coreos
    - labring/metrics-server:v0.6.4
    - labring/nfs-subdir-external-provisioner:v4.0.18
status: {}
---
apiVersion: kubeadm.k8s.io/v1beta3
kind: InitConfiguration
nodeRegistration:
  kubeletExtraArgs:
    volume-plugin-dir: "/opt/libexec/kubernetes/kubelet-plugins/volume/exec/"
---
apiVersion: kubeadm.k8s.io/v1beta3
kind: ClusterConfiguration
controllerManager:
  extraArgs:
    flex-volume-plugin-dir: "/opt/libexec/kubernetes/kubelet-plugins/volume/exec/"
networking:
  podSubnet: 10.244.0.0/16,2001:db8:42:0::/56
  serviceSubnet: 10.96.0.0/16,2001:db8:42:1::/112
# ---
# apiVersion: apps.sealos.io/v1beta1
# kind: Config
# metadata:
#   name: cilium
# spec:
#   path: charts/cilium/values.yaml
#   strategy: merge
#   data: |
#     ipv6:
#       enabled: true
---
apiVersion: apps.sealos.io/v1beta1
kind: Config
metadata:
  name: nfs-subdir-external-provisioner
spec:
  path: charts/nfs-subdir-external-provisioner/values.yaml
  strategy: merge
  data: |
    nfs:
      server: 192.168.56.117
      path: /data/nfs
    storageClass:
      name: managed-nfs-storage
```

## 镜像顺序

```yaml
image:
  - labring/kubernetes:v1.26.11-coreos
  - labring/cilium:v1.13.4-coreos
  - labring/metrics-server:v0.6.4
  - labring/nfs-subdir-external-provisioner:v4.0.18
```

sealos 会按顺序调用镜像中的部署脚本，所以头 2 个镜像的顺序不能调换（先部署 k8s 集群，然后部署网络插件）。

## `/usr` 只读挂载处理

```yaml
......
    volume-plugin-dir: "/opt/libexec/kubernetes/kubelet-plugins/volume/exec/"
......
    flex-volume-plugin-dir: "/opt/libexec/kubernetes/kubelet-plugins/volume/exec/"
......
```

kubeadm 默认参数是在 `/usr/libexec` 下面，`/usr/libexec` 在 CoreOS 上是只读的，这里改为 `/opt/libexec`。

> 参考 https://kubernetes.io/zh-cn/docs/setup/production-environment/tools/kubeadm/troubleshooting-kubeadm/#usr-mounted-read-only

## 设置 CIDR

```yaml
......
networking:
  podSubnet: 10.244.0.0/16,2001:db8:42:0::/56
  serviceSubnet: 10.96.0.0/16,2001:db8:42:1::/112
......
```

参考 https://kubernetes.io/zh-cn/docs/setup/production-environment/tools/kubeadm/dual-stack-support/#create-a-dual-stack-cluster

## 修改 chart values 文件

sealos 集群镜像目前已经适配了很多应用，包括 metrics-server、nfs……这里通过修改 chart values.yaml 来做一些配置。

> 仓库地址 https://github.com/labring-actions/cluster-image

```yaml
apiVersion: apps.sealos.io/v1beta1
kind: Config
metadata:
  name: nfs-subdir-external-provisioner
spec:
  path: charts/nfs-subdir-external-provisioner/values.yaml
  strategy: merge
  data: |
    nfs:
      server: 192.168.56.117
      path: /data/nfs
    storageClass:
      name: managed-nfs-storage
```

## 查看集群状态

![](imgs/Snipaste_2023-12-12_15-12-05.png)

![](imgs/Snipaste_2023-12-12_15-13-06.png)

# 参考文档

- sealos k8s 相关文档：https://sealos.io/zh-Hans/docs/self-hosting/lifecycle-management/
- sealos rootfs：https://github.com/labring-actions/runtime
- sealos 集群镜像：https://github.com/labring-actions/cluster-image

# sealos 常用命令

## sealos images

打印本地镜像缓存，sealos 的本地镜像缓存在 `/var/lib/containers/storage`

## sealos build --debug -t \<tag\> \<Kubefile\>

修改 sealos 镜像

## sealos pull \<img\>

下载 sealos 镜像到本地缓存

## sealos save -o \<tar\> -m --format docker-archive \<imgs\>

导出镜像到 tar 包

## sealos load -i \<tar\>

导入镜像

## sealos rmi \<imgs\>

删除镜像

## sealos add

添加 k8s 节点

## sealos delete

删除 k8s 节点

## sealos reset

销毁集群
