package websocket

import (
	"encoding/json"
	"errors"
	//"fmt"

	"github.com/hexya-erp/hexya/src/actions"
	"github.com/hexya-erp/hexya/src/server"
	//	"github.com/hexya-erp/hexya/hexya/controllers"
	//	"github.com/hexya-erp/hexya/hexya/models"
	//	"github.com/hexya-erp/hexya/hexya/models/security"
	//	"github.com/hexya-erp/hexya/src/models/types"
	//	"github.com/hexya-erp/pool"

	hc "github.com/hexya-addons/web/controllers"
)

// URL: /dataset/call_kw/*path
func JsonRPCCallKW(s *Session, r *RequestRPC) (interface{}, error) {
	uid := s.UID
	if uid == 0 {
		return nil, errors.New("Access denied")
	}
	var params hc.CallParams
	err := json.Unmarshal(*r.Params, &params)
	if err != nil {
		return nil, errors.New("JsonRPCCallKW error: Invalid format")
	}
	res, err := hc.Execute(uid, params)
	response := &server.ResponseRPC{
		JsonRPC: r.JsonRPC,
		ID:      r.ID,
		Result:  &res,
	}
	return response, nil
}

// URL: /dataset/call_button
func JsonRPCCallButton(s *Session, r *RequestRPC) (interface{}, error) {
	uid := s.UID
	if uid == 0 {
		return nil, errors.New("Access denied")
	}
	var params hc.CallParams
	err := json.Unmarshal(*r.Params, &params)
	if err != nil {
		return nil, errors.New("JsonRPCCallButton error: Invalid format")
	}

	res, err := hc.Execute(uid, params)
	_, isAction := res.(actions.Action)
	_, isActionPtr := res.(*actions.Action)
	if !isAction && !isActionPtr {
		res = false
	}

	response := &server.ResponseRPC{
		JsonRPC: r.JsonRPC,
		ID:      r.ID,
		Result:  &res,
	}
	return response, nil
}

// URL: /dataset/search_read
func JsonRPCSearchRead(s *Session, r *RequestRPC) (interface{}, error) {
	uid := s.UID
	if uid == 0 {
		return nil, errors.New("Access denied")
	}
	var params hc.SearchReadParams
	err := json.Unmarshal(*r.Params, &params)
	if err != nil {
		return nil, errors.New("JsonRPCSearchRead error: Invalid format")
	}

	res, err := hc.SearchRead(uid, params)

	response := &server.ResponseRPC{
		JsonRPC: r.JsonRPC,
		ID:      r.ID,
		Result:  &res,
	}
	return response, nil
}
