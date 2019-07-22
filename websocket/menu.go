package websocket

import (
	//"encoding/json"
	"errors"
	//"fmt"
	//"net/http"

	//	"github.com/hexya-erp/hexya/hexya/actions"
	//	"github.com/hexya-erp/hexya/hexya/controllers"
	//	"github.com/hexya-erp/hexya/hexya/models"
	//	"github.com/hexya-erp/hexya/hexya/models/security"
	//	"github.com/hexya-erp/hexya/hexya/models/types"
	"github.com/hexya-erp/hexya/src/server"
	//	"github.com/hexya-erp/pool"
)

// URL: /menu/load_needaction
func JsonRPCMenuLoadNeedaction(s *Session, r *RequestRPC) (interface{}, error) {
	uid := s.UID
	if uid == 0 {
		return nil, errors.New("Access denied")
	}
	response := &server.ResponseRPC{
		JsonRPC: r.JsonRPC,
		ID:      r.ID,
		Result:  "success",
	}
	return response, nil
}
