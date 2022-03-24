package graph

import (
	"git.cplus.link/crema/backend/pkg/domain"
)

type ShotPath struct {
	Data []*domain.Price
}

func NewShopPath() *ShotPath {
	return &ShotPath{}
}
