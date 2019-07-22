package websocket

import (
	"encoding/json"
	"errors"

	"github.com/hexya-erp/hexya/src/actions"
	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/security"
	"github.com/hexya-erp/hexya/src/models/types"
	"github.com/hexya-erp/hexya/src/server"
	"github.com/hexya-erp/pool/h"
	"github.com/hexya-erp/pool/q"

	"github.com/hexya-addons/web/controllers"
)

// URL: /menu/load_needaction
/*
func JsonRPCMenuLoadNeedaction(s *melody.Session, r *server.RequestRPC) (interface{}, error) {
	_, exist := s.Get("uid")
	if !exist {
		return nil, errors.New("Not login")
	}
	response := &server.ResponseRPC{
		JsonRPC: r.JsonRPC,
		ID:      r.ID,
		Result:  "success",
	}
	return response, nil
}
*/

// URL: /action/load
// ActionLoad returns the action with the given id
func JsonRPCActionLoad(s *Session, r *RequestRPC) (interface{}, error) {
	var uid int64
	var lang string
	//uidtmp, exist := s.Get("uid")
	uid = s.UID
	if uid > 0 {
		models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			user := h.User().Search(env, q.User().ID().Equals(uid))
			lang = user.ContextGet().GetString("lang")
		})
	}

	params := struct {
		ActionID          string         `json:"action_id"`
		AdditionalContext *types.Context `json:"additional_context"`
	}{}

	err := json.Unmarshal(*r.Params, &params)
	if err != nil {

	}
	action := *actions.Registry.MustGetById(params.ActionID)
	action.Name = action.TranslatedName(lang)
	response := &server.ResponseRPC{
		JsonRPC: r.JsonRPC,
		ID:      r.ID,
		Result:  &action,
	}
	return response, nil
}

// URL: /action/run
// ActionRun runs the given server action
func JsonRPCActionRun(s *Session, r *RequestRPC) (interface{}, error) {

	uid := s.UID
	if uid == 0 {
		return nil, errors.New("Access denied")
	}

	params := struct {
		ActionID string         `json:"action_id"`
		Context  *types.Context `json:"context"`
	}{}

	err := json.Unmarshal(*r.Params, &params)
	if err != nil {
		return nil, errors.New("JsonRPCActionRun error: Invalid format")
	}
	action := actions.Registry.MustGetById(params.ActionID)

	// Process context ids into args
	var ids []int64
	if params.Context.Get("active_ids") != nil {
		ids = params.Context.Get("active_ids").([]int64)
	} else if params.Context.Get("active_id") != nil {
		ids = []int64{params.Context.Get("active_id").(int64)}
	}
	idsJSON, err := json.Marshal(ids)
	if err != nil {
		//log.Panic("Unable to marshal ids")
		return nil, err
	}

	// Process context into kwargs
	contextJSON, _ := json.Marshal(params.Context)
	kwargs := make(map[string]json.RawMessage)
	kwargs["context"] = contextJSON

	// Execute the function
	resAction, _ := controllers.Execute(uid, controllers.CallParams{
		Model:  action.Model,
		Method: action.Method,
		Args:   []json.RawMessage{idsJSON},
		KWArgs: kwargs,
	})

	if _, ok := resAction.(*actions.Action); ok {
		//c.RPC(http.StatusOK, resAction)
		response := &server.ResponseRPC{
			JsonRPC: r.JsonRPC,
			ID:      r.ID,
			Result:  "success",
		}
		return response, nil
	} else {
		//c.RPC(http.StatusOK, false)
		return nil, errors.New("Not action")
	}
}
