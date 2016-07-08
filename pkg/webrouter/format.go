package webrouter

import (
	"fmt"
)

const (
	MODULE_CONF            = "Conf: "
	MODULE_SERVER_FRONTEND = "Frontend: "
	MODULE_SERVER_BACKEND  = "Backend: "
	MODULE_SERVER_MGR      = "Manager: "
	MODULE_SERVER_NET      = "Net: "
)

func formatSession(sid int64, module string, format string, args ...interface{}) string {
	newFormat := fmt.Sprintf("%s Session<%d> -- %s", module, sid, format)
	return fmt.Sprintf(newFormat, args...)
}
