# k8s教程说明
- [k8s底层原理和源码讲解之精华篇](https://ke.qq.com/course/4093533)
- [k8s底层原理和源码讲解之进阶篇](https://ke.qq.com/course/4236389)
- [k8s纯源码解读课程，助力你变成k8s专家](https://ke.qq.com/course/4697341)
- [k8s-operator和crd实战开发 助你成为k8s专家](https://ke.qq.com/course/5458555)
- [tekton全流水线实战和pipeline运行原理源码解读](https://ke.qq.com/course/5467720)

# prometheus全组件的教程
- [01_prometheus全组件配置使用、底层原理解析、高可用实战](https://ke.qq.com/course/3549215)
- [02_prometheus-thanos使用和源码解读](https://ke.qq.com/course/3883439)
- [03_kube-prometheus和prometheus-operator实战和原理介绍](https://ke.qq.com/course/3912017)
- [04_prometheus源码讲解和二次开发](https://ke.qq.com/course/4236995)

# go语言课程
- [golang基础课程](https://ke.qq.com/course/4334898)
- [golang运维平台实战，服务树,日志监控，任务执行，分布式探测](https://ke.qq.com/course/4334675)

# 关于白嫖和付费
- 白嫖当然没关系，我已经贡献了很多文章和开源项目，当然还有免费的视频
- 但是客观的讲，如果你能力超强是可以一直白嫖的，可以看源码。什么问题都可以解决
- 看似免费的资料很多，但大部分都是边角料，核心的东西不会免费，更不会有大神给你答疑
- thanos和kube-prometheus如果你对prometheus源码把控很好的话，再加上k8s知识的话就觉得不难了

# 更新说明
- 这个stree-index之前写的比较挫
- 比如node_path的自身id拼接
- 比如全局变量没有使用接口容器
- 比如没有mock公有云更新的方法
- 比如没有考虑物理机agent rpc上报信息的case
- 比如倒排索引模块没有抽取出来
# 等等这些 处理的都不算好，现在有改进版的的大项目 open-devops 
- 更有课程放送，让你也学会写go项目
![image](https://github.com/ning1875/stree-index/blob/main/images/open-devops.png)

# 开源项目地址: 
项目地址: [https://github.com/ning1875/stree-index](https://github.com/ning1875/stree-index)

# 什么是服务树及其核心功能
**服务树效果图**
可以看我之前写的文章 [服务树系列\(一\)：什么是服务树及其核心功能](https://segmentfault.com/a/1190000023761706)
![image](https://github.com/ning1875/stree-index/blob/main/images/tree_node.png)
![image](https://github.com/ning1875/stree-index/blob/main/images/tree_mon.png)
## 核心功能有三个
- 树级结构
- 灵活的资源查询
- 权限相关
今天仅讨论前两种的实现
# 树级结构实现
## 调研后发现有下列几种实现方式
- 左右值编码
- 区间嵌套
- 闭包表
- 物化路径
而`stree-index`采用的是物化路径
## 物化路径
### 原理
在创建节点时，将节点的完整路径进行记录，方案借助了unix文件目录的思想，主要时以空间换时间
```
+-----+-------+------------+-------------------------+
| id  | level | path       | node_name               |
+-----+-------+------------+-------------------------+
|  84 |     3 | /1/18/84   | ads-schedule            |
| 213 |     3 | /1/212/213 | share                   |
| 317 |     3 | /1/212/317 | ssr                     |
| 320 |     3 | /1/212/320 | prod                    |
| 475 |     3 | /1/212/475 | share-server-plus       |
| 366 |     3 | /1/365/366 | minivideo               |
| 368 |     3 | /1/365/368 | userinfo                |
+-----+-------+------------+-------------------------+
```
### 实现说明
- 接口代码在 `E:\go_path\src\stree-index\pkg\web\controller\node-path\path_controller.go`中
- level 字段代表层级 eg: a.b.c 对应的level分别为2.3.4
- path 路径  path最后的字段为其id值，上一级为其父节点id值
- node_name 叶子节点name
```
+-----------+--------------+------+-----+---------+----------------+
| Field     | Type         | Null | Key | Default | Extra          |
+-----------+--------------+------+-----+---------+----------------+
| id        | int(11)      | NO   | PRI | NULL    | auto_increment |
| level     | tinyint(4)   | NO   | MUL | NULL    |                |
| path      | varchar(200) | YES  |     | NULL    |                |
| node_name | varchar(200) | YES  |     | NULL    |                |
+-----------+--------------+------+-----+---------+----------------+

```
#### 增加节点
- 只需要判断传入的g.p.a各层级是否存在并添加即可
- 为了避免过多查询，我这里使用getall后变更。TODO 改为事务型
![image.png](/img/bVcHcD7)
#### 删除节点 大体同增加
#### 查询节点
需要分两种情况
- 查询下一级子节点 即 根据获取g下gp列表
```python
def get_gp():
    """

    :return:  返回请求的子节点
    t = ['hawkeye', 'stree', 'gateway']
    """
    data = {"node": "sgt", "level": 2, "max_dep": 1}
    tree_uri = '{}/query/node-path'.format(base_url)
    res = requests.get(tree_uri, params=data)
    print(res.json())
```
- 查询所有子孙节点
```python
def get_node():
    """
    :return:  返回请求的子节点
    t = ['sgt.hawkeye.m3coordinator',
         'sgt.hawkeye.etcd',
         'sgt.hawkeye.collecter',
         'sgt.hawkeye.rule-eval',
         'sgt.hawkeye.query',
         'sgt.hawkeye.m3db',
         'sgt.stree.index',
         ]
    """
    data = {"node": "sgt", "level": 2}
    tree_uri = '{}/node-path'.format(base_url)
    res = requests.get(tree_uri, params=data)
    print(res.json())

```

# 核心查询
## 如何满足灵活且高效的查询 
### 常规思路拼 sql查询
比如:查询 G.P.A=a.hawkeye.etcd的ecs资源
拼出的sql类似 
```sql
select * from  ecs where group="a" and product="hawkeye" and app="etcd"
```
同时mysql也满足下面5中查询条件
```
- eq 等于            : key=value 
- not_eq 不等于      : key!=value
- reg 正则匹配       : key=~value
- not_reg 正则非匹配 : key!~value
- 对比               : key> value
```
**弊端**
- 需要拼接的sql中每个条件是table 的字段
- 当然经常变动的字段可以存json字段使用 ` tags->'$."stree-project"'='hawkeye'`
- 性能问题：单表字段过多导致的查询问题
- 不能直接给出分布情况，只能再叠加count
### 更好的方法是使用倒排索引缓存记录
#### 什么是倒排索引
简单来说根据id查询记录叫索引，反过来根据tag匹配查找id就是倒排索引
#### 具体实现

- 核心结构MemPostings是一个双层map ，把tag=value 的记录按照tag作为第一层map的key， value作为内层map的key 内存map值为对应id set
```golang
type MemPostings struct {
   mtx     sync.RWMutex
   m       map[string]map[string][]uint64
   ordered bool
}
```
- 同时为了`反向匹配`和`统计需求`需要维护 values 和symbols
```golang
type HeadIndexReader struct {
   postings *index.MemPostings
   values   map[string]stringset
   symbols  map[string]struct{}
   symMtx   sync.RWMutex
}
```
- 将db的记录每条记录按照tag和id的对应关系构建索引
- 最内层set存储的是db 记录的主键id
- 这样就能够根据一组标签查询到主键id再去db中获取全量信息即可
- 这样查询速度是最快的
- db中所有的字段出timestamp外都可以用来构建索引，而后能所有的字段都可以被用作查询条件
举例
```python
req_data = {
    'resource_type': 'elb',
    'use_index': True,
    'labels': [
        # 查询 group 不等于CBS，name ，正则匹配.*0dff.*，stree-app等于collecter的elb资源列表
        {'key': 'group', 'value': 'CBS', 'type': 2},
        {'key': 'name', 'value': '.*0dff.*', 'type': 3},
        {'key': 'stree-app', 'value': 'collecter', 'type': 1}]
}
```
### 按key查询分布情况的实现
![image.png](/img/bVcHcIS)

- 匹配过程和上述一致
- 再用构建一个堆就可以得到分布情况
举例：根据kv组合查询 某一个key的分布情况
	eg:  查询 G.P.A=SGT.hawkeye.m3db 的ecs资源按cluster标签分布情况
```python
def query_dis():
    """

    :return:
    返回的是条件查询后按照目标label的分布情况
    dis = {
        'group': [
            {'name': 'business', 'value': 9},
            {'name': 'inf', 'value': 9},
            {'name': 'middleware', 'value': 9},
            {'name': 'bigdata', 'value': 9}
        ]
    }
    """


    req_data = {
        'resource_type': 'ecs',
        'use_index': True,
        'labels': [
            # 查询 G.P.A=SGT.hawkeye.m3db 的ecs资源按cluster标签分布情况
            {'key': 'group', 'value': 'SGT', 'type': 1},
            {'key': 'stree-project', 'value': 'hawkeye', 'type': 1},
            {'key': 'stree-app', 'value': 'm3db', 'type': 1}],
        'target_label': 'cluster'
    }
    query_uri = "{}/query/resource-distribution".format(base_url)
    res = requests.post(query_uri, json=req_data)
    print(res.json())

```

# 使用
## step 1 准备工作
> 准备mysql和redis，并修改配置文件对应字段
## 创建表

> 根据scripts/db_schema.sql 建表

> 资源数据表 
- ecs 云服务器
- elb 云负载均衡器
- rds 云关系型数据库
- dcs 云缓存
- 对应表名为
`service_tree_ecs` `service_tree_elb` `service_tree_rds` `service_tree_dcs`

> ecs 云服务器规格表  `service_tree_cloud_instance_type`
> 树结构path表 
>
## 灌入数据
- 资源数据可以由同步得来，自行实现即可
- 各个资源表中数据tags字段为json类型，切必须包含stree-index.yml的服务树tag
 
```yaml
# g.p.a模型key对应table中json字段名称
tree_info:
 name_g: group
  name_p: stree-project
  name_a: stree-app
```

![image.png](/img/bVcHcKn)

- ecs 云服务器规格表 可以由`scripts/instance_type_insert.sh`灌入，其中包含华为和aws的大部分规格数据

## step 2 编译或下载
> 直接下载
```shell script
wget https://github.com/ning1875/stree-index/releases/download/v1.0/stree-index-1.0.linux-amd64.tar.gz
```
> 自行编译

```shell script
git clone https://github.com/ning1875/stree-index.git
cd  stree-index && make 
```
## step 3 补充信息
> 补充stree-index.yml 中db，redis等信息

## step 4 启动服务
> 直接启动
```shell script
./stree-index --config.file=stree-index.yml
```
> 使用systemd启动
```shell script
/bin/cp -f stree-index.service /etc/systemd/system/
/bin/cp -f stree-index /bin/
mkdir -pv /etc/stree-index 
/bin/cp -f stree-index.yml /etc/stree-index/
systemctl enable /etc/systemd/system/stree-index.service
systemctl start stree-index.service

```
> 观察日志
```shell script
tail -f /var/log/messages |grep stree-index
```
## step 5 stree-index会自动根据资源表中服务树tag构建服务树
> 查询path表应该有数据
 
```shell script
select * from service_tree_path_tree limit 20;
```

# 查询
## 使用stree-index
### 查询接口数据结构
```
{
    "resource_type":"ecs",
    "use_index":true,
    "labels":[
        {
            "key":"group",
            "value":"sgt",
            "type":1
        }],
	'target_label': 'cluster' # 查询分布时才需要
}
```
### 查询资源返回数据结构
```python
 {
        'code': 0,
        'current_page': 2, # 当前分页
        'page_size': 10,   # 每页limit
        'page_count': 0,   # 页数
        'total_count': 0,  # 总数
        'result': [] # 为符合查询条件的资源列表
    }
```

### 查询数据分页参数，传在path里面
- page_size 代表每页数量，默认10
- current_page  代表当前分页num，默认1
- **get_all 等于1时代表不分页获取符合条件的全部数据，默认0**

### 查询资源类型:对应字段 resource_type

**目前支持的类型如下**
- ecs 云服务器
- elb 云负载均衡器
- rds 云关系型数据库
- dcs 云缓存

### 查询条件支持:对应字段 labels中的type字段
- 1：eq 等于            : key=value 
- 2：not_eq 不等于      : key!=value
- 3：reg 正则匹配       : key=~value
- 4：not_reg 正则非匹配 : key!~value
~~- 对比               : key> value~~ (暂时没支持)

### 查询条件自由组合
- `labels`可传入多个key和value组合，可自由组合不同kv查询

## 运维stree-index
### 监控
![image.png](/img/bVcHcLD)

stree-index 会打点，使用prometheus采集查看即可
```yaml
  - job_name: stree
    honor_labels: true
    honor_timestamps: true
    scrape_interval: 60s
    scrape_timeout: 4s
    metrics_path: /metrics
    scheme: http
    static_configs:
    - targets:
      - $stree-index:9393
```
