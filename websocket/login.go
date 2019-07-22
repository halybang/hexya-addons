package websocket

import (
	"encoding/json"
	"errors"
	//"fmt"

	"github.com/gin-gonic/gin"

	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/security"
	"github.com/hexya-erp/hexya/src/models/types"
	"github.com/hexya-erp/hexya/src/server"
	"github.com/hexya-erp/pool/h"
	"github.com/hexya-erp/pool/q"
)

type Login struct {
	User     string `form:"user" json:"user" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
	Database string `form:"database" json:"domain,,omitempty"`
}

type LoginResponse struct {
	ID           int64       `json:"id"`
	User         string      `json:"username"`
	Email        string      `json:"email"`
	Ulid         string      `json:"ulid"`
	Token        string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	Database     string      `json:"domain,omitempty"`
	Company      int64       `json:"company,omitempty"`
	Modules      interface{} `json:"modules,omitempty"`
}

// URL: /login
func JsonRPCLogin(s *Session, r *RequestRPC) (interface{}, error) {
	var login Login
	err := json.Unmarshal(*r.Params, &login)
	if err != nil {
		return nil, errors.New("Login in error: Invalid format")
	}

	var ulid string
	var userName string
	var company_id int64

	uid, err := security.AuthenticationRegistry.Authenticate(login.User, login.Password, new(types.Context))
	if err != nil {
		rpcErr := server.JSONRPCError{Code: -32000, Message: err.Error()}
		response := &server.ResponseError{JsonRPC: r.JsonRPC,
			ID:    r.ID,
			Error: rpcErr,
		}
		return response, err //errors.New("Wrong username or password")
	}
	err = models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
		userInfo := h.User().Search(env, q.User().ID().Equals(uid))
		if !userInfo.IsEmpty() {
			company_id = userInfo.Company().ID()
			ulid = userInfo.Ulid()
			userName = userInfo.Name()
			s.UID = uid
			s.ULID = ulid
			s.Set("login", userName)
			s.Set("company_id", company_id)
		} else {
			s.UID = 0
			s.ULID = ""
			s.Set("login", nil)
			s.Set("company_id", nil)
		}
		return
	})

	if err != nil || ulid == "" {
		return nil, errors.New("User not exist")
	}

	if s.ULID != "" {
		/*
			models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				cnnInfo := h.JsonServiceConnection().Search(env, q.JsonServiceConnection().Session().Equals(s.SID))
				if !cnnInfo.IsEmpty() {
					cnnData := &h.JsonServiceConnectionData{
						ULID: s.ULID,
					}
					cnnInfo.Write(cnnData, h.JsonServiceConnection().ULID())
				}
			})
		*/
	}

	mods := make([]string, len(server.Modules))
	for i, m := range server.Modules {
		mods[i] = m.Name
	}
	token, _ := idp.TemporaryKey(ulid)
	refresh, _ := idp.RefreshKey(ulid)
	res := &LoginResponse{ID: uid,
		User:         login.User,
		Ulid:         ulid,
		Database:     "default",
		Company:      company_id,
		Token:        token,
		RefreshToken: refresh,
		Modules:      mods,
	}

	// TODO: Log response

	response := &server.ResponseRPC{
		JsonRPC: r.JsonRPC,
		ID:      r.ID,
		Result:  &res,
	}
	return response, nil
}

// URL: /session/logout
func JsonRPCLogout(s *Session, r *RequestRPC) (interface{}, error) {
	s.UID = 0
	s.ULID = ""
	response := &server.ResponseRPC{
		JsonRPC: r.JsonRPC,
		ID:      r.ID,
		Result:  "success",
	}
	return response, nil
}

// URL: /version
// VersionInfo returns server version information to the client
func JsonRPCVersionInfo(s *Session, r *RequestRPC) (interface{}, error) {
	data := gin.H{
		"server_serie":        "0.9beta",
		"server_version_info": []int8{0, 9, 0, 0, 0},
		"server_version":      "0.9beta",
		"protocol":            1,
	}
	response := &server.ResponseRPC{
		JsonRPC: r.JsonRPC,
		ID:      r.ID,
		Result:  &data,
	}
	return response, nil
}

func JsonRPCHandleResponsePing(s *Session, response *ResponseRPC) {
	if response.Result == nil {
		if response.Error == nil {
			log.Info("Invalid ResponseRPC: Result and error is nil")
		} else {
			log.Info("Process ResponseRPC->JSONRPCError", response.Error)
		}
	} else {
		log.Info("Process ResponseRPC->Result", response.Result)
	}
}
