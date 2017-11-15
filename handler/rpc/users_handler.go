package rpc

import (
	"github.com/micro/go-micro/server"
	"github.com/pborman/uuid"
	"github.com/sirupsen/logrus"
	"github.com/tomogoma/authms/logging"
	"github.com/tomogoma/authms/model"
	"golang.org/x/net/context"
	errors "github.com/tomogoma/go-typed-errors"
	"github.com/tomogoma/authms/api"
	"github.com/tomogoma/authms/config"
)

type Guard interface {
	APIKeyValid(key string) (string, error)
}

type UsersModel interface {
	errors.AllErrChecker
	GetUserDetails(JWT string, userID string) (*model.User, error)
}

type UsersHandler struct {
	errors.AllErrCheck
	guard  Guard
	usersM UsersModel
}

const (
	internalErrorMessage = "whoops! Something wicked happened"

	ctxKeyLog = "log"
)

func NewHandler(g Guard, um UsersModel) (*UsersHandler, error) {
	if g == nil {
		return nil, errors.New("nil Guard")
	}
	if um == nil {
		return nil, errors.New("nil UsersModel")
	}
	return &UsersHandler{usersM: um}, nil
}

func LogWrapper(next server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) error {
		log := logrus.WithField(logging.FieldTransID, uuid.New())
		log.WithFields(logrus.Fields{
			logging.FieldTransID:        uuid.New(),
			logging.FieldService:        req.Service(),
			logging.FieldRPCMethod:      req.Method(),
			logging.FieldRequestHandler: "RPC",
		}).Info("new request")
		ctx = context.WithValue(ctx, ctxKeyLog, log)
		return next(ctx, req, rsp)
	}
}

func (h *UsersHandler) Wrapper(next server.HandlerFunc) server.HandlerFunc {
	return LogWrapper(next)
}

func (h *UsersHandler) GetDetails(ctx context.Context, req *api.GetDetailsReq, resp *api.User) error {

	if req == nil || resp == nil {
		return errors.Newf("req/response had nil value")
	}

	ctx, err := h.APIKeyValid(ctx, req.APIKey)
	if err != nil {
		return h.processError(ctx, err)
	}

	usr, err := h.usersM.GetUserDetails(req.JWT, req.UserID)
	if err != nil {
		return h.processError(ctx, err)
	}

	packageUser(usr, resp)
	return nil
}

func (h *UsersHandler) APIKeyValid(ctx context.Context, APIKey string) (context.Context, error) {

	clUsrID, err := h.guard.APIKeyValid(APIKey)

	log := ctx.Value(ctxKeyLog).(logging.Logger).
		WithField(logging.FieldClientAppUserID, clUsrID)
	ctx = context.WithValue(ctx, ctxKeyLog, log)

	return ctx, err
}

func packageUser(usr *model.User, resp *api.User) {
	if usr == nil || resp == nil {
		return
	}

	resp.ID = usr.ID
	resp.JWT = usr.JWT
	resp.Created = usr.CreateDate.Format(config.TimeFormat)
	resp.LastUpdated = usr.UpdateDate.Format(config.TimeFormat)

	packageUserType(&usr.Type, resp.Type)
	packageUserName(&usr.UserName, resp.Username)
	packageVerifLogin(&usr.Phone, resp.Phone)
	packageVerifLogin(&usr.Email, resp.Email)
	packageFacebook(&usr.Facebook, resp.Facebook)
	packageGroups(usr.Groups, resp.Groups)
	packageDevices(usr.Devices, resp.Devices)
}

func packageUserType(ut *model.UserType, resp *api.UserType) {
	if ut == nil || !ut.HasValue() || resp == nil {
		return
	}
	resp.ID = ut.ID
	resp.Name = ut.Name
	resp.Created = ut.CreateDate.Format(config.TimeFormat)
	resp.LastUpdated = ut.UpdateDate.Format(config.TimeFormat)
}

func packageUserName(un *model.Username, resp *api.UserName) {
	if un == nil || !un.HasValue() || resp == nil {
		return
	}
	resp.ID = un.ID
	resp.UserID = un.UserID
	resp.Value = un.Value
	resp.Created = un.CreateDate.Format(config.TimeFormat)
	resp.LastUpdated = un.UpdateDate.Format(config.TimeFormat)
}

func packageVerifLogin(vl *model.VerifLogin, resp *api.VerifLogin) {
	if vl == nil || !vl.HasValue() || resp == nil {
		return
	}
	resp.ID = vl.ID
	resp.UserID = vl.UserID
	resp.Value = vl.Address
	resp.Verified = vl.Verified
	resp.Created = vl.CreateDate.Format(config.TimeFormat)
	resp.LastUpdated = vl.UpdateDate.Format(config.TimeFormat)
}

func packageFacebook(vl *model.Facebook, resp *api.Facebook) {
	if vl == nil || !vl.HasValue() || resp == nil {
		return
	}
	resp.ID = vl.ID
	resp.UserID = vl.UserID
	resp.FacebookID = vl.FacebookID
	resp.Verified = vl.Verified
	resp.Created = vl.CreateDate.Format(config.TimeFormat)
	resp.LastUpdated = vl.UpdateDate.Format(config.TimeFormat)
}

func packageGroup(g *model.Group, resp *api.Group) {
	if g == nil || !g.HasValue() || resp == nil {
		return
	}
	resp.ID = g.ID
	resp.Name = g.Name
	resp.AccessLevel = g.AccessLevel
	resp.Created = g.CreateDate.Format(config.TimeFormat)
	resp.LastUpdated = g.UpdateDate.Format(config.TimeFormat)
}

func packageGroups(gs []model.Group, resp []*api.Group) {
	if gs == nil || resp == nil {
		return
	}
	for _, g := range gs {
		if !g.HasValue() {
			continue
		}
		rg := &api.Group{}
		packageGroup(&g, rg)
		resp = append(resp, rg)
	}
}

func packageDevice(d *model.Device, resp *api.Device) {
	if d == nil || !d.HasValue() || resp == nil {
		return
	}
	resp.ID = d.ID
	resp.UserID = d.UserID
	resp.DeviceID = d.DeviceID
	resp.Created = d.CreateDate.Format(config.TimeFormat)
	resp.LastUpdated = d.UpdateDate.Format(config.TimeFormat)
}

func packageDevices(ds []model.Device, resp []*api.Device) {
	if ds == nil || resp == nil {
		return
	}
	for _, d := range ds {
		if !d.HasValue() {
			continue
		}
		rd := &api.Device{}
		packageDevice(&d, rd)
		resp = append(resp, rd)
	}
}
