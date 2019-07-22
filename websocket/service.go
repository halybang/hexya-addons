package websocket

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	_ "net"
	_ "strconv"
	"sync"

	_ "github.com/gin-gonic/gin"
	"github.com/oklog/ulid"
	"github.com/olahol/melody"

	"github.com/hexya-erp/hexya/src/models"
	_ "github.com/hexya-erp/hexya/src/models/security"
	_ "github.com/hexya-erp/hexya/src/models/types"
	"github.com/hexya-erp/hexya/src/server"
	"github.com/hexya-erp/pool/h"
	_ "github.com/hexya-erp/pool/q"
)

// Handler links a method of JSON-RPC request.
//type JsonRPCHandler interface {
//	ServeJSONRPC(s *melody.Session, params *json.RawMessage) (result interface{}, err error)
//}

//type HexyaJsonRpcHandler struct {
//}

//func (h *HexyaJsonRpcHandler) ServeJSONRPC(s *melody.Session, params *json.RawMessage) (result interface{}, err error) {
//	return nil, nil
//}

const (
	// ErrorCodeParse is parse error code.
	ErrorCodeParse ErrorCode = -32700
	// ErrorCodeInvalidRequest is invalid request error code.
	ErrorCodeInvalidRequest ErrorCode = -32600
	// ErrorCodeMethodNotFound is method not found error code.
	ErrorCodeMethodNotFound ErrorCode = -32601
	// ErrorCodeInvalidParams is invalid params error code.
	ErrorCodeInvalidParams ErrorCode = -32602
	// ErrorCodeInternal is internal error code.
	ErrorCodeInternal ErrorCode = -32603
)

type ErrorCode int

// A RequestRPC is the message format expected from a client
type RequestRPC struct {
	JsonRPC string           `json:"jsonrpc"`
	ID      int64            `json:"id"`
	Method  string           `json:"method,omitempty"`
	Params  *json.RawMessage `json:"params,omitempty"`
	Result  *json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError    `json:"error,omitempty"`
}

// Result  interface{}      `json:"result,omitempty"`
// Error   *Error           `json:"error,omitempty"`
// A ResponseRPC is the message format sent back to a client
// in case of success
type ResponseRPC struct {
	JsonRPC string           `json:"jsonrpc"`
	ID      int64            `json:"id"`
	Method  string           `json:"method,omitempty"`
	Result  *json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError    `json:"error,omitempty"`
}

// A ResponseError is the message format sent back to a
// client in case of failure
type ResponseError struct {
	JsonRPC string       `json:"jsonrpc"`
	ID      int64        `json:"id"`
	Error   JSONRPCError `json:"error"`
}

// JSONRPCErrorData is the format of the Data field of an Error Response
type JSONRPCErrorData struct {
	Arguments string `json:"arguments"`
	Debug     string `json:"debug"`
}

// JSONRPCError is the format of an Error in a ResponseError
type JSONRPCError struct {
	Epoch   int64       `json:"epoch"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type JSONRPCGenericParams struct {
	Epoch int64 `json:"epoch"`
}

// Websocket own session
type Session struct {
	*melody.Session
	Service *Service
	Epoch   int64  `json:"epoch"` // Last
	UID     int64  `json:"uid"`
	SID     string `json:"sid"`
	ULID    string `json:"ulid"`
}

//type JsonRPCHandler func(c context.Context, params *json.RawMessage) (result interface{}, err *error)
//type JsonRPCHandler func(s *melody.Session, params *json.RawMessage) (result interface{}, err *error)

type JsonRPCHandleResponseFunc func(s *Session, response *ResponseRPC)

type JsonRPCHandleFunc func(s *Session, r *RequestRPC) (result interface{}, err error)

type HandleMessageFunc func(*Session, []byte)

type Service struct {
	*melody.Melody
	mutex     sync.RWMutex
	Name      string
	mw        []HandleMessageFunc
	mwb       []HandleMessageFunc
	methods   map[string]JsonRPCHandleFunc
	responses map[string]JsonRPCHandleResponseFunc
	Sessions  sync.Map
}

var Services sync.Map

// RegisterMethod registers jsonrpc.Func to MethodRepository.
func (s *Service) RegisterMethod(method string, h JsonRPCHandleFunc) error {
	if method == "" || h == nil {
		return errors.New("RegisterMethod: method name and function should not be empty")
	}
	_, ok := s.methods[method]
	if ok {
		return errors.New("RegisterMethod: Method is registered")
	}
	s.mutex.Lock()
	s.methods[method] = h
	s.mutex.Unlock()
	return nil
}

func (s *Service) RegisterResponser(method string, h JsonRPCHandleResponseFunc) error {
	if method == "" || h == nil {
		return errors.New("RegisterResponser: method name and function should not be empty")
	}
	_, ok := s.responses[method]
	if ok {
		return errors.New("RegisterResponser: Method is registered")
	}

	s.mutex.Lock()
	s.responses[method] = h
	s.mutex.Unlock()
	return nil
}

func (service *Service) Log(s *Session, request *RequestRPC, msg []byte) {

	/*
		models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			var peerTime int64
			var params JSONRPCGenericParams
			var rsp bool
			if request == nil {
				return
			}

			if request.Params != nil {
				json.Unmarshal(*request.Params, &params)
				peerTime = params.Epoch
			} else {
				rsp = true
				if request.Error != nil {
					peerTime = request.Error.Epoch
				} else if request.Result != nil {
					json.Unmarshal(*request.Result, &params)
					peerTime = params.Epoch
				}
			}
			logData := &h.JsonServiceLogsData{
				Service:  service.Name,
				Session:  s.SID,
				Method:   request.Method,
				Response: rsp,
				PeerTime: peerTime,
				//Content:  types.JSONText(msg),
			}
			h.JsonServiceLogs().Create(env, logData)
		})
	*/
	return
}

/*
func (service *Service) Send(sid string, data interface{}) error {
	var session interface{}
	var ok bool
	var err error
	if session, ok = service.Sessions.Load(sid); ok {
		return errors.New("Session not found")
	}
	if s, ok := session.(*melody.Session); ok {
		byteSlice, err := json.Marshal(data)
		if err == nil {
			s.Write(byteSlice)
		}
	}
	return err
}
*/

func (service *Service) Dispatch(s *Session, msg []byte) (interface{}, error) {
	var data interface{}
	var err error
	var request RequestRPC
	err = json.Unmarshal(msg, &request)
	if err != nil {
		return nil, errors.New("Unmarshal JSON data error:" + err.Error())
	}
	if request.JsonRPC == "" {
		return nil, errors.New("Unmarshal JSON data error(JsonRPC = null):" + err.Error())
	}
	service.Log(s, &request, msg)
	if request.Params == nil {
		var response ResponseRPC
		err = json.Unmarshal(msg, &response)
		if err != nil {
			ss := fmt.Sprintf("Service `%s` ResponseRPC: Unmarshal Error", service.Name)
			log.Info(ss)
			return nil, nil
		}
		if response.Result == nil && response.Error == nil {
			ss := fmt.Sprintf("Service `%s` ResponseRPC: Invalid response, result and error is nil", service.Name)
			log.Info(ss)
			return nil, nil
		}

		methodName := response.Method

		if methodName == "" {
			ss := fmt.Sprintf("Service `%s` ResponseRPC: empty method. TODO: Search in session", service.Name)
			log.Info(ss)
			return nil, nil
		}
		fn, ok := service.responses[methodName]
		if !ok {
			ss := fmt.Sprintf("Service `%s` ResponseRPC Method `%s` was not registered ", service.Name, methodName)
			log.Info(ss)
			return nil, nil
		}
		if fn == nil {
			ss := fmt.Sprintf("Service `%s` ResponseRPC Method `%s` is nil", service.Name, methodName)
			log.Info(ss)
			return nil, nil
		}
		fn(s, &response)
		return nil, nil
	}

	methodName := request.Method
	fn, ok := service.methods[methodName]
	if !ok {
		rpcErr := server.JSONRPCError{Code: -32601, Message: "Method not found:" + methodName}
		response := &server.ResponseError{JsonRPC: request.JsonRPC,
			ID:    request.ID,
			Error: rpcErr,
		}
		return response, errors.New("Method not found")
	}
	if fn != nil {
		var strUlid string
		strIface, exists := s.Get("ulid")
		if exists {
			strUlid = strIface.(string)
			ss := fmt.Sprintf("Service `%s` %s \"%s\" send request.", service.Name, methodName, strUlid)
			log.Info(ss)
		}
		return fn(s, &request)
	} else {
		rpcErr := server.JSONRPCError{Code: -32601, Message: "Method not found:" + methodName}
		response := &server.ResponseError{
			JsonRPC: request.JsonRPC,
			ID:      request.ID,
			Error:   rpcErr,
		}
		return response, errors.New("Method not found")
	}
	return data, nil
}

// wrapContextFuncs
func MakeHandleFunc(service *Service) server.HandlerFunc {
	wrappedHandler := func(service *Service) server.HandlerFunc {
		return func(ctx *server.Context) {
			service.HandleRequest(ctx.Writer, ctx.Request)
		}
	}(service)
	return wrappedHandler
}
func (service *Service) GetSession(s *melody.Session) *Session {
	var session *Session
	if s, ok := service.Sessions.Load(s); ok {
		session, ok = s.(*Session)
	}
	if session == nil {
		return nil
	}
	return session
}

func NewService(name string) (*Service, error) {
	var err error
	var data interface{}
	if _, ok := Services.Load(name); ok {
		return nil, errors.New("Already exist")
	}
	service := &Service{Melody: melody.New(), Name: name}
	service.Config.MaxMessageSize = 4096
	service.methods = make(map[string]JsonRPCHandleFunc)
	service.responses = make(map[string]JsonRPCHandleResponseFunc)

	service.HandleMessage(func(s *melody.Session, msg []byte) {
		session := service.GetSession(s)
		if session == nil {
			return
		}
		session.Epoch = int64(ulid.Now())
		/*
			models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				cnnInfo := h.JsonServiceConnection().Search(env, q.JsonServiceConnection().Session().Equals(session.SID))
				if !cnnInfo.IsEmpty() {
					cnnData := &h.JsonServiceConnectionData{
						LastActiveEpoch: session.Epoch,
					}
					cnnInfo.Write(cnnData, h.JsonServiceConnection().LastActiveEpoch())
				}
			})
		*/
		// Call middleware for websocket text
		for _, fn := range service.mw {
			fn(session, msg)
		}
		data, err = service.Dispatch(session, msg)
		if err != nil {
			log.Info("Dispatch error: " + err.Error())
			if data == nil {
				return
			}
		}
		if data == nil {
			return
		}
		byteSlice, err := json.Marshal(data)
		if err == nil {
			session.Write(byteSlice)
		} else {
			session.Write(msg)
		}
	})

	service.HandleMessageBinary(func(s *melody.Session, msg []byte) {
		session := service.GetSession(s)
		if session == nil {
			return
		}
		session.Epoch = int64(ulid.Now())
		/*
			models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				cnnInfo := h.JsonServiceConnection().Search(env, q.JsonServiceConnection().Session().Equals(session.SID))
				if !cnnInfo.IsEmpty() {
					cnnData := &h.JsonServiceConnectionData{
						LastActiveEpoch: session.Epoch,
					}
					cnnInfo.Write(cnnData, h.JsonServiceConnection().LastActiveEpoch())
				}
			})
		*/
		// Call middleware for websocket binary
		for _, fn := range service.mwb {
			fn(session, msg)
		}
		// Echo binary data
		session.Write(msg)
	})
	service.HandleConnect(func(s *melody.Session) {
		suid := NewULID()
		session := &Session{Session: s,
			Service: service,
			Epoch:   int64(ulid.Now()),
			SID:     suid}

		service.Sessions.Store(s, session)
		ss := fmt.Sprintf("%s: Websocket client %s connected (sessionid: %s)", service.Name, s.Request.RemoteAddr, suid)
		log.Info(ss)
		/*
			models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				cnnData := &h.JsonServiceConnectionData{Service: service.Name,
					Session:         suid,
					LastActiveEpoch: session.Epoch,
				}
				cnnData.PeerAddress, cnnData.PeerPort, _ = net.SplitHostPort(s.Request.RemoteAddr)
				h.JsonServiceConnection().Create(env, cnnData)
			})
		*/
	})
	service.HandleDisconnect(func(s *melody.Session) {
		session := service.GetSession(s)
		if session == nil {
			return
		}
		service.Sessions.Delete(s)
		session.Epoch = int64(ulid.Now())
		/*
			models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				cnnInfo := h.JsonServiceConnection().Search(env, q.JsonServiceConnection().Session().Equals(session.SID))
				if !cnnInfo.IsEmpty() {
					cnnData := &h.JsonServiceConnectionData{
						Offline:         true,
						LastActiveEpoch: session.Epoch,
					}
					cnnInfo.Write(cnnData, h.JsonServiceConnection().Offline(),
						h.JsonServiceConnection().LastActiveEpoch())
				}
			})
		*/
		ss := fmt.Sprintf("%s: Websocket client %s disconnected", service.Name, s.Request.RemoteAddr, session.ULID)
		log.Info(ss)
	})
	Services.Store(name, service)
	return service, nil
}

func GetService(name string) (*Service, error) {
	var service interface{}
	var ok bool
	if service, ok = Services.Load(name); !ok {
		return nil, errors.New("Service not found")
	}
	if srv, ok := service.(*Service); ok {
		return srv, nil
	}
	return nil, errors.New("GetService invalid type")
}

func NewULID() string {
	uid := ulid.MustNew(ulid.Now(), rand.Reader)
	return uid.String()
}

func ResetAllSession() {
	/*
		models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			env.Cr().Execute("UPDATE json_service_connection SET offline=true where offline=false")
			//		cnnInfo := h.JsonServiceConnection().Search(env, q.JsonServiceConnection().Session().Equals(session.SID))
			//		if !cnnInfo.IsEmpty() {
			//			cnnData := &h.JsonServiceConnectionData{
			//				LastActiveEpoch: session.Epoch,
			//			}
			//			cnnInfo.Write(cnnData, h.JsonServiceConnection().LastActiveEpoch())
			//		}
		})
	*/
}
func init() {
	serviceConnectionModel := h.JsonServiceConnection().DeclareModel()
	serviceConnectionModel.AddFields(map[string]models.FieldDefinition{
		"Service":     models.CharField{String: "Service", Index: true},
		"Session":     models.CharField{String: "Session", Index: true},
		"PeerAddress": models.CharField{String: "PeerAddress", Index: true},
		"PeerPort":    models.CharField{String: "PeerPort", Help: "Peer Port"},
		"ConnectedEpoch": models.IntegerField{
			String: "Connected",
			Help:   "Connected ",
			GoType: new(int64),
			Default: func(env models.Environment) interface{} {
				return ulid.Now()
			},
		},
		"LastActiveEpoch": models.IntegerField{
			String: "LastActive",
			GoType: new(int64),
		},
		"Offline": models.BooleanField{String: "Offline"},
		"ULID":    models.CharField{String: "ULID", Index: true},
	})
	serviceConnectionModel.SetDefaultOrder("ID DESC")

	serviceLogModel := h.JsonServiceLogs().DeclareModel()
	serviceLogModel.AddFields(map[string]models.FieldDefinition{
		"Service": models.CharField{String: "Service", Index: true},
		"Session": models.CharField{String: "Session", Index: true},
		"UnixTime": models.IntegerField{
			String: "UnixTime",
			Help:   "Server Time ",
			GoType: new(int64),
			Default: func(env models.Environment) interface{} {
				return ulid.Now()
			},
		},
		"PeerTime": models.IntegerField{
			String: "PeerTime",
			Help:   "Peer Time ",
			GoType: new(int64),
		},
		"Response": models.BooleanField{String: "IsResponse"},
		"Method":   models.CharField{String: "Method", Required: false},
		//"Content":  models.JSONField{String: "Content", Index: true},
	})
	serviceLogModel.SetDefaultOrder("ID DESC")
}
