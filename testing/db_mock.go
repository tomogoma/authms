package testing

import (
	"database/sql"
	"time"
	"reflect"

	"github.com/tomogoma/authms/api"
	"github.com/tomogoma/authms/model"
	"github.com/tomogoma/go-commons/errors"
)

type DBMock struct {
	errors.NotFoundErrCheck

	ExpInsGrpErr error
	ExpGrpBNm    *model.Group
	ExpGrpBNmErr error
	ExpGrp       *model.Group
	ExpGrpErr    error

	ExpInsUsrTypErr error
	ExpUsrTypBNm    *model.UserType
	ExpUsrTypBNmErr error

	ExpInsUsrErr    error
	ExpInsUsrAtmErr error
	ExpUsr          *model.User
	ExpUsrErr       error
	ExpUsrBPhn      *model.User
	ExpUsrBPhnErr   error
	ExpUsrBMail     *model.User
	ExpUsrBMailErr  error
	ExpUsrBUsrNm    *model.User
	ExpUsrBUsrNmErr error
	ExpUsrBFb       *model.User
	ExpUsrBFbErr    error
	ExpUsrBDev      *model.User
	ExpUsrBDevErr   error

	ExpupdPassErr    error
	ExpupdPassAtmErr error
	ExpUsrPass       []byte
	ExpUsrBDevPass   []byte
	ExpUsrBUsrNmPass []byte
	ExpUsrBPhnPass   []byte
	ExpUsrBMailPass  []byte

	ExpInsAPIKErr     error
	ExpAPIKsBUsrID    []api.Key
	ExpAPIKsBUsrIDErr error

	ExpAddUsrTGrpAtmcErr error

	ExpInsDevAtmErr error

	ExpInsUsrNmErr    error
	ExpInsUsrNmAtmErr error
	ExpUpdUsrNm       *model.Username
	ExpUpdUsrNmErr    error

	ExpInsUsrPhnErr    error
	ExpInsUsrPhnAtmErr error
	ExpUpdUsrPhn       *model.VerifLogin
	ExpUpdUsrPhnErr    error
	ExpUpdUsrPhnAtm    *model.VerifLogin
	ExpUpdUsrPhnAtmErr error

	ExpInsPhnTknErr    error
	ExpInsPhnTknAtmErr error
	ExpPhnTkns         []model.DBToken
	ExpPhnTknsErr      error

	ExpInsUsrMailErr    error
	ExpInsUsrMailAtmErr error
	ExpUpdUsrMail       *model.VerifLogin
	ExpUpdUsrMailErr    error
	ExpUpdUsrMailAtm    *model.VerifLogin
	ExpUpdUsrMailAtmErr error

	ExpInsMailTknErr    error
	ExpInsMailTknAtmErr error
	ExpMailTkns         []model.DBToken
	ExpMailTknsErr      error

	ExpInsFbAtmErr error

	ExpUpsSMTPConfErr error
	ExpSMTPConf       model.SMTPConfig
	ExpSMTPConfErr    error

	isInTx bool
}

func (db *DBMock) ExecuteTx(fn func(*sql.Tx) error) error {
	db.isInTx = true
	defer func() {
		db.isInTx = false
	}()
	return fn(new(sql.Tx))
}

func (db *DBMock) GetSMTPConfig(conf interface{}) error {
	if db.ExpSMTPConfErr != nil {
		return db.ExpSMTPConfErr
	}
	rve := reflect.ValueOf(conf).Elem()
	rve.FieldByName("Username").SetString(db.ExpSMTPConf.Username)
	rve.FieldByName("Password").SetString(db.ExpSMTPConf.Password)
	rve.FieldByName("FromEmail").SetString(db.ExpSMTPConf.FromEmail)
	rve.FieldByName("ServerAddress").SetString(db.ExpSMTPConf.ServerAddress)
	rve.FieldByName("TLSPort").SetInt(int64(db.ExpSMTPConf.TLSPort))
	rve.FieldByName("SSLPort").SetInt(int64(db.ExpSMTPConf.SSLPort))
	return nil
}

func (db *DBMock) UpsertSMTPConfig(conf interface{}) error {
	return db.ExpUpsSMTPConfErr
}

func (db *DBMock) GroupByName(string) (*model.Group, error) {
	if db.isInTx {
		return nil, errors.Newf("direct db call while in tx")
	}
	if db.ExpGrpBNm == nil {
		return nil, errors.NewNotFound("not found")
	}
	return db.ExpGrpBNm, db.ExpGrpBNmErr
}

func (db *DBMock) Group(string) (*model.Group, error) {
	if db.isInTx {
		return nil, errors.Newf("direct db call while in tx")
	}
	if db.ExpGrp == nil {
		return nil, errors.NewNotFound("not found")
	}
	return db.ExpGrp, db.ExpGrpErr
}

func (db *DBMock) InsertGroup(name string, acl float32) (*model.Group, error) {
	if db.isInTx {
		return nil, errors.Newf("direct db call while in tx")
	}
	if db.ExpInsGrpErr != nil {
		return nil, db.ExpInsGrpErr
	}
	return &model.Group{ID: currentID(), Name: name, AccessLevel: acl}, db.ExpInsGrpErr
}

func (db *DBMock) AddUserToGroupAtomic(tx *sql.Tx, userID, groupID string) error {
	return db.ExpAddUsrTGrpAtmcErr
}

func (db *DBMock) UserTypeByName(string) (*model.UserType, error) {
	if db.isInTx {
		return nil, errors.Newf("direct db call while in tx")
	}
	if db.ExpUsrTypBNm == nil {
		return nil, errors.NewNotFound("not found")
	}
	return db.ExpUsrTypBNm, db.ExpUsrTypBNmErr
}

func (db *DBMock) InsertUserType(name string) (*model.UserType, error) {
	if db.isInTx {
		return nil, errors.Newf("direct db call while in tx")
	}
	if db.ExpInsUsrTypErr != nil {
		return nil, db.ExpInsUsrTypErr
	}
	return &model.UserType{ID: currentID(), Name: name}, db.ExpInsUsrTypErr
}

func (db *DBMock) InsertUserAtomic(tx *sql.Tx, t model.UserType, password []byte) (*model.User, error) {
	if db.ExpInsUsrAtmErr != nil {
		return nil, db.ExpInsUsrAtmErr
	}
	return &model.User{ID: currentID(), Type: t}, db.ExpInsUsrAtmErr
}

func (db *DBMock) APIKeysByUserID(userID string, offset, count int64) ([]api.Key, error) {
	if db.isInTx {
		return nil, errors.Newf("direct db call while in tx")
	}
	if db.ExpAPIKsBUsrID == nil {
		return nil, errors.NewNotFound("not found")
	}
	return db.ExpAPIKsBUsrID, db.ExpAPIKsBUsrIDErr
}

func (db *DBMock) InsertAPIKey(userID, key string) (*api.Key, error) {
	if db.isInTx {
		return nil, errors.Newf("direct db call while in tx")
	}
	if db.ExpInsAPIKErr != nil {
		return nil, db.ExpInsAPIKErr
	}
	return &api.Key{ID: currentID(), UserID: userID, APIKey: key}, db.ExpInsAPIKErr
}

func (db *DBMock) UpdateUsername(userID, username string) (*model.Username, error) {
	if db.isInTx {
		return nil, errors.Newf("direct db call while in tx")
	}
	if db.ExpUpdUsrNm == nil {
		return nil, errors.NewNotFound("not found")
	}
	return db.ExpUpdUsrNm, db.ExpUpdUsrNmErr
}

func (db *DBMock) UpdateUserPhone(userID, phone string, verified bool) (*model.VerifLogin, error) {
	if db.isInTx {
		return nil, errors.Newf("direct db call while in tx")
	}
	if db.ExpUpdUsrPhn == nil {
		return nil, errors.NewNotFound("not found")
	}
	return db.ExpUpdUsrPhn, db.ExpUpdUsrPhnErr
}

func (db *DBMock) UpdateUserEmail(userID, email string, verified bool) (*model.VerifLogin, error) {
	if db.isInTx {
		return nil, errors.Newf("direct db call while in tx")
	}
	if db.ExpUpdUsrMail == nil {
		return nil, errors.NewNotFound("not found")
	}
	return db.ExpUpdUsrMail, db.ExpUpdUsrMailErr
}

func (db *DBMock) UpdatePassword(userID string, password []byte) error {
	if db.isInTx {
		return errors.Newf("direct db call while in tx")
	}
	return db.ExpupdPassErr
}

func (db *DBMock) UpdatePasswordAtomic(tx *sql.Tx, userID string, password []byte) error {
	return db.ExpupdPassAtmErr
}

func (db *DBMock) UpdateUserPhoneAtomic(tx *sql.Tx, userID, phone string, verified bool) (*model.VerifLogin, error) {
	if db.ExpUpdUsrPhnAtm == nil {
		return nil, errors.NewNotFound("not found")
	}
	return db.ExpUpdUsrPhnAtm, db.ExpUpdUsrPhnAtmErr
}

func (db *DBMock) UpdateUserEmailAtomic(tx *sql.Tx, userID, email string, verified bool) (*model.VerifLogin, error) {
	if db.ExpUpdUsrMailAtm == nil {
		return nil, errors.NewNotFound("not found")
	}
	return db.ExpUpdUsrMailAtm, db.ExpUpdUsrMailAtmErr
}

func (db *DBMock) InsertUserPhone(userID, phone string, verified bool) (*model.VerifLogin, error) {
	if db.isInTx {
		return nil, errors.Newf("direct db call while in tx")
	}
	if db.ExpInsUsrPhnErr != nil {
		return nil, db.ExpInsUsrPhnErr
	}
	return &model.VerifLogin{ID: currentID(), UserID: userID, Address: phone, Verified: verified}, db.ExpInsUsrPhnErr
}

func (db *DBMock) InsertUserEmail(userID, email string, verified bool) (*model.VerifLogin, error) {
	if db.isInTx {
		return nil, errors.Newf("direct db call while in tx")
	}
	if db.ExpInsUsrMailErr != nil {
		return nil, db.ExpInsUsrMailErr
	}
	return &model.VerifLogin{ID: currentID(), UserID: userID, Address: email, Verified: verified}, db.ExpInsUsrMailErr
}

func (db *DBMock) InsertUserName(userID, username string) (*model.Username, error) {
	if db.isInTx {
		return nil, errors.Newf("direct db call while in tx")
	}
	if db.ExpInsUsrNmErr != nil {
		return nil, db.ExpInsUsrNmErr
	}
	return &model.Username{ID: currentID(), UserID: userID, Value: username}, db.ExpInsUsrNmErr
}

func (db *DBMock) InsertUserPhoneAtomic(tx *sql.Tx, userID, phone string, verified bool) (*model.VerifLogin, error) {
	if db.ExpInsUsrPhnAtmErr != nil {
		return nil, db.ExpInsUsrPhnAtmErr
	}
	return &model.VerifLogin{ID: currentID(), UserID: userID, Address: phone, Verified: verified}, db.ExpInsUsrPhnAtmErr
}

func (db *DBMock) InsertUserEmailAtomic(tx *sql.Tx, userID, email string, verified bool) (*model.VerifLogin, error) {
	if db.ExpInsUsrMailAtmErr != nil {
		return nil, db.ExpInsUsrMailAtmErr
	}
	return &model.VerifLogin{ID: currentID(), UserID: userID, Address: email, Verified: verified}, db.ExpInsUsrMailAtmErr
}

func (db *DBMock) InsertUserNameAtomic(tx *sql.Tx, userID, username string) (*model.Username, error) {
	if db.ExpInsUsrNmAtmErr != nil {
		return nil, db.ExpInsUsrNmAtmErr
	}
	return &model.Username{ID: currentID(), UserID: userID, Value: username}, db.ExpInsUsrNmAtmErr
}

func (db *DBMock) InsertUserFbIDAtomic(tx *sql.Tx, userID, fbID string, verified bool) (*model.Facebook, error) {
	if db.ExpInsFbAtmErr != nil {
		return nil, db.ExpInsFbAtmErr
	}
	return &model.Facebook{ID: currentID(), UserID: userID, FacebookID: fbID, Verified: verified}, db.ExpInsFbAtmErr
}

func (db *DBMock) InsertUserDeviceAtomic(tx *sql.Tx, userID, devID string) (*model.Device, error) {
	if db.ExpInsDevAtmErr != nil {
		return nil, db.ExpInsDevAtmErr
	}
	return &model.Device{ID: currentID(), UserID: userID, DeviceID: devID}, db.ExpInsDevAtmErr
}

func (db *DBMock) InsertPhoneToken(userID, phone string, dbt []byte, isUsed bool, expiry time.Time) (*model.DBToken, error) {
	if db.isInTx {
		return nil, errors.Newf("direct db call while in tx")
	}
	if db.ExpInsPhnTknErr != nil {
		return nil, db.ExpInsPhnTknErr
	}
	return &model.DBToken{ID: currentID(), UserID: userID, Address: phone, Token: dbt}, db.ExpInsPhnTknErr
}

func (db *DBMock) InsertEmailToken(userID, email string, dbt []byte, isUsed bool, expiry time.Time) (*model.DBToken, error) {
	if db.isInTx {
		return nil, errors.Newf("direct db call while in tx")
	}
	if db.ExpInsMailTknErr != nil {
		return nil, db.ExpInsMailTknErr
	}
	return &model.DBToken{ID: currentID(), UserID: userID, Address: email, Token: dbt}, db.ExpInsMailTknErr
}

func (db *DBMock) InsertPhoneTokenAtomic(tx *sql.Tx, userID, phone string, dbt []byte, isUsed bool, expiry time.Time) (*model.DBToken, error) {
	if db.ExpInsPhnTknAtmErr != nil {
		return nil, db.ExpInsPhnTknAtmErr
	}
	return &model.DBToken{ID: currentID(), UserID: userID, Address: phone, Token: dbt}, db.ExpInsPhnTknAtmErr
}

func (db *DBMock) InsertEmailTokenAtomic(tx *sql.Tx, userID, email string, dbt []byte, isUsed bool, expiry time.Time) (*model.DBToken, error) {
	if db.ExpInsMailTknAtmErr != nil {
		return nil, db.ExpInsMailTknAtmErr
	}
	return &model.DBToken{ID: currentID(), UserID: userID, Address: email, Token: dbt}, db.ExpInsMailTknAtmErr
}

func (db *DBMock) PhoneTokens(userID string, offset, count int64) ([]model.DBToken, error) {
	if db.isInTx {
		return nil, errors.Newf("direct db call while in tx")
	}
	if len(db.ExpPhnTkns) == 0 {
		return nil, errors.NewNotFound("not found")
	}
	return db.ExpPhnTkns, db.ExpPhnTknsErr
}

func (db *DBMock) EmailTokens(userID string, offset, count int64) ([]model.DBToken, error) {
	if db.isInTx {
		return nil, errors.Newf("direct db call while in tx")
	}
	if len(db.ExpMailTkns) == 0 {
		return nil, errors.NewNotFound("not found")
	}
	return db.ExpMailTkns, db.ExpMailTknsErr
}

func (db *DBMock) User(id string) (*model.User, []byte, error) {
	if db.isInTx {
		return nil, nil, errors.Newf("direct db call while in tx")
	}
	if db.ExpUsr == nil {
		return nil, nil, errors.NewNotFound("not found")
	}
	return db.ExpUsr, db.ExpUsrPass, db.ExpUsrErr
}

func (db *DBMock) UserByPhone(phone string) (*model.User, []byte, error) {
	if db.isInTx {
		return nil, nil, errors.Newf("direct db call while in tx")
	}
	if db.ExpUsrBPhn == nil {
		return nil, nil, errors.NewNotFound("not found")
	}
	return db.ExpUsrBPhn, db.ExpUsrBPhnPass, db.ExpUsrBPhnErr
}

func (db *DBMock) UserByEmail(email string) (*model.User, []byte, error) {
	if db.isInTx {
		return nil, nil, errors.Newf("direct db call while in tx")
	}
	if db.ExpUsrBMail == nil {
		return nil, nil, errors.NewNotFound("not found")
	}
	return db.ExpUsrBMail, db.ExpUsrBMailPass, db.ExpUsrBMailErr
}

func (db *DBMock) UserByUsername(username string) (*model.User, []byte, error) {
	if db.isInTx {
		return nil, nil, errors.Newf("direct db call while in tx")
	}
	if db.ExpUsrBUsrNm == nil {
		return nil, nil, errors.NewNotFound("not found")
	}
	return db.ExpUsrBUsrNm, db.ExpUsrBUsrNmPass, db.ExpUsrBUsrNmErr
}

func (db *DBMock) UserByFacebook(facebookID string) (*model.User, error) {
	if db.isInTx {
		return nil, errors.Newf("direct db call while in tx")
	}
	if db.ExpUsrBFb == nil {
		return nil, errors.NewNotFound("not found")
	}
	return db.ExpUsrBFb, db.ExpUsrBFbErr
}

func (db *DBMock) UserByDeviceID(devID string) (*model.User, []byte, error) {
	if db.isInTx {
		return nil, nil, errors.Newf("direct db call while in tx")
	}
	if db.ExpUsrBDev == nil {
		return nil, nil, errors.NewNotFound("not found")
	}
	return db.ExpUsrBDev, db.ExpUsrBDevPass, db.ExpUsrBDevErr
}
