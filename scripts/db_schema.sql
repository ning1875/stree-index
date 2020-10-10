-- ecs 资源表
CREATE TABLE `service_tree_ecs` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增id',
  `uid` varchar(100) NOT NULL COMMENT '实例id',
  `name` varchar(200) NOT NULL COMMENT '资源名称',
  `hash` varchar(100) NOT NULL COMMENT '哈希',
  `cloud_provider` varchar(20) NOT NULL COMMENT '云类型',
  `account_id` int(11) NOT NULL COMMENT '对应账户在account表中的id',
  `region` varchar(20) NOT NULL,
  `tags` json DEFAULT NULL COMMENT '标签',
  `instance_type` varchar(100) NOT NULL COMMENT '资产规格类型',
  `charging_mode` varchar(10) DEFAULT NULL COMMENT '付费类型',
  `private_ip` json DEFAULT NULL COMMENT '私有IP',
  `public_ip` json DEFAULT NULL COMMENT '公有IP',
  `availability_zone` varchar(20) NOT NULL COMMENT '可用区',
  `vpc_id` varchar(40) DEFAULT NULL COMMENT 'VPC ID',
  `subnet_id` varchar(40) DEFAULT NULL COMMENT '子网ID',
  `status` varchar(20) NOT NULL COMMENT '状态',
  `security_groups` json DEFAULT NULL COMMENT '安全组',
  `created_at` datetime NOT NULL COMMENT '启动时间',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `cmdb_server_instance_instance_id_uindex` (`uid`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4

-- elb 资源表
CREATE TABLE `service_tree_elb` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增id',
  `uid` varchar(100) NOT NULL COMMENT 'ELB ID,仅对huawei有效',
  `name` varchar(100) NOT NULL COMMENT '名称',
  `hash` varchar(100) NOT NULL COMMENT '哈希',
  `cloud_provider` varchar(20) NOT NULL COMMENT '云类型',
  `account_id` int(11) NOT NULL COMMENT '账户ID,和account表中id相关',
  `region` varchar(20) NOT NULL COMMENT '区域',
  `elb_type` varchar(20) DEFAULT NULL COMMENT '类型',
  `status` varchar(20) DEFAULT NULL COMMENT '状态',
  `ip_address` varchar(20) DEFAULT NULL COMMENT 'IP地址,仅对huawei有效',
  `dns_name` varchar(100) DEFAULT NULL COMMENT 'DNS名称,仅对aws有效',
  `port` json DEFAULT NULL,
  `backends` json DEFAULT NULL COMMENT '后端服务器',
  `target_group` json DEFAULT NULL,
  `tags` json DEFAULT NULL COMMENT '标签',
  `created_at` datetime DEFAULT '2006-01-02 15:04:05' COMMENT '创建时间',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `service_tree_elb_name_uid_region_uindex` (`name`,`uid`,`region`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4


-- rds 资源表
CREATE TABLE `service_tree_rds` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增 id',
  `uid` varchar(100) NOT NULL COMMENT '实例 id',
  `name` varchar(100) NOT NULL COMMENT '实例名称',
  `hash` varchar(100) NOT NULL COMMENT '哈希',
  `cloud_provider` varchar(15) NOT NULL COMMENT '云服务商',
  `account_id` int(11) NOT NULL COMMENT '账户 id',
  `region` varchar(50) NOT NULL COMMENT '实例所在区域',
  `cluster_id` int(11) DEFAULT NULL COMMENT '集群 id',
  `cluster_name` varchar(100) DEFAULT NULL COMMENT '集群名称',
  `master_id` varchar(100) DEFAULT NULL COMMENT 'master 节点 id',
  `status` varchar(20) NOT NULL COMMENT '实例状态',
  `private_ip` json DEFAULT NULL COMMENT '私有 IP 数组',
  `public_ip` json DEFAULT NULL COMMENT '公有 IP 数组',
  `port` int(4) NOT NULL COMMENT '实例端口',
  `architecture_type` varchar(30) DEFAULT NULL COMMENT '实例类型',
  `tags` json DEFAULT NULL COMMENT '实例标签对象',
  `engine` varchar(15) NOT NULL COMMENT '数据库引擎',
  `engine_version` varchar(30) NOT NULL COMMENT '数据库引擎版本',
  `vpc_id` varchar(150) DEFAULT NULL COMMENT 'VPC id',
  `subnet_id` varchar(150) DEFAULT NULL COMMENT '子网名称或 id',
  `security_groups` json DEFAULT NULL COMMENT '安全组 id 数组',
  `instance_type` varchar(50) DEFAULT NULL COMMENT '实例规格',
  `resource_id` varchar(200) DEFAULT '' COMMENT '资源 id',
  `is_writer` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否可写',
  `charging_mode` varchar(50) DEFAULT NULL COMMENT '付费类型',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `service_tree_rds_uid_uindex` (`uid`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4

-- dcs 资源表
CREATE TABLE `service_tree_dcs` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增id',
  `uid` varchar(100) NOT NULL COMMENT '实例id',
  `name` varchar(100) NOT NULL COMMENT '名称',
  `hash` varchar(100) NOT NULL COMMENT '哈希',
  `cloud_provider` varchar(10) NOT NULL COMMENT '云服务提供商',
  `account_id` int(11) NOT NULL COMMENT '账户id',
  `region` varchar(20) NOT NULL COMMENT '区域',
  `cluster_id` varchar(100) NOT NULL COMMENT '集群id',
  `instance_type` varchar(20) NOT NULL COMMENT '实例类型',
  `charging_mode` varchar(20) DEFAULT NULL COMMENT '付费方式',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `private_ip` json NOT NULL COMMENT '内网ip',
  `port` varchar(10) NOT NULL COMMENT '端口',
  `public_ip` json DEFAULT NULL COMMENT '公网ip',
  `status` varchar(20) NOT NULL COMMENT '状态',
  `engine` varchar(20) NOT NULL COMMENT '缓存引擎',
  `engine_version` varchar(20) NOT NULL COMMENT '引擎版本',
  `security_groups` json DEFAULT NULL COMMENT '安全组',
  `vpc_id` varchar(20) NOT NULL COMMENT 'vpc id',
  `tags` json DEFAULT NULL COMMENT '标签',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `service_tree_dcs_uid_uindex` (`uid`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4


-- ecs vm规格表
CREATE TABLE `service_tree_cloud_instance_type` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `resource_type` varchar(255) NOT NULL COMMENT '资源类型: ecs /rds/dcs',
  `instance_type` varchar(255) NOT NULL COMMENT 'ecs规格',
  `cloud_provider` varchar(255) NOT NULL DEFAULT '' COMMENT '公有云名称',
  `cpu` int(10) unsigned NOT NULL COMMENT 'vcpus',
  `mem` int(10) unsigned NOT NULL COMMENT '单位GB',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8


-- 树结构path表
CREATE TABLE `sgt_service_tree_path_tree` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` tinyint(4) NOT NULL,
  `path` varchar(200) DEFAULT NULL,
  `node_name` varchar(200) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_unique_key` (`level`,`path`,`node_name`) USING BTREE COMMENT '唯一索引'
) ENGINE=InnoDB  DEFAULT CHARSET=utf8