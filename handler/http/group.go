package http

import (
	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/model"
)

type Group struct {
	ID          string  `json:"ID,omitempty"`
	Name        string  `json:"name,omitempty"`
	AccessLevel float32 `json:"accessLevel,omitempty"`
	CreateDate  string  `json:"createDate,omitempty"`
	UpdateDate  string  `json:"updateDate,omitempty"`
}

func NewGroup(g model.Group) *Group {
	if !g.HasValue() {
		return nil
	}
	return &Group{
		ID:          g.ID,
		Name:        g.Name,
		AccessLevel: g.AccessLevel,
		CreateDate:  g.CreateDate.Format(config.TimeFormat),
		UpdateDate:  g.UpdateDate.Format(config.TimeFormat),
	}
}

func NewGroups(gs []model.Group) []Group {
	var grps []Group
	for _, g := range gs {
		grp := NewGroup(g)
		if grp == nil {
			continue
		}
		grps = append(grps, *grp)
	}
	return grps
}
