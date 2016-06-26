package http

import (
	"errors"

	"net/http"

	"io/ioutil"

	"encoding/json"

	"time"

	"bitbucket.org/tomogoma/auth-ms/auth"
	"bitbucket.org/tomogoma/auth-ms/auth/model/history"
	"bitbucket.org/tomogoma/auth-ms/auth/model/user"
	"github.com/gorilla/mux"
)

const (
	loginPath            = "/login"
	internalErrorMessage = "whoops! Something wicked happened"
	regPath              = "/register"
	tokenPath            = "/token"
)

type Logger interface {
	Error(interface{}, ...interface{}) error
	Warn(interface{}, ...interface{}) error
	Info(interface{}, ...interface{})
}

type Request struct {
	FrSrvcID  string `json:"forServiceID,omitempty"`
	RefSrvcID string `json:"refererServiceID,omitempty"`
	DevID     string `json:"devID,omitempty"`
}

type Token struct {
	ID     int       `json:"id,omitempty"`
	UserID int       `json:"userID,omitempty"`
	DevID  string    `json:"devID,omitempty"`
	Token  string    `json:"token,omitempty"`
	Issued time.Time `json:"issueDate,omitempty"`
	Expiry time.Time `json:"expiryDate,omitempty"`
}

type History struct {
	ID            int       `json:"id,omitempty"`
	UserID        int       `json:"userID,omitempty"`
	IpAddress     string    `json:"ipAddress,omitempty"`
	Date          time.Time `json:"date,omitempty"`
	AccessType    string    `json:"accessType,omitempty"`
	SuccessStatus bool      `json:"successStatus"`
}

type User struct {
	Request
	ID         int       `json:"id,omitempty"`
	UName      string    `json:"userName,omitempty"`
	Pass       string    `json:"password,omitempty"`
	FName      string    `json:"firstName,omitempty"`
	MName      string    `json:"middleName,omitempty"`
	LName      string    `json:"lastName,omitempty"`
	Token      *Token    `json:"token,omitempty"`
	PrevLogins []History `json:"prevLogins,omitempty"`
}

func (u *User) UserName() string   { return u.UName }
func (u *User) FirstName() string  { return u.FName }
func (u *User) MiddleName() string { return u.MName }
func (u *User) LastName() string   { return u.LName }

type Server struct {
	auth   *auth.Auth
	lg     Logger
	tIDCh  chan int
	quitCh chan error
}

var ErrorNilAuth = errors.New("Auth cannot be nil")
var ErrorNilLogger = errors.New("Logger cannot be nil")
var ErrorEmptyAddress = errors.New("Address cannot be empty")
var ErrorNilQuitChanel = errors.New("Quit chanel cannot be nil")

func New(auth *auth.Auth, lg Logger, quitCh chan error) (*Server, error) {

	if auth == nil {
		return nil, ErrorNilAuth
	}

	if lg == nil {
		return nil, ErrorNilLogger
	}

	if quitCh == nil {
		return nil, ErrorNilQuitChanel
	}

	tIDCh := make(chan int)
	go transactionSerializer(tIDCh)

	return &Server{auth: auth, lg: lg, tIDCh: tIDCh, quitCh: quitCh}, nil
}

func (s *Server) Start(address string) {

	if address == "" {
		s.quitCh <- ErrorEmptyAddress
		return
	}

	r := mux.NewRouter()
	r.PathPrefix(loginPath).Methods("POST").HandlerFunc(s.handleLogin)
	r.PathPrefix(regPath).Methods("POST").HandlerFunc(s.handleRegistration)
	r.PathPrefix(tokenPath).Methods("POST").HandlerFunc(s.handleToken)

	s.lg.Info("Ready to listen at %s", address)
	s.quitCh <- http.ListenAndServe(address, r)
}

func (s *Server) handleRegistration(w http.ResponseWriter, r *http.Request) {

	tID, dataB, ok := s.readReqBody("register", w, r)
	if !ok {
		return
	}

	req := &User{}
	err := json.Unmarshal(dataB, req)
	if err != nil {
		s.lg.Warn("%d - unmarshal json request body fail: %s", tID, err)
		http.Error(w, "failed to unmarshal json request body", http.StatusBadRequest)
		return
	}

	svdUsr, err := s.auth.RegisterUser(req, req.Pass, r.RemoteAddr, req.FrSrvcID, req.RefSrvcID)
	if err != nil {
		if !auth.AuthError(err) {
			s.lg.Error("%d - registration error: %s", tID, err)
			http.Error(w, internalErrorMessage, http.StatusInternalServerError)
			return
		}
		s.lg.Warn("%d - registration error: %s", tID, err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	respUsr := packageUser(svdUsr)
	b, err := jsonResponse(w, "user", respUsr, http.StatusOK)
	if err != nil {
		s.lg.Error("%d - json response error: %s", tID, err)
		return
	}

	s.lg.Info("%d - [%s] registration complete, wrote %d bytes", tID, r.RemoteAddr, b)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {

	tID, dataB, ok := s.readReqBody("register", w, r)
	if !ok {
		return
	}

	req := &User{}
	err := json.Unmarshal(dataB, req)
	if err != nil {
		s.lg.Warn("%d - unmarshal json request body fail: %s", tID, err)
		http.Error(w, "failed to unmarshal json request body", http.StatusBadRequest)
		return
	}

	authUsr, err := s.auth.Login(req.UName, req.Pass, req.DevID, r.RemoteAddr, req.FrSrvcID, req.RefSrvcID)
	if err != nil {
		if !auth.AuthError(err) {
			s.lg.Error("%d - login error: %s", tID, err)
			http.Error(w, internalErrorMessage, http.StatusInternalServerError)
			return
		}
		s.lg.Warn("%d - login error: %s", tID, err)
		http.Error(w, "invalid userName/password combo", http.StatusUnauthorized)
		return
	}

	respUsr := packageUser(authUsr)
	b, err := jsonResponse(w, "user", respUsr, http.StatusOK)
	if err != nil {
		s.lg.Error("%d - json response error: %s", tID, err)
		return
	}

	s.lg.Info("%d - [%s] login complete, wrote %d bytes", tID, r.RemoteAddr, b)
}

func (s *Server) handleToken(w http.ResponseWriter, r *http.Request) {

	tID, dataB, ok := s.readReqBody("validate token", w, r)
	if !ok {
		return
	}

	req := &User{}
	err := json.Unmarshal(dataB, req)
	if err != nil {
		s.lg.Warn("%d - unmarshal json request body fail: %s", tID, err)
		http.Error(w, "failed to unmarshal json request body", http.StatusBadRequest)
		return
	}

	userID := req.ID
	token := ""
	devID := req.DevID
	if req.Token != nil {
		userID = req.Token.UserID
		devID = req.Token.DevID
		token = req.Token.Token
	}
	authUsr, err := s.auth.AuthenticateToken(userID, devID, token, r.RemoteAddr, req.FrSrvcID, req.RefSrvcID)
	if err != nil {
		if !auth.AuthError(err) {
			s.lg.Error("%d - token authentication error: %s", tID, err)
			http.Error(w, internalErrorMessage, http.StatusInternalServerError)
			return
		}
		s.lg.Warn("%d - token authentication error: %s", tID, err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	respUsr := packageUser(authUsr)
	b, err := jsonResponse(w, "user", respUsr, http.StatusOK)
	if err != nil {
		s.lg.Error("%d - json response error: %s", tID, err)
		return
	}

	s.lg.Info("%d - [%s] token validation complete, wrote %d bytes", tID, r.RemoteAddr, b)
}

func (s *Server) readReqBody(handlerName string, w http.ResponseWriter, r *http.Request) (int, []byte, bool) {

	tID := <-s.tIDCh
	s.lg.Info("%d - [%s] %s request", tID, handlerName, r.RemoteAddr)

	defer r.Body.Close()
	dataB, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.lg.Warn("%d - read request body fail: %s", tID, err)
		http.Error(w, "failed to read request body", http.StatusRequestTimeout)
		return tID, dataB, false
	}

	return tID, dataB, true
}

func packageUser(rcv user.User) User {

	prevLogins := make([]History, len(rcv.PreviousLogins()))
	for i, h := range rcv.PreviousLogins() {
		prevLogins[i] = History{
			ID:            h.ID(),
			UserID:        h.UserID(),
			IpAddress:     h.IPAddress(),
			Date:          h.Date(),
			AccessType:    history.DecodeAccessMethod(h.AccessMethod()),
			SuccessStatus: h.Successful(),
		}
	}

	var token *Token
	rt := rcv.Token()
	if rt != nil {
		token = &Token{
			ID:     rt.ID(),
			UserID: rt.UserID(),
			DevID:  rt.DevID(),
			Token:  rt.Token(),
			Issued: rt.Issued(),
			Expiry: rt.Expiry(),
		}
	}

	return User{
		ID:         rcv.ID(),
		UName:      rcv.UserName(),
		FName:      rcv.FirstName(),
		MName:      rcv.MiddleName(),
		LName:      rcv.LastName(),
		PrevLogins: prevLogins,
		Token:      token,
	}
}

func jsonResponse(w http.ResponseWriter, dataKey string, data interface{}, status int) (int, error) {

	respM := map[string]interface{}{"status": status}
	respM[dataKey] = data

	respB, err := json.Marshal(respM)
	if err != nil {
		return 0, err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return w.Write(respB)
}

func transactionSerializer(tIDCh chan int) {

	tID := 0
	for {
		tID++
		tIDCh <- tID
	}
}
