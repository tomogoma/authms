package http

import (
	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/model"
)

/**
 * @api {NULL} Group
 * @apiName Group
 * @apiGroup Objects
 *
 * @apiSuccess {String} ID Unique ID of the group (can be cast to long Integer).
 * @apiSuccess {String} name The unique group name string value.
 * @apiSuccess {Integer} accessLevel The access level for this group in (0 >= accessLevel <= 10)
 * @apiSuccess {String} created ISO8601 date the group was created.
 * @apiSuccess {String} lastUpdated ISO8601 date the group was last updated.
 */
type Group struct {
	ID          string  `json:"ID,omitempty"`
	Name        string  `json:"name,omitempty"`
	AccessLevel float32 `json:"accessLevel,omitempty"`
	CreateDate  string  `json:"created,omitempty"`
	UpdateDate  string  `json:"lastUpdated,omitempty"`
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
