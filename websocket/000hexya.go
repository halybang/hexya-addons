package websocket

import (
	"net/http"

	"github.com/hexya-erp/hexya/src/controllers"
	"github.com/hexya-erp/hexya/src/server"
	"github.com/hexya-erp/hexya/src/tools/logging"
	//"github.com/olahol/melody"
	//"github.com/hexya-erp/hexya/src/actions"
	//"github.com/hexya-erp/hexya/src/models"
	//"github.com/hexya-erp/hexya/src/models/security"
	//"github.com/hexya-erp/hexya/src/models/types"
	//"github.com/hexya-erp/pool/h"
)

const (
	MODULE_NAME string = "websocket"
	SEQUENCE    uint8  = 150
	NAME        string = "Websocket"
	VERSION     string = "0.1"
	//CATEGORY    string = "Hidden"
	DESCRIPTION string = `
The IOT Module
==============================================
	`
	AUTHOR     string = "Dinh Duong Ha"
	MAINTAINER string = "Dinh Duong Ha"
	WEBSITE    string = "http://www.topdoo.com"
)

var log logging.Logger

func init() {
	log = logging.GetLogger("websocket")
	server.RegisterModule(&server.Module{
		Name:     MODULE_NAME,
		PreInit:  PreInit,
		PostInit: PostInit,
	})
	// Move to PreInit
	// initWebsocket()
}

func PreInit() {
	initWebsocket()
}

func PostInit() {
	ResetAllSession()
}

func initWebsocket() {

	/*
		{"id": 1, "jsonrpc": "2.0", "method": "login", "params": {"user": "admin", "password": "admin"}}
		{"id": 2, "jsonrpc": "2.0", "method": "login", "result": {"user": "admin", "password": "admin"}}
		{"id": 3, "jsonrpc": "2.0", "result": {"user": "admin", "password": "admin"}}
		{"id": 4, "jsonrpc": "2.0", "method": "login", "error": {"code": -333, "message": "Error message", "data": {"user": "admin", "password": "admin"}}}
		{"id": 5, "jsonrpc": "2.0", "method": "login", "error": {"code": -333, "message": "Error message 2"}}
		Implement JsonRPC using websocket

		/version
		/login
		/session/logout


		/locale
		/session/get_session_info
		/session/modules

		/translations
		/proxy/load

		/action/load
		/action/run
		/menu/load_needaction

		/dataset/call_kw/*path
		/dataset/search_read
		/dataset/call_button

	*/
	root := controllers.Registry
	jsonHexya, err := NewService("jsonrpc")
	if err == nil {
		jsonHexya.RegisterMethod("version", JsonRPCVersionInfo)
		jsonHexya.RegisterMethod("login", JsonRPCLogin)
		jsonHexya.RegisterMethod("logout", JsonRPCLogout)

		jsonHexya.RegisterMethod("locale", nil)
		jsonHexya.RegisterMethod("session", JsonRPCSessionInfo) // Session info, locale, modules, token
		jsonHexya.RegisterMethod("modules", nil)

		jsonHexya.RegisterMethod("translations", nil)
		jsonHexya.RegisterMethod("proxy_load", nil)

		jsonHexya.RegisterMethod("actionload", JsonRPCActionLoad)
		jsonHexya.RegisterMethod("actionrun", JsonRPCActionRun)
		jsonHexya.RegisterMethod("menu", JsonRPCMenuLoadNeedaction)

		jsonHexya.RegisterMethod("call_kw", JsonRPCCallKW)
		jsonHexya.RegisterMethod("search_read", JsonRPCSearchRead)
		jsonHexya.RegisterMethod("call_button", JsonRPCCallButton)

		jsonHexya.RegisterMethod("token", JsonRPCToken) // New Token

		jsonHexya.RegisterResponser("ping", JsonRPCHandleResponsePing)

		root.AddController(http.MethodGet, "/jsonrpc", MakeHandleFunc(jsonHexya))
	}
}
