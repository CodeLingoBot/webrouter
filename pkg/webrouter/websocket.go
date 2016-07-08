package webrouter

import (
	"github.com/CodisLabs/codis/pkg/utils/log"
	ws "golang.org/x/net/websocket"
	"time"
)

func (self *Server) websocket(conn *ws.Conn) {
	longpolling := self.manager.NewWebSocketLongpolling(conn)

	log.Info(formatSession(longpolling.Sid, MODULE_SERVER_FRONTEND, "Created: <%s>, %s", conn.Request().RemoteAddr, longpolling.ToString()))

	defer func(conn *ws.Conn) {
		self.manager.DeleteWebSocketLongpolling(conn)
		conn.Close()
		log.Info(formatSession(longpolling.Sid, MODULE_SERVER_FRONTEND, "Closed"))
	}(conn)

	err := self.manager.SendToFrondend(longpolling.Sid, "")
	if nil != err {
		log.InfoErrorf(err, formatSession(longpolling.Sid, MODULE_SERVER_FRONTEND, "Send failure"))
		return
	}

	// wait for client message
	for {
		err := conn.SetDeadline(time.Now().Add(self.cnf.LongpollingInterval + self.cnf.NetDelay*2))
		if err != nil {
			log.InfoErrorf(err, formatSession(longpolling.Sid, MODULE_SERVER_FRONTEND, "Set deadline failure"))
			return
		}

		msg := ""
		err = ws.Message.Receive(conn, &msg)
		if err != nil {
			log.InfoErrorf(err, formatSession(longpolling.Sid, MODULE_SERVER_FRONTEND, "Receive message failure"))
			return
		}

		log.Info(formatSession(longpolling.Sid, MODULE_SERVER_FRONTEND, "Receive a message <%s>", msg))

		data, err := UnMarshalData([]byte(msg))
		if nil != err {
			log.InfoErrorf(err, formatSession(longpolling.Sid, MODULE_SERVER_FRONTEND, "UnMarshal message <%s> failure", msg))
			return
		}

		if data.Sid != longpolling.Sid {
			log.InfoErrorf(err, formatSession(longpolling.Sid, MODULE_SERVER_FRONTEND, "Invalid sid, expect is <%s>, acture is <%s>", longpolling.Sid, data.Sid))
			return
		}

		if len(data.Value) == 0 {
			continue
		}

		err = self.manager.SendToBackend(longpolling.Sid, data)

		if nil != err {
			log.InfoErrorf(err, formatSession(longpolling.Sid, MODULE_SERVER_FRONTEND, "Send failure"))
			return
		}
	}
}
