// Copyright 2014 Wandoujia Inc. All Rights Reserved.
// Licensed under the MIT (MIT-LICENSE.txt) license.

package main

import (
	"flag"
	"runtime"
	"strings"
	"time"

	clog "github.com/CodisLabs/codis/pkg/utils/log"
	"github.com/morya/webrouter/pkg/log"
	w "github.com/morya/webrouter/pkg/webrouter"
)

var (
	cpus                = flag.Int("cpus", 1, "use cpu nums")
	addr                = flag.String("addr", ":443", "listen addr.(e.g. ip:port)")
	longPoolingInterval = flag.Duration("long-polling", 15*time.Second, "long polling interval")
	netDelay            = flag.Duration("net-delay", 10*time.Second, "net transfer delay")
	backendSyncTimeout  = flag.Duration("backend-sync-timeout", 30*time.Second, "send msg to backend, how long to wait response.")
	backends            = flag.String("backends", "", "backends servers, use ',' splite")
	name                = flag.String("name", "", "server name")
	connectType         = flag.Int("connect-type", 3, "connect type")
	msgType             = flag.Int("msg-type", 0, "message type")
)

var (
	logFile  = flag.String("log-file", "", "which file to record log, if not set stdout to use.")
	logLevel = flag.String("log-level", "info", "log level.")
)

func main() {
	flag.Parse()

	log.InitLog(*logFile)
	log.SetLogLevel(*logLevel)

	clog.Infof("Conf: cpus<%d>", *cpus)
	clog.Infof("Conf: listen addr<%s>", *addr)
	clog.Infof("Conf: backends<%s>", *backends)
	clog.Infof("Conf: name<%s>", *name)
	clog.Infof("Conf: connect-type<%d>", *connectType)
	clog.Infof("Conf: msg-type<%d>", *msgType)

	runtime.GOMAXPROCS(*cpus)

	conf := &w.Conf{
		Addr:                *addr,
		LongpollingInterval: *longPoolingInterval,
		NetDelay:            *netDelay,
		BackendSyncTimeout:  *backendSyncTimeout,
		Backends:            strings.Split(*backends, ","),
		Name:                *name,
		ConnectType:         byte(*connectType),
		MsgType:             byte(*msgType),
	}

	server := w.NewServer(conf)
	server.Start()
}
