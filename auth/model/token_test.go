package model_test

//var quitCh chan error

//func TestModel_Save_n_Get(t *testing.T) {

//m := newModel(t)
//defer testhelper.TearDown(db, t)
//
//tkn, err := token.New(expToken.explodeParams())
//if err != nil {
//	t.Fatalf("token.New(): %s", err)
//}
//
//i, err := m.Save(*tkn)
//if err != nil {
//	t.Fatalf("tokenModel.Save(): %s", err)
//}
//
//if i < 1 {
//	t.Errorf("Expected token id > 1 but got %d", i)
//}
//
//act, err := m.Get(tkn.UserID(), tkn.DevID(), tkn.Token())
//if err != nil {
//	t.Fatalf("tokenModel.Get(): %s", err)
//}
//
//compareToken(act, expToken, t)
//}

//func TestModel_Get_notExist(t *testing.T) {

//m := newModel(t)
//defer testhelper.TearDown(db, t)
//
//tkn, err := token.New(expToken.explodeParams())
//if err != nil {
//	t.Fatalf("token.New(): %s", err)
//}
//
//_, err = m.Save(*tkn)
//if err != nil {
//	t.Fatalf("tokenModel.Save(): %s", err)
//}
//
//act, err := m.Get(578013, tkn.DevID(), "some-nonexist-token")
//if err == nil || err != token.ErrorInvalidToken {
//	t.Fatalf("Expected error %s but got %v", token.ErrorInvalidToken, err)
//}
//
//if act != nil {
//	t.Fatalf("Expected nil result, got %v", act)
//}
//}

//func TestModel_Get_userIDNotExist(t *testing.T) {

//m := newModel(t)
//defer testhelper.TearDown(db, t)
//
//tkn, err := token.New(expToken.explodeParams())
//if err != nil {
//	t.Fatalf("token.New(): %s", err)
//}
//
//_, err = m.Save(*tkn)
//if err != nil {
//	t.Fatalf("tokenModel.Save(): %s", err)
//}
//
//act, err := m.Get(578013, tkn.DevID(), tkn.Token())
//if err == nil || err != token.ErrorInvalidToken {
//	t.Fatalf("Expected error %s but got %v", token.ErrorInvalidToken, err)
//}
//
//if act != nil {
//	t.Fatalf("Expected nil result, got %v", act)
//}
//}

//func TestModel_Get_tokenNotExist(t *testing.T) {

//m := newModel(t)
//defer testhelper.TearDown(db, t)
//
//tkn, err := token.New(expToken.explodeParams())
//if err != nil {
//	t.Fatalf("token.New(): %s", err)
//}
//
//_, err = m.Save(*tkn)
//if err != nil {
//	t.Fatalf("tokenModel.Save(): %s", err)
//}
//
//act, err := m.Get(tkn.UserID(), tkn.DevID(), "some-noexist-token")
//if err == nil || err != token.ErrorInvalidToken {
//	t.Fatalf("Expected error %s but got %v", token.ErrorInvalidToken, err)
//}
//
//if act != nil {
//	t.Fatalf("Expected nil result, got %v", act)
//}
//}

//func TestModel_Get_NoResults(t *testing.T) {

//m := newModel(t)
//defer testhelper.TearDown(db, t)
//
//tkn, err := token.New(expToken.explodeParams())
//if err != nil {
//	t.Fatalf("token.New(): %s", err)
//}
//
//act, err := m.Get(tkn.UserID(), tkn.DevID(), tkn.Token())
//if err == nil || err != token.ErrorInvalidToken {
//	t.Fatalf("Expected error %s but got %v", token.ErrorInvalidToken, err)
//}
//
//if act != nil {
//	t.Fatalf("Expected nil result, got %v", act)
//}
//}

//func TestModel_Get_emptyUserID(t *testing.T) {

//m := newModel(t)
//defer testhelper.TearDown(db, t)
//
//tkn, err := token.New(expToken.explodeParams())
//if err != nil {
//	t.Fatalf("token.New(): %s", err)
//}
//
//act, err := m.Get(0, tkn.DevID(), tkn.Token())
//if err == nil || err != token.ErrorInvalidToken {
//	t.Fatalf("Expected error %s but got %v", token.ErrorInvalidToken, err)
//}
//
//if act != nil {
//	t.Fatalf("Expected nil result, got %v", act)
//}
//}

//func TestModel_Get_emptyToken(t *testing.T) {

//m := newModel(t)
//defer testhelper.TearDown(db, t)
//
//tkn, err := token.New(expToken.explodeParams())
//if err != nil {
//	t.Fatalf("token.New(): %s", err)
//}
//
//act, err := m.Get(tkn.UserID(), tkn.DevID(), "")
//if err == nil || err != token.ErrorInvalidToken {
//	t.Fatalf("Expected error %s but got %v", token.ErrorInvalidToken, err)
//}
//
//if act != nil {
//	t.Fatalf("Expected nil result, got %v", act)
//}
//}

//func TestModel_Delete(t *testing.T) {

//m := newModel(t)
//defer testhelper.TearDown(db, t)
//
//tknToDel, err := token.New(expToken.explodeParams())
//tknToLeaveUserID := tknToDel.UserID() - 1
//testhelper.InsertDummyUser(db, tknToLeaveUserID, t)
//tknToLeave, err := token.New(tknToLeaveUserID, tknToDel.DevID(),
//	tknToDel.Token() + "ab", tknToDel.Issued(), tknToDel.Expiry())
//if err != nil {
//	t.Fatalf("token.New(): %s", err)
//}
//
//_, err = m.Save(*tknToLeave)
//if err != nil {
//	t.Fatalf("tokenModel.Save(): %s", err)
//}
//
//_, err = m.Save(*tknToDel)
//if err != nil {
//	t.Fatalf("tokenModel.Save(): %s", err)
//}
//
//dltd, err := m.Delete(tknToDel.Token())
//if err != nil {
//	t.Fatalf("tokenModel.Delete(): %s", err)
//}
//
//if !dltd {
//	t.Fatalf("Expected deleted to be true got %v", dltd)
//}
//
//act, err := m.Get(tknToDel.UserID(), tknToDel.DevID(), tknToDel.Token())
//if err == nil || err != token.ErrorInvalidToken {
//	t.Fatalf("tokenModel.Get(): Expected error %s but got %v", token.ErrorInvalidToken, err)
//}
//
//if act != nil {
//	t.Errorf("Expected token to be deleted but in db as %+v", act)
//}
//
//act, err = m.Get(tknToLeave.UserID(), tknToLeave.DevID(), tknToLeave.Token())
//if err != nil {
//	t.Fatalf("tokenModel.Get(): %s", err)
//}
//
//if act == nil {
//	t.Fatal("Token not expected to be deleted was deleted", act)
//}
//}

//func TestModel_Delete_notExist(t *testing.T) {

//m := newModel(t)
//defer testhelper.TearDown(db, t)
//
//tknToLeave, err := token.New(expToken.explodeParams())
//if err != nil {
//	t.Fatalf("token.New(): %s", err)
//}
//
//_, err = m.Save(*tknToLeave)
//if err != nil {
//	t.Fatalf("tokenModel.Save(): %s", err)
//}
//
//dltd, err := m.Delete(expToken.token + "abd")
//if err != nil {
//	t.Fatalf("tokenModel.Delete(): %s", err)
//}
//
//if dltd {
//	t.Errorf("Expected non exist not to be deleted (false) but got %v", dltd)
//}
//
//act, err := m.Get(tknToLeave.UserID(), tknToLeave.DevID(), tknToLeave.Token())
//if err != nil {
//	t.Fatalf("tokenModel.Get(): %s", err)
//}
//
//if act == nil {
//	t.Fatal("Token not expected to be deleted was deleted", act)
//}
//}

//func TestModel_ValidateExpiry(t *testing.T) {

//m := newModel(t)
//defer testhelper.TearDown(db, t)
//
//tkn, err := token.New(expToken.explodeParams())
//if err != nil {
//	t.Fatalf("token.New(): %s", err)
//}
//
//vTkn, err := m.ValidateExpiry(tkn)
//if err != nil {
//	t.Fatalf("tokenModel.ValidateExpiry(): %s", err)
//}
//
//compareToken(vTkn, expToken, t)
//}

//func TestModel_ValidateExpiry_expiredToken_DeleteError(t *testing.T) {

//m := newModel(t)
//defer testhelper.TearDown(db, t)
//
//expT := expToken
//expT.expAdd = -8 * time.Hour
//expT.token = ""
//vTkn, err := m.ValidateExpiry(expT)
//if err == nil || !strings.Contains(err.Error(), token.ErrorEmptyToken.Error()) {
//	t.Errorf("expected error to contain %s but got %v", token.ErrorEmptyToken, err)
//}
//
//if vTkn != nil {
//	t.Errorf("Expected returned invalid token to be nil, but got %v", vTkn)
//}
//}

//func TestModel_ValidateExpiry_expired(t *testing.T) {

//m := newModel(t)
//defer testhelper.TearDown(db, t)
//
//expT := expToken
//
//tkn, err := token.New(expT.explodeParams())
//if err != nil {
//	t.Fatalf("token.New(): %s", err)
//}
//
//id, err := m.Save(*tkn)
//if err != nil {
//	t.Fatalf("tokenModel.Save(): %s", err)
//}
//
//expT.id = id
//expT.token = tkn.Token()
//expT.expAdd = -8 * time.Hour
//
//vTkn, err := m.ValidateExpiry(expT)
//if err == nil || err != token.ErrorExpiredToken {
//	t.Fatalf("Expected error %s but got %v", token.ErrorExpiredToken, err)
//}
//
//if vTkn != nil {
//	t.Errorf("Expect return invalid token to be nil, got %v", vTkn)
//}
//
//dbTkn, err := m.Get(tkn.UserID(), tkn.DevID(), tkn.Token())
//if err == nil || err != token.ErrorInvalidToken {
//	t.Fatalf("tokenModel.Get(): Expected error %s but got %v", token.ErrorInvalidToken, err)
//}
//
//if dbTkn != nil {
//	t.Errorf("Expected invalid token to be deleted, found %+v", dbTkn)
//}
//}
//
//func TestModel_GetSmallestExpiry(t *testing.T) {
//
//	m := newModel(t)
//	defer testhelper.TearDown(db, t)
//
//	medUsrID := 5
//	lrgstUsrID := 6
//	testhelper.InsertDummyUser(db, medUsrID, t)
//	testhelper.InsertDummyUser(db, lrgstUsrID, t)
//
//	med, err := token.New(medUsrID, "dev1id", "some-token1", time.Now(), time.Now().Add(medExpTime))
//	if err != nil {
//		t.Fatalf("token.New(): %s", err)
//	}
//	lrgst, err := token.New(lrgstUsrID, "dev2id", "some-other-token2", time.Now(), time.Now().Add(longExpTime))
//	if err != nil {
//		t.Fatalf("token.New(): %s", err)
//	}
//	smlst, err := token.New(expToken.explodeParams())
//	if err != nil {
//		t.Fatalf("token.New(): %s", err)
//	}
//
//	_, err = m.Save(*med)
//	if err != nil {
//		t.Fatalf("tokenModel.Save(): %s", err)
//	}
//	_, err = m.Save(*smlst)
//	if err != nil {
//		t.Fatalf("tokenModel.Save(): %s", err)
//	}
//	_, err = m.Save(*lrgst)
//	if err != nil {
//		t.Fatalf("tokenModel.Save(): %s", err)
//	}
//
//	tkn, err := m.GetSmallestExpiry()
//	if err != nil {
//		t.Fatalf("tokenModel.GetSmallestExpiry(): %s", err)
//	}
//
//	compareToken(tkn, expToken, t)
//}
//
//func TestModel_GetSmallestExpiry_noRecords(t *testing.T) {
//
//	m := newModel(t)
//	defer testhelper.TearDown(db, t)
//
//	tkn, err := m.GetSmallestExpiry()
//	if err != nil {
//		t.Fatalf("tokenModel.GetSmallestExpiry(): %s", err)
//	}
//
//	if tkn != nil {
//		t.Errorf("Expected nil token but got %v", tkn)
//	}
//}
//
//func TestModel_RunGarbageCollector_nilQuitCh(t *testing.T) {
//
//	m := newModel(t)
//	defer testhelper.TearDown(db, t)
//	lg := log4go.NewDefaultLogger(log4go.FINEST)
//
//	err := m.RunGarbageCollector(nil, lg)
//	if err == nil || err != token.ErrorNilQuitChanel {
//		t.Fatalf("Expected error %s but got %v", token.ErrorNilQuitChanel, err)
//	}
//
//	timer := time.NewTimer(1 * time.Second)
//	select {
//	case <-timer.C:
//		break
//	case err := <-quitCh:
//		t.Fatalf("Garbage collector quit with error %v", err)
//	}
//}
//
//func TestModel_RunGarbageCollector_nilLogger(t *testing.T) {
//
//	m := newModel(t)
//	defer testhelper.TearDown(db, t)
//
//	quitCh := make(chan error)
//	err := m.RunGarbageCollector(quitCh, nil)
//	if err == nil || err != token.ErrorNilLogger {
//		t.Fatalf("Expected error %s but got %v", token.ErrorNilLogger, err)
//	}
//
//	timer := time.NewTimer(1 * time.Second)
//	select {
//	case <-timer.C:
//		break
//	case err := <-quitCh:
//		t.Fatalf("Garbage collector quit with error %v", err)
//	}
//}
//
//func TestModel_RunGarbageCollector_singleCollector(t *testing.T) {
//
//	m := newModel(t)
//	defer testhelper.TearDown(db, t)
//	lg := log4go.NewDefaultLogger(log4go.FINEST)
//
//	quitCh1 := make(chan error)
//	err := m.RunGarbageCollector(quitCh1, lg)
//	if err != nil {
//		t.Fatalf("tokenModel.RunGarbageCollector(): %s", err)
//	}
//
//	quitCh2 := make(chan error)
//	err = m.RunGarbageCollector(quitCh2, lg)
//	if err != nil {
//		t.Fatalf("tokenModel.RunGarbageCollector(): %s", err)
//	}
//
//	timer := time.NewTimer(1 * time.Second)
//	select {
//	case <-timer.C:
//		t.Fatal("Garbage collector took too long to report error")
//	case err := <-quitCh1:
//		t.Fatalf("First garbage collector quit with error %v", err)
//	case err := <-quitCh2:
//		if err == nil || err != token.ErrorGCRunning {
//			t.Fatalf("Expected error %s but got %v", token.ErrorGCRunning, err)
//		}
//	}
//}
//
//func newModel(t *testing.T) *token.Model {
//
//	db = testhelper.SQLDB(t)
//	m, err := token.NewModel(db)
//	if err != nil {
//		testhelper.TearDown(db, t)
//		t.Fatalf("token.NewModel(): %s", err)
//	}
//	testhelper.InsertDummyUser(db, userID, t)
//	return m
//}
