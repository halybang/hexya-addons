package websocket

import (
	_ "github.com/hexya-addons/base"
	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/pool/h"
)

func init() {
	h.Company().AddFields(map[string]models.FieldDefinition{
		"Ulid": models.CharField{JSON: `ulid`, Required: true, Index: true, Unique: true,
			Default: func(env models.Environment) interface{} { return NewULID() },
		},
	},
	)
	h.Partner().AddFields(map[string]models.FieldDefinition{
		"Ulid": models.CharField{JSON: `ulid`, Required: true, Index: true, Unique: true,
			Default: func(env models.Environment) interface{} { return NewULID() },
		},
	},
	)
	h.User().AddFields(map[string]models.FieldDefinition{
		"Ulid": models.CharField{JSON: `ulid`, Required: true, Index: true, Unique: true,
			Default: func(env models.Environment) interface{} { return NewULID() },
		},
	},
	)
}
