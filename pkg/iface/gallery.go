package iface

import (
	"git.cplus.link/crema/backend/chain/sol"
)

type GetGalleryReq struct {
	Query      string   `json:"query"                      binding:"omitempty"`
	Level      []string `json:"level,omitempty"            binding:"omitempty"`
	Body       []string `json:"body,omitempty"             binding:"omitempty"`
	ISPositive bool     `json:"is_positive"                binding:"omitempty"`
	Offset     int64    `json:"offset"                     form:"offset"           binding:"omitempty"`
	Limit      int64    `json:"limit"                      form:"limit"            binding:"required"`
}

type GetGalleryResp struct {
	Total  int64          `json:"total"`
	Limit  int64          `json:"limit"  gquery:"-"`
	Offset int64          `json:"offset" gquery:"-"`
	List   []*sol.Gallery `json:"list"`
}
