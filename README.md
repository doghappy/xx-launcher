# xx-launcher

https://github.com/doghappy/xx-launcher

修真情缘服务端启动器。  

服务端分了 n 多个区，公司没有采用 CI/CD，做了这个工具，帮助服务端快速部署。 

## 如何工作？

此工具是一个服务端程序，接收客户端的请求，执行客户端需要的功能所对应的脚本，从而实现了以下功能：

- 开启服务
- 关闭服务
- 更新配置
- 更新服务
- 查询 dmp 文件数量

### 开启/关闭服务

调用服务器上事先放置好的 bat 脚本，脚本由他人维护的，此工具只管调用。

### 更新配置/服务

从远程 FTP 下载 rar 更新包（大陆、港澳台、海外的更新包可能不一样，需要在 FTP 配置节点下配置好 Path）并在服务器上保留更新包。存储在工作目录，解压覆盖旧文件。

## 如何配置？

配置文件使用 yml，需要符合 yml 的语法。

```yml
LauncherUrl: 127.0.0.1:8080
# 白名单，可配置多个。服务器最好在外层设置白名单
Whitelist:
  - 127.0.0.1
# 从配置的 FTP 上下载更新包
Ftp:
  Host: ftp://example.com
  User: launcher
  Password: B:6t8*e<hA+&]Xte
  Path: /GangAoTai
Regions:
  -
    # 区 id，一台服务器可能会部署多个区
    RegionId: 1
    WorkDir: D:\XiuZhenServer\release_unicode
    Start: a启动游戏.bat
    Stop: b关闭游戏.bat
```

## 日志

运行的日志会打印在控制台上，同时也会追加日志到此工具同级目录中。
