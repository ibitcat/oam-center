-- ----------------------------
-- Records of kgo_menu
-- ----------------------------
INSERT INTO `kgo_menu` VALUES (1, 0, '所有权限', '', 1, '', 1, 1, 1, 1, 1, 1, 1505620970, 0);
INSERT INTO `kgo_menu` VALUES (2, 1, '权限管理', '/', 999, 'fa-id-card', 1, 1, 1, 1, 1, 1, 1505622360, 0);
INSERT INTO `kgo_menu` VALUES (3, 2, '用户管理', '/home/user', 1, 'fa-user-o', 1, 1, 0, 0, 1, 1, 1528385411, 0);
INSERT INTO `kgo_menu` VALUES (4, 2, '角色管理', '/home/role', 2, 'fa-user-circle-o', 1, 1, 0, 1, 1, 1, 1505621852, 0);
INSERT INTO `kgo_menu` VALUES (5, 2, '菜单管理', '/home/menu', 3, 'fa-list', 1, 1, 1, 1, 1, 1, 1505621986, 0);
INSERT INTO `kgo_menu` VALUES (6, 1, '运维管理', '/oam', 1, 'fa-tasks', 1, 1, 1, 1, 1, 1, 0, 0);
INSERT INTO `kgo_menu` VALUES (7, 6, 'vps列表', '/home/vps', 1, 'fa-cloud', 1, 1, 1, 1, 1, 1, 0, 0);
INSERT INTO `kgo_menu` VALUES (8, 1, '个人中心', '/personal', 999, 'fa-user-circle-o', 1, 1, 1, 1, 1, 1, 1547000410, 0);
INSERT INTO `kgo_menu` VALUES (9, 6, '游服管理', '/home/game', 3, 'fa-server', 1, 1, 1, 1, 1, 1, 1546920455, 0);
INSERT INTO `kgo_menu` VALUES (10, 6, '合服管理', '/home/merge', 999, 'fa-compress', 0, 1, 1, 1, 1, 1, 1546920606, 0);
INSERT INTO `kgo_menu` VALUES (11, 6, 'cdn管理', '/home/cdn', 4, 'fa-file-code-o', 1, 1, 1, 1, 1, 1, 1546921447, 0);
INSERT INTO `kgo_menu` VALUES (12, 8, '资料修改', '/home/personal', 1, 'fa-edit', 1, 1, 1, 1, 1, 1, 1547000565, 1548147829);
INSERT INTO `kgo_menu` VALUES (13, 6, '平台列表', '/home/agent', 2, 'fa-map-signs', 1, 1, 1, 1, 1, 1, 1547724451, 0);
INSERT INTO `kgo_menu` VALUES (14, 6, '版本管理', '/home/version', 5, 'fa-file-archive-o', 1, 1, 1, 1, 1, 1, 1547785510, 0);
INSERT INTO `kgo_menu` VALUES (15, 6, 'nginx管理', '/home/nginx', 6, 'fa-sitemap', 1, 1, 1, 1, 1, 1, 1547785510, 0);


-- ----------------------------
-- Records of kgo_role
-- ----------------------------
INSERT INTO `kgo_role` VALUES (1, '超级管理员', '超级管理员，具有所有权限', 1, 1, 1, 0, 1548147334);
INSERT INTO `kgo_role` VALUES (2, '普通管理员', '普通管理员，无菜单管理权限', 1, 1, 1, 0, 1548145786);


-- ----------------------------
-- Records of kgo_user
-- ----------------------------
INSERT INTO `kgo_user` VALUES (1, 1, 'admin', 'admin', '4e391f7a8670cce46557dee39e734594', '13866668888', 'xxx@qq.com', '58QM', 1548669685, '192.168.2.34', 1, 0, 0, 0, 1548146771);
