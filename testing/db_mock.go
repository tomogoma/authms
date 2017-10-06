package testing

import (
	"database/sql"
	"time"

	"github.com/tomogoma/authms/model"
	"github.com/tomogoma/go-commons/errors"
)

type DBMock struct {
	errors.NotFoundErrCheck

	ExpInsGrp    *model.Group
	ExpInsGrpErr error
	ExpGrpBNm    *model.Group
	ExpGrpBNmErr error
	ExpGrp       *model.Group
	ExpGrpErr    error

	ExpInsUsrTyp    *model.UserType
	ExpInsUsrTypErr error
	ExpUsrTypBNm    *model.UserType
	ExpUsrTypBNmErr error

	ExpInsUsr       *model.User
	ExpInsUsrErr    error
	ExpInsUsrAtm    *model.User
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

	ExpInsAPIK        *model.APIKey
	ExpInsAPIKErr     error
	ExpAPIKsBUsrID    []model.APIKey
	ExpAPIKsBUsrIDErr error

	ExpAddUsrTGrpAtmcErr error

	ExpInsDevAtm    *model.Device
	ExpInsDevAtmErr error

	ExpInsUsrNm       *model.Username
	ExpInsUsrNmErr    error
	ExpInsUsrNmAtm    *model.Username
	ExpInsUsrNmAtmErr error
	ExpUpdUsrNm       *model.Username
	ExpUpdUsrNmErr    error

	ExpInsUsrPhn       *model.VerifLogin
	ExpInsUsrPhnErr    error
	ExpInsUsrPhnAtm    *model.VerifLogin
	ExpInsUsrPhnAtmErr error
	ExpUpdUsrPhn       *model.VerifLogin
	ExpUpdUsrPhnErr    error
	ExpUpdUsrPhnAtm    *model.VerifLogin
	ExpUpdUsrPhnAtmErr error

	ExpInsPhnTkn       *model.DBToken
	ExpInsPhnTknErr    error
	ExpInsPhnTknAtm    *model.DBToken
	ExpInsPhnTknAtmErr error
	ExpPhnTkns         []model.DBToken
	ExpPhnTknsErr      error

	ExpInsUsrMail       *model.VerifLogin
	ExpInsUsrMailErr    error
	ExpInsUsrMailAtm    *model.VerifLogin
	ExpInsUsrMailAtmErr error
	ExpUpdUsrMail       *model.VerifLogin
	ExpUpdUsrMailErr    error
	ExpUpdUsrMailAtm    *model.VerifLogin
	ExpUpdUsrMailAtmErr error

	ExpInsMailTkn       *model.DBToken
	ExpInsMailTknErr    error
	ExpInsMailTknAtm    *model.DBToken
	ExpInsMailTknAtmErr error
	ExpMailTkns         []model.DBToken
	ExpMailTknsErr      error

	ExpInsFbAtm    *model.Facebook
	ExpInsFbAtmErr error
}

func (db *DBMock) ExecuteTx(fn func(*sql.Tx) error) error {
	return fn(new(sql.Tx))
}

func (db *DBMock) GroupByName(string) (*model.Group, error) {
	return db.ExpGrpBNm, db.ExpGrpBNmErr
}
func (db *DBMock) Group(string) (*model.Group, error) {
	return db.ExpGrp, db.ExpGrpErr
}
func (db *DBMock) InsertGroup(name string, acl float32) (*model.Group, error) {
	return db.ExpInsGrp, db.ExpInsGrpErr
}
func (db *DBMock) AddUserToGroupAtomic(tx *sql.Tx, userID, groupID string) error {
	return db.ExpAddUsrTGrpAtmcErr
}

func (db *DBMock) UserTypeByName(string) (*model.UserType, error) {
	return db.ExpUsrTypBNm, db.ExpUsrTypBNmErr
}
func (db *DBMock) InsertUserType(name string) (*model.UserType, error) {
	return db.ExpInsUsrTyp, db.ExpInsUsrTypErr
}
func (db *DBMock) InsertUserAtomic(tx *sql.Tx, t model.UserType, password []byte) (*model.User, error) {
	return db.ExpInsUsrAtm, db.ExpInsUsrAtmErr
}

func (db *DBMock) APIKeysByUserID(userID string, offset, count int64) ([]model.APIKey, error) {
	return db.ExpAPIKsBUsrID, db.ExpAPIKsBUsrIDErr
}
func (db *DBMock) InsertAPIKey(userID, key string) (*model.APIKey, error) {
	return db.ExpInsAPIK, db.ExpInsAPIKErr
}

func (db *DBMock) UpdateUsername(userID, username string) (*model.Username, error) {
	return db.ExpUpdUsrNm, db.ExpUpdUsrNmErr
}
func (db *DBMock) UpdateUserPhone(userID, phone string, verified bool) (*model.VerifLogin, error) {
	return db.ExpUpdUsrPhn, db.ExpUpdUsrPhnErr
}
func (db *DBMock) UpdateUserEmail(userID, email string, verified bool) (*model.VerifLogin, error) {
	return db.ExpUpdUsrMail, db.ExpUpdUsrMailErr
}

func (db *DBMock) UpdatePassword(userID string, password []byte) error {
	return db.ExpupdPassErr
}
func (db *DBMock) UpdatePasswordAtomic(tx *sql.Tx, userID string, password []byte) error {
	return db.ExpupdPassAtmErr
}
func (db *DBMock) UpdateUserPhoneAtomic(tx *sql.Tx, userID, phone string, verified bool) (*model.VerifLogin, error) {
	return db.ExpUpdUsrPhnAtm, db.ExpUpdUsrPhnAtmErr
}
func (db *DBMock) UpdateUserEmailAtomic(tx *sql.Tx, userID, email string, verified bool) (*model.VerifLogin, error) {
	return db.ExpUpdUsrMailAtm, db.ExpUpdUsrMailAtmErr
}

func (db *DBMock) InsertUserPhone(userID, phone string, verified bool) (*model.VerifLogin, error) {
	return db.ExpInsUsrPhn, db.ExpInsUsrPhnErr
}
func (db *DBMock) InsertUserEmail(userID, email string, verified bool) (*model.VerifLogin, error) {
	return db.ExpInsUsrMail, db.ExpInsUsrMailErr
}
func (db *DBMock) InsertUserName(userID, username string) (*model.Username, error) {
	return db.ExpInsUsrNm, db.ExpInsUsrNmErr
}
func (db *DBMock) InsertUserPhoneAtomic(tx *sql.Tx, userID, phone string, verified bool) (*model.VerifLogin, error) {
	return db.ExpInsUsrPhnAtm, db.ExpInsUsrPhnAtmErr
}
func (db *DBMock) InsertUserEmailAtomic(tx *sql.Tx, userID, email string, verified bool) (*model.VerifLogin, error) {
	return db.ExpInsUsrMailAtm, db.ExpInsUsrMailAtmErr
}
func (db *DBMock) InsertUserNameAtomic(tx *sql.Tx, userID, username string) (*model.Username, error) {
	return db.ExpInsUsrNmAtm, db.ExpInsUsrNmAtmErr
}
func (db *DBMock) InsertUserFbIDAtomic(tx *sql.Tx, userID, fbID string, verified bool) (*model.Facebook, error) {
	return db.ExpInsFbAtm, db.ExpInsFbAtmErr
}
func (db *DBMock) InsertUserDeviceAtomic(tx *sql.Tx, userID, devID string) (*model.Device, error) {
	return db.ExpInsDevAtm, db.ExpInsDevAtmErr
}

func (db *DBMock) InsertPhoneToken(userID, phone string, dbt []byte, isUsed bool, expiry time.Time) (*model.DBToken, error) {
	return db.ExpInsPhnTkn, db.ExpInsPhnTknErr
}
func (db *DBMock) InsertEmailToken(userID, email string, dbt []byte, isUsed bool, expiry time.Time) (*model.DBToken, error) {
	return db.ExpInsMailTkn, db.ExpInsMailTknErr
}
func (db *DBMock) InsertPhoneTokenAtomic(tx *sql.Tx, userID, phone string, dbt []byte, isUsed bool, expiry time.Time) (*model.DBToken, error) {
	return db.ExpInsPhnTknAtm, db.ExpInsPhnTknAtmErr
}
func (db *DBMock) InsertEmailTokenAtomic(tx *sql.Tx, userID, email string, dbt []byte, isUsed bool, expiry time.Time) (*model.DBToken, error) {
	return db.ExpInsMailTknAtm, db.ExpInsMailTknAtmErr
}
func (db *DBMock) PhoneTokens(userID string, offset, count int64) ([]model.DBToken, error) {
	return db.ExpPhnTkns, db.ExpPhnTknsErr
}
func (db *DBMock) EmailTokens(userID string, offset, count int64) ([]model.DBToken, error) {
	return db.ExpMailTkns, db.ExpMailTknsErr
}

func (db *DBMock) User(id string) (*model.User, []byte, error) {
	return db.ExpUsr, db.ExpUsrPass, db.ExpUsrErr
}
func (db *DBMock) UserByPhone(phone string) (*model.User, []byte, error) {
	return db.ExpUsrBPhn, db.ExpUsrBPhnPass, db.ExpUsrBPhnErr
}
func (db *DBMock) UserByEmail(email string) (*model.User, []byte, error) {
	return db.ExpUsrBMail, db.ExpUsrBMailPass, db.ExpUsrBMailErr
}
func (db *DBMock) UserByUsername(username string) (*model.User, []byte, error) {
	return db.ExpUsrBUsrNm, db.ExpUsrBUsrNmPass, db.ExpUsrBUsrNmErr
}
func (db *DBMock) UserByFacebook(facebookID string) (*model.User, error) {
	return db.ExpUsrBFb, db.ExpUsrBFbErr
}
func (db *DBMock) UserByDeviceID(devID string) (*model.User, []byte, error) {
	return db.ExpUsrBDev, db.ExpUsrBDevPass, db.ExpUsrBDevErr
}
