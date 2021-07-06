CREATE DATABASE If Not Exists kgo_admin DEFAULT CHARSET utf8 COLLATE utf8_general_ci;

CREATE TABLE `kgo_vps` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT '流水号id',
  `ip` varchar(32) NOT NULL COMMENT 'ip地址',
  `type` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '类型（1=game,2=cdn,4=nginx,8=center）',
  `create_time` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '注册时间',
  `domain` varchar(128) NOT NULL COMMENT '被控端域名',
  `detail` varchar(32) NOT NULL COMMENT '备注',
  `status` tinyint(1) unsigned NOT NULL DEFAULT '0' COMMENT '状态0=离线，1=在线',
  `version` varchar(64) NOT NULL DEFAULT '' COMMENT 'oam-ctl 版本',
  `vps_time` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '当前时间',
  UNIQUE KEY `ip` (`ip`),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='云服务器列表';

CREATE TABLE `kgo_agent` (
  `aid` int(10) unsigned NOT NULL COMMENT '平台id(对应conf的plat)',
  `flag` varchar(32) NOT NULL COMMENT '平台标示, eg.: dalan',
  `name` varchar(32) NOT NULL COMMENT '平台名称, eg.: 大蓝',
  `lang` varchar(32) NOT NULL COMMENT '平台语言',
  `miniapp` tinyint(1) unsigned NOT NULL DEFAULT '0' COMMENT '是否是小程序（1=小程序）',
  `audit_version` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '当前提审版本',
  `last_audit` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '最后提审通过的版本',
  `vpsid` int(10) unsigned NOT NULL DEFAULT '0' COMMENT 'cdn所在vps',
  `source` varchar(64) NOT NULL DEFAULT '' COMMENT 'cdn回源域名',
  `domain` varchar(64) NOT NULL DEFAULT '' COMMENT 'cdn正式域名',
  `create_time` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `status` tinyint(1) unsigned NOT NULL DEFAULT '1' COMMENT '状态1-正常，0-删除',
  UNIQUE KEY `flag` (`flag`),
  PRIMARY KEY (`aid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='平台列表';

CREATE TABLE `kgo_nginx` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT '流水号id',
  `aid` int(10) unsigned NOT NULL COMMENT '平台id',
  `sid` int(10) unsigned NOT NULL COMMENT '起始sid',
  `vpsid` int(10) unsigned NOT NULL DEFAULT '0' COMMENT 'nginx所在的vps',
  `domain` varchar(64) NOT NULL COMMENT '转发域名',
  `ws` varchar(64) NOT NULL COMMENT 'websocket 转发域名',
  `single` varchar(64) NOT NULL COMMENT 'single 转发域名',
  `create_time` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '添加时间',
  `status` tinyint(1) unsigned NOT NULL DEFAULT '1' COMMENT '状态1-正常，0-删除',
  UNIQUE KEY `idx` (`aid`,`sid`),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='平台nginx转发服务器';

CREATE TABLE `kgo_game` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT '流水号id',
  `aid` int(10) unsigned NOT NULL COMMENT '平台id',
  `sid` int(10) unsigned NOT NULL COMMENT '服务器id（大于10000表示外测服）',
  `serial` int(8) unsigned NOT NULL COMMENT '游服唯一id，对应conf的serial',
  `gid` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '合服后的组id, 0=未被合服',
  `mid` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '数据源服务器ID(0=无导入)',
  `version` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '当前服务器版本',
  `name` varchar(32) NOT NULL COMMENT '服务器名称(例如：大鹏展翅)',
  `vpsid` int(10) unsigned NOT NULL COMMENT '游服所在的vps',
  `create_time` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `install_time` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '安装时间',
  `update_time` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '更新时间',
  `port` smallint(5) unsigned NOT NULL DEFAULT '0' COMMENT '游服端口',
  `db_port` smallint(5) unsigned NOT NULL DEFAULT '0' COMMENT '游服mysql端口',
  `db_share` smallint(5) unsigned NOT NULL DEFAULT '0' COMMENT '共享数据库的id',
  `open_time` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '开服时间',
  `merge_time` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '合服时间',
  `is_tls` tinyint(1) unsigned NOT NULL DEFAULT '0' COMMENT '是否走tls',
  `domain` varchar(255) NOT NULL COMMENT '单服域名(eg.: s1-andou-jmxy.kgogame.com)',
  `procs` varchar(255) NOT NULL DEFAULT '' COMMENT '进程信息',
  `status` tinyint(1) unsigned NOT NULL DEFAULT '0' COMMENT '状态 0=未安装 1=已安装 2=运行中 3=已暂停',
  `nginx_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT 'nginx转发服务器id',
  `ws` varchar(255) NOT NULL DEFAULT '' COMMENT '游服websocket入口地址（支持域名和ip:port）',
  `single` varchar(255) NOT NULL DEFAULT '' COMMENT '游服single入口地址',
  `mode` tinyint(1) unsigned NOT NULL DEFAULT '0' COMMENT '客户端显示模式 1=新服 2=火爆 3=维护',
  `install_log` text NOT NULL COMMENT '安装日志',
  `hoted` text NOT NULL COMMENT '已安装的补丁',
  `hot_log` text NOT NULL COMMENT '安装补丁日志',
  `start_log` text NOT NULL COMMENT '启服日志',
  `stop_log` text NOT NULL COMMENT '关服日志',
  `cid` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '跨服组id',
  `join_time` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '加入跨服的时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `server` (`aid`,`sid`),
  UNIQUE KEY `serial` (`serial`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='服务器列表';

CREATE TABLE `kgo_cross` (
  `id` int(10) unsigned NOT NULL COMMENT '流水号id',
  `plat` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '平台集合id(目前认为是aid，后期有需要再改)',
  `version` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '当前版本',
  `vpsid` int(10) unsigned NOT NULL COMMENT '所在的vps',
  `port` smallint(5) unsigned NOT NULL DEFAULT '0' COMMENT '跨服服端口',
  `db_port` smallint(5) unsigned NOT NULL DEFAULT '0' COMMENT '游服mysql端口',
  `create_time` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `status` tinyint(1) unsigned NOT NULL DEFAULT '0' COMMENT '状态 0=未安装 1=已安装 2=运行中 3=已暂停',
  `procs` varchar(255) NOT NULL DEFAULT '' COMMENT '进程信息',
  `install_log` text NOT NULL COMMENT '安装日志',
  `hoted` text NOT NULL COMMENT '已安装的补丁',
  `hot_log` text NOT NULL COMMENT '安装补丁日志',
  `start_log` text NOT NULL COMMENT '启服日志',
  `stop_log` text NOT NULL COMMENT '关服日志',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='跨服列表';

CREATE TABLE `kgo_menu` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `pid` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '上级ID，0为顶级',
  `auth_name` varchar(64) NOT NULL DEFAULT '' COMMENT '权限名称',
  `auth_url` varchar(255) NOT NULL DEFAULT '' COMMENT 'URL地址',
  `sort` int(11) unsigned NOT NULL DEFAULT '999' COMMENT '排序，越小越前',
  `icon` varchar(255) NOT NULL,
  `is_show` tinyint(1) unsigned NOT NULL DEFAULT '0' COMMENT '是否显示，0-隐藏，1-显示',
  `auth_bit` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '菜单权限',
  `user_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '操作者ID',
  `create_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '创建者ID',
  `update_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '修改者ID',
  `status` tinyint(1) unsigned NOT NULL DEFAULT '1' COMMENT '状态，1-正常，0-删除',
  `create_time` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `update_time` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '更新时间',
  UNIQUE KEY `auth_url` (`auth_url`),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='菜单列表';

CREATE TABLE `kgo_role` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `role_name` varchar(32) NOT NULL DEFAULT '0' COMMENT '角色名称',
  `detail` varchar(255) NOT NULL DEFAULT '0' COMMENT '备注',
  `create_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '创建者ID',
  `update_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '修改这ID',
  `status` tinyint(1) unsigned NOT NULL DEFAULT '1' COMMENT '状态1-正常，0-删除',
  `create_time` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '添加时间',
  `update_time` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '修改时间',
  UNIQUE KEY `role_name` (`role_name`),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='角色表';

CREATE TABLE `kgo_user` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '用户id',
  `role_id` int(11) unsigned NOT NULL COMMENT '角色id',
  `login_name` varchar(20) NOT NULL DEFAULT '' COMMENT '用户名',
  `real_name` varchar(32) NOT NULL DEFAULT '0' COMMENT '真实姓名',
  `password` char(32) NOT NULL DEFAULT '' COMMENT '密码',
  `phone` varchar(20) NOT NULL DEFAULT '0' COMMENT '手机号码',
  `email` varchar(50) NOT NULL DEFAULT '' COMMENT '邮箱',
  `salt` char(10) NOT NULL DEFAULT '' COMMENT '密码盐',
  `last_login` int(11) NOT NULL DEFAULT '0' COMMENT '最后登录时间',
  `last_ip` char(15) NOT NULL DEFAULT '' COMMENT '最后登录IP',
  `status` tinyint(4) NOT NULL DEFAULT '0' COMMENT '状态，1-正常 0禁用',
  `create_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '创建者ID',
  `update_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '修改者ID',
  `create_time` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `update_time` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '修改时间',
  UNIQUE KEY `idx_user_name` (`login_name`),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='用户表';

CREATE TABLE `kgo_cdn` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '流水号id',
  `aid` int(10) unsigned NOT NULL COMMENT '平台id',
  `version` varchar(64) NOT NULL DEFAULT '' COMMENT 'cdn版本, eg.: cn_andou_20190101',
  `create_time` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `status` tinyint(1) unsigned NOT NULL DEFAULT '1' COMMENT '0=未安装,1=已安装',
  `install_log` text NOT NULL COMMENT 'CDN安装日志',
  UNIQUE KEY `version` (`version`),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='CDN列表';
