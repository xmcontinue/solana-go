package iface

import (
	"git.cplus.link/crema/backend/chain/sol"
)

type GetGalleryReq struct {
	Query            string   `json:"query"                      redisKey:"-"                   binding:"omitempty"`
	CoffeeMembership []string `json:"Coffee Membership"          redisKey:"CoffeeMembership"    binding:"omitempty"`
	Body             []string `json:"Body"                       redisKey:"Body"                binding:"omitempty"`
	FacialFeatures   []string `json:"Facial Features"            redisKey:"FacialFeatures"      binding:"omitempty"`
	Head             []string `json:"Head"                       redisKey:"Head"                binding:"omitempty"`
	FacialAccessory  []string `json:"Facial Accessory"           redisKey:"FacialAccessory"     binding:"omitempty"`
	Clothes          []string `json:"Clothes"                    redisKey:"Clothes"             binding:"omitempty"`
	Accessory        []string `json:"Accessory"                  redisKey:"Accessory"           binding:"omitempty"`
	Shell            []string `json:"Shell"                      redisKey:"Shell"               binding:"omitempty"`
	Cup              []string `json:"Cup"                        redisKey:"Cup"                 binding:"omitempty"`
	Background       []string `json:"Background"                 redisKey:"Background"          binding:"omitempty"`
	ISPositive       bool     `json:"is_positive"                redisKey:"-"                   binding:"omitempty"`
	Offset           int64    `json:"offset"                     redisKey:"-"                form:"offset"           binding:"omitempty"`
	Limit            int64    `json:"limit"                      redisKey:"-"                form:"limit"            binding:"required"`
}

type GetGalleryResp struct {
	Total  int64          `json:"total"`
	Limit  int64          `json:"limit"  gquery:"-"`
	Offset int64          `json:"offset" gquery:"-"`
	List   []*sol.Gallery `json:"list"`
}
