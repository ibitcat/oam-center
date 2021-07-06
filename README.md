# 游戏服务器运维后台

## 安装
推荐使用 `golang 1.12` 及以上的版本编译。

编译命令：`go build -v -x` 或者 `go build -ldflags "-w -s"`(生成的可执行文件小一点，因为删除了一些调试信息)


## 初始化
可以使用我的 [mysqldiff](https://github.com/shuimu98/mysqldiff/releases) 来初始化数据库，命令如下：

```
mysqldiff -u root -p 123456 -d kgo_admin -f kgo_admin.sql
```
当然，你也可以手动初始化。

数据库创建后，需要初始化菜单表以及添加管理员，导入 `kgo_amdin_init.sql` 到 mysql 中即可。

```
mysql -uroot -p123456 kgo_admin < kgo_admin_init.sql
```

## 运行

```
nohup ./oam-center > runtime.log &
```

## 参考
本项目前端参考[该项目](https://github.com/george518/PPGo_ApiAdmin)。