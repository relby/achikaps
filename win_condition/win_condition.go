package win_condition

import "github.com/relby/achikaps/model"

type WinCondition struct {
	MaterialType model.MaterialType
	Count        int
}

func New(materialType model.MaterialType, count int) *WinCondition {
	return &WinCondition{
		materialType,
		count,
	}
}
