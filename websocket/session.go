package websocket

import (
	"encoding/json"
	"errors"
	//"fmt"
	//"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid"

	hc "github.com/hexya-addons/web/controllers"
	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/security"
	"github.com/hexya-erp/hexya/src/models/types"
	"github.com/hexya-erp/hexya/src/server"
	"github.com/hexya-erp/pool/h"
	"github.com/hexya-erp/pool/q"
)

// URL: /locale
// URL: /session/get_session_info
// SessionInfo returns a map with information about the given session
func JsonRPCSessionInfo(s *Session, r *RequestRPC) (interface{}, error) {
	uid := s.UID
	var (
		userContext *types.Context
		companyID   int64
		userName    string
	)
	if uid > 0 {
		models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			user := h.User().Search(env, q.User().ID().Equals(uid))
			userContext = user.ContextGet()
			companyID = user.Company().ID()
			userName = user.Name()
		})
		sid, _ := s.Get("sid")
		data := gin.H{
			"epoch":        int64(ulid.Now()),
			"session_id":   sid,
			"uid":          uid,
			"user_context": userContext.ToMap(),
			"db":           "default",
			"username":     userName,
			"company_id":   companyID,
		}
		return &server.ResponseRPC{
			JsonRPC: r.JsonRPC,
			ID:      r.ID,
			Result:  data,
		}, nil
	}
	nodata := gin.H{
		"epoch": int64(ulid.Now()),
		"uid":   0,
	}
	return &server.ResponseRPC{
		JsonRPC: r.JsonRPC,
		ID:      r.ID,
		//Result:  gin.H{},
		Result: &nodata,
	}, nil
}

// URL: /session/modules
func JsonRPCModules(s *Session, r *RequestRPC) (interface{}, error) {
	uid := s.UID
	if uid == 0 {
		return nil, errors.New("Access denied")
	}
	mods := make([]string, len(server.Modules))
	for i, m := range server.Modules {
		mods[i] = m.Name
	}
	data := gin.H{
		"epoch": int64(ulid.Now()),
		//"session_id": sid,
		"uid":   uid,
		"menus": mods,
	}
	return &server.ResponseRPC{
		JsonRPC: r.JsonRPC,
		ID:      r.ID,
		Result:  data,
	}, nil
}

// URL: /change_password
// ChangePassword is called by the client to change the current user password
func JsonRPCChangePassword(s *Session, r *RequestRPC) (interface{}, error) {
	uid := s.UID
	if uid == 0 {
		return nil, errors.New("Access denied")
	}

	var params hc.ChangePasswordData
	err := json.Unmarshal(*r.Params, &params)
	if err != nil {
		return nil, errors.New("JsonRPCChangePassword error: Invalid format")
	}
	var oldPassword, newPassword, confirmPassword string
	for _, d := range params.Fields {
		switch d.Name {
		case "old_pwd":
			oldPassword = d.Value
		case "new_password":
			newPassword = d.Value
		case "confirm_pwd":
			confirmPassword = d.Value
		}
	}
	res := make(gin.H)
	err = models.ExecuteInNewEnvironment(uid, func(env models.Environment) {
		rs := h.User().NewSet(env)
		if strings.TrimSpace(oldPassword) == "" ||
			strings.TrimSpace(newPassword) == "" ||
			strings.TrimSpace(confirmPassword) == "" {
			log.Panic(rs.T("You cannot leave any password empty."))
		}
		if newPassword != confirmPassword {
			log.Panic(rs.T("The new password and its confirmation must be identical."))
		}
		if rs.ChangePassword(oldPassword, newPassword) {
			res["new_password"] = newPassword
			return
		}
		log.Panic(rs.T("Error, password not changed !"))
	})
	if err != nil {
		return nil, errors.New("Change password error")
	}
	data := gin.H{
		"epoch": int64(ulid.Now()),
		//"session_id":  sid,
		"uid":         uid,
		"changpasswd": "success",
	}
	response := &server.ResponseRPC{
		JsonRPC: r.JsonRPC,
		ID:      r.ID,
		Result:  &data,
	}
	return response, nil
}

func JsonRPCToken(s *Session, r *RequestRPC) (interface{}, error) {
	var (
		company_id int64
		lid        string
		login      string
		email      string
	)

	uid := s.UID
	if uid == 0 {
		return nil, errors.New("Access denied")
	}

	err := models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
		userInfo := h.User().Search(env, q.User().ID().Equals(uid))
		if !userInfo.IsEmpty() {
			lid = userInfo.Ulid()
			company_id = userInfo.Company().ID()
			login = userInfo.Name()
			email = userInfo.Email()
		}
		return
	})

	if err != nil || lid == "" {
		return nil, errors.New("User not exist")
	}

	token, _ := idp.TemporaryKey(lid)
	refresh, _ := idp.RefreshKey(lid)

	data := gin.H{
		"epoch":         int64(ulid.Now()),
		"uid":           uid,
		"company_id":    company_id,
		"ulid":          lid,
		"login":         login,
		"email":         email,
		"token":         token,
		"refresh_token": refresh,
	}
	response := &server.ResponseRPC{
		JsonRPC: r.JsonRPC,
		ID:      r.ID,
		Result:  &data,
	}
	return response, nil
}
