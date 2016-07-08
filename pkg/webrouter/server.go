package webrouter

import (
	"github.com/CodisLabs/codis/pkg/utils/log"
	"github.com/labstack/echo"
	sd "github.com/labstack/echo/engine/standard"
	mw "github.com/labstack/echo/middleware"
	ws "golang.org/x/net/websocket"
	"net/http"
	"strconv"
)

const (
	API_CONNECT      = "/connect"
	API_DIS_CONNECT  = "/disconnect"
	API_SEND         = "/send"
	API_WEBSOCKET    = "/ws"
	API_QUERY_STATUS = "/user_status"

	PARAM_SESSSION_ID = "sid"
	PARAM_VERSION_ID  = "vid"
	PARAM_CALLBACK    = "callback"
	PARAM_DATA        = "data"

	PARAM_USER_IDS = "user_ids"
)

type Server struct {
	e       *echo.Echo
	cnf     *Conf
	manager *LongpollingManager
}

func NewServer(cnf *Conf) *Server {
	s := &Server{
		e:       echo.New(),
		cnf:     cnf,
		manager: NewLongpollingManager(cnf),
	}

	s.init()

	return s
}

func (self *Server) init() {
	self.e.Use(mw.Recover())

	self.e.Get(API_WEBSOCKET, sd.WrapHandler(ws.Handler(self.websocket)))
	self.e.Get(API_CONNECT, self.connect())
	self.e.Get(API_DIS_CONNECT, self.disconnect())
	self.e.Get(API_SEND, self.send())
	self.e.Get(API_QUERY_STATUS, self.queryUserStatus())

	self.e.Static("/assets/js", "public/js")
	self.e.Static("/html", "public")

	log.Infof("%s Init complete.", MODULE_SERVER_FRONTEND)
}

func (self *Server) Start() {
	log.Infof("%s Started at<%s>", MODULE_SERVER_FRONTEND, self.cnf.Addr)
	self.e.Run(sd.New(self.cnf.Addr))
	log.Infof("%s Exit.", MODULE_SERVER_FRONTEND)
}

func (self *Server) queryUserStatus() echo.HandlerFunc {
	return func(c echo.Context) error {
		userIds := c.QueryParam(PARAM_USER_IDS)
		cb := c.QueryParam(PARAM_CALLBACK)
		msg, err := self.manager.QueryUserStatus(userIds)

		errMsg := ""
		var status interface{}

		if err != nil {
			log.InfoErrorf(err, "%s Query <%s> status failure.", MODULE_SERVER_FRONTEND, userIds)
			errMsg = err.Error()
		} else {
			status = msg["userstatus"]
		}

		if cb != "" {
			return c.JSONP(http.StatusOK, cb, map[string]interface{}{
				"err":    errMsg,
				"status": status,
			})
		} else {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"err":    errMsg,
				"status": status,
			})
		}

	}
}

func (self *Server) connect() echo.HandlerFunc {
	return func(c echo.Context) error {
		sid := c.QueryParam(PARAM_SESSSION_ID)
		vid := c.QueryParam(PARAM_VERSION_ID)
		cb := c.QueryParam(PARAM_CALLBACK)

		log.Debugf("%s Connect Sid<%s>, Cid<%s>, Cb<%s>", MODULE_SERVER_FRONTEND, sid, vid, cb)

		var longpolling *Longpolling

		if sid == "" {
			// new connect client
			longpolling = self.manager.NewLongpolling()
			longpolling.ResetTimeout()

			log.Infof("%s Accept a new longpolling. %s", MODULE_SERVER_FRONTEND, longpolling.ToString())
		} else {
			// re connect client
			longpolling = self.getLongpolling(sid, vid)

			if longpolling.Err == ERR_NONE {
				longpolling.ResetTimeout()
				log.Infof("%s Reconnect. Sid<%s>", MODULE_SERVER_FRONTEND, longpolling.Sid)
			}
		}

		if sid != "" && nil != longpolling && longpolling.Err == ERR_NONE {
			longpolling.WaitResponse()
		}

		return c.JSONP(http.StatusOK, cb, longpolling.GetData())
	}
}

func (self *Server) disconnect() echo.HandlerFunc {
	return func(c echo.Context) error {
		return nil
	}
}

func (self *Server) send() echo.HandlerFunc {
	return func(c echo.Context) error {
		psid := c.QueryParam(PARAM_SESSSION_ID)
		cb := c.QueryParam(PARAM_CALLBACK)
		data := c.QueryParam(PARAM_DATA)

		log.Debugf("%s Send Sid<%s>, Data<%s>", MODULE_SERVER_FRONTEND, psid, data)

		var sid int64

		if "" != psid {
			sid, _ = strconv.ParseInt(psid, 10, 64)
		}

		longpolling := self.manager.GetLongpolling(sid)

		if longpolling == nil {
			longpolling = NewErrEmptyLongpolloing(ERR_KNOWN_SID)
		}

		if longpolling.Err == ERR_NONE {
			self.manager.SendToBackend(sid, newData(longpolling.Sid, longpolling.Vid, data))
		}

		return c.JSONP(http.StatusOK, cb, newData(longpolling.Sid, longpolling.Vid, ""))
	}
}

func (self *Server) getLongpolling(psid string, vid string) *Longpolling {
	var sid int64

	if "" != psid {
		sid, _ = strconv.ParseInt(psid, 10, 64)
	}

	longpolling := self.manager.GetLongpolling(sid)

	parsedVid, err := strconv.Atoi(vid)

	if longpolling == nil {
		longpolling = NewErrEmptyLongpolloing(ERR_KNOWN_SID)
		log.Infof("%s Unknow longpolling. Sid<%s>", MODULE_SERVER_FRONTEND, sid)
	} else if err != nil || longpolling.Vid != parsedVid {
		longpolling = NewErrEmptyLongpolloing(ERR_KNOWN_VID)
		log.Infof("%s Unknow longpolling. Sid<%s>, Vid<%s>, ExpectVid<%d>", MODULE_SERVER_FRONTEND, sid, vid, longpolling.Vid)
	} else {
		longpolling.Err = ERR_NONE
	}

	return longpolling
}
