# webrouter
webrouter


## build

webrouter depends on `godep`, you need install `godep` first.

1. install godep

`$ go get github.com/tools/godep`

1. clone project into GOPATH

`$ go get github.com/morya/webrouter`

1. use `godep` to restore all dependency


```
$ cd $GOPATH/src/github.com/morya/webrouter
$ godep restore
```

如果，如果失败了，类似如下

```
[websock@websocket webrouter]$ godep restore
godep: Dep (github.com/labstack/echo) restored, but was unable to load it with error:
       	Package (golang.org/x/net/context) not found
godep: Dep (github.com/labstack/gommon/log) restored, but was unable to load it with error:
       	Package (golang.org/x/sys/unix) not found
godep: Dep (github.com/mattn/go-colorable) restored, but was unable to load it with error:
       	Package (golang.org/x/sys/unix) not found
godep: Dep (github.com/mattn/go-isatty) restored, but was unable to load it with error:
       	Package (golang.org/x/sys/unix) not found
```

可以从其它台机器copy $GOPATH/src/golang.org 的相关内容过来到本机对应的GOPATH.

