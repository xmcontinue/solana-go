package iface

import (
	"git.cplus.link/crema/backend/chain/sol"
)

type GetGalleryReq struct {
	Query            string   `json:"query"                      binding:"omitempty"`
	CoffeeMembership []string `json:"Coffee Membership"          binding:"omitempty"`
	Body             []string `json:"Body"                       binding:"omitempty"`
	FacialFeatures   []string `json:"Facial Features"            binding:"omitempty"`
	Head             []string `json:"Head"                       binding:"omitempty"`
	FacialAccessory  []string `json:"Facial Accessory"           binding:"omitempty"`
	Clothes          []string `json:"Clothes"                    binding:"omitempty"`
	Accessory        []string `json:"Accessory"                  binding:"omitempty"`
	Shell            []string `json:"Shell"                      binding:"omitempty"`
	Cup              []string `json:"Cup"                        binding:"omitempty"`
	Background       []string `json:"Background"                 binding:"omitempty"`
	ISPositive       bool     `json:"is_positive"                binding:"omitempty"`
	Offset           int64    `json:"offset"                     form:"offset"           binding:"omitempty"`
	Limit            int64    `json:"limit"                      form:"limit"            binding:"required"`
}

type GetGalleryResp struct {
	Total  int64          `json:"total"`
	Limit  int64          `json:"limit"  gquery:"-"`
	Offset int64          `json:"offset" gquery:"-"`
	List   []*sol.Gallery `json:"list"`
}
