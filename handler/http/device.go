package http

import (
	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/model"
)

/**
 * @api {NULL} Device
 * @apiName Device
 * @apiGroup Objects
 *
 * @apiSuccess {String} ID Unique ID of the device (can be cast to long Integer).
 * @apiSuccess {String} userID ID for user who owns this device ID.
 * @apiSuccess {String} deviceID The unique device ID string value.
 * @apiSuccess {String} created ISO8601 date the device was created.
 * @apiSuccess {String} lastUpdated ISO8601 date the device was last updated.
 */
type Device struct {
	ID         string `json:"ID,omitempty"`
	UserID     string `json:"userID,omitempty"`
	DeviceID   string `json:"deviceID,omitempty"`
	CreateDate string `json:"created,omitempty"`
	UpdateDate string `json:"lastUpdated,omitempty"`
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
