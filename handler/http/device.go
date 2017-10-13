package http

import (
	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/model"
)

type Device struct {
	ID         string `json:"ID,omitempty"`
	UserID     string `json:"userID,omitempty"`
	DeviceID   string `json:"deviceID,omitempty"`
	CreateDate string `json:"createDate,omitempty"`
	UpdateDate string `json:"updateDate,omitempty"`
}

func NewDevice(g model.Device) *Device {
	if !g.HasValue() {
		return nil
	}
	return &Device{
		ID:         g.ID,
		UserID:     g.UserID,
		DeviceID:   g.DeviceID,
		CreateDate: g.CreateDate.Format(config.TimeFormat),
		UpdateDate: g.UpdateDate.Format(config.TimeFormat),
	}
}

func NewDevices(ds []model.Device) []Device {
	var devs []Device
	for _, d := range ds {
		dev := NewDevice(d)
		if dev == nil {
			continue
		}
		devs = append(devs, *dev)
	}
	return devs
}
