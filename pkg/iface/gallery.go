package iface

import (
	"git.cplus.link/crema/backend/chain/sol"
)

type GetGalleryReq struct {
	Query string `json:"query"                      redisKey:"-"                   binding:"omitempty"`
	*GalleryType
	ISPositive bool  `json:"is_positive"                redisKey:"-"                   binding:"omitempty"`
	Offset     int64 `json:"offset"                     redisKey:"-"                form:"offset"           binding:"omitempty"`
	Limit      int64 `json:"limit"                      redisKey:"-"                form:"limit"            binding:"required"`
}

type GalleryType struct {
	CoffeeMembership []string `json:"Coffee Membership"          yaml:"CoffeeMembership"    binding:"omitempty"`
	Body             []string `json:"Body"                       yaml:"Body"                binding:"omitempty"`
	FacialFeatures   []string `json:"Facial Features"            yaml:"FacialFeatures"      binding:"omitempty"`
	Head             []string `json:"Head"                       yaml:"Head"                binding:"omitempty"`
	FacialAccessory  []string `json:"Facial Accessory"           yaml:"FacialAccessory"     binding:"omitempty"`
	Clothes          []string `json:"Clothes"                    yaml:"Clothes"             binding:"omitempty"`
	Accessory        []string `json:"Accessory"                  yaml:"Accessory"           binding:"omitempty"`
	Shell            []string `json:"Shell"                      yaml:"Shell"               binding:"omitempty"`
	Cup              []string `json:"Cup"                        yaml:"Cup"                 binding:"omitempty"`
	Background       []string `json:"Background"                 yaml:"Background"          binding:"omitempty"`
}

type GetGalleryResp struct {
	Total  int64          `json:"total"`
	Limit  int64          `json:"limit"  gquery:"-"`
	Offset int64          `json:"offset" gquery:"-"`
	List   []*sol.Gallery `json:"list"`
}

type GetGalleryTypeResp struct {
	*GalleryType
}
