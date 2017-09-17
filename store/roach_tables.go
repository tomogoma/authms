package store

const (
	UsersTBLDesc = `
CREATE TABLE IF NOT EXISTS users (
  id         SERIAL PRIMARY KEY NOT NULL,
  password   BYTES              NOT NULL,
  createDate TIMESTAMP          NOT NULL,
  updateDate TIMESTAMP          NOT NULL DEFAULT CURRENT_TIMESTAMP()
);
`
	UsernamesTblDesc = `
CREATE TABLE IF NOT EXISTS userNames (
  userID     SERIAL             NOT NULL REFERENCES users (id),
  id         SERIAL		NOT NULL,
  userName   STRING UNIQUE      NOT NULL,
  createDate TIMESTAMP          NOT NULL,
  updateDate TIMESTAMP          NOT NULL DEFAULT CURRENT_TIMESTAMP(),
  PRIMARY KEY	(userID, id)
);
`
	EmailsTblDesc = `
CREATE TABLE IF NOT EXISTS emails (
  userID     SERIAL             NOT NULL REFERENCES users (id),
  id         SERIAL		NOT NULL,
  email      STRING UNIQUE      NOT NULL,
  validated  BOOL               NOT NULL DEFAULT FALSE,
  createDate TIMESTAMP          NOT NULL,
  updateDate TIMESTAMP          NOT NULL DEFAULT CURRENT_TIMESTAMP(),
  PRIMARY KEY	(userID, id)
);
`
	PhonesTblDesc = `
CREATE TABLE IF NOT EXISTS phones (
  userID     SERIAL             NOT NULL REFERENCES users (id),
  id         SERIAL		NOT NULL,
  phone      STRING UNIQUE      NOT NULL,
  validated  BOOL               NOT NULL DEFAULT FALSE,
  createDate TIMESTAMP          NOT NULL,
  updateDate TIMESTAMP          NOT NULL DEFAULT CURRENT_TIMESTAMP(),
  PRIMARY KEY	(userID, id)
);
`
	AppUserIDsTblDesc = `
CREATE TABLE IF NOT EXISTS appUserIDs (
  userID     SERIAL             NOT NULL REFERENCES users (id),
  id         SERIAL		NOT NULL,
  appUserID  STRING             NOT NULL,
  appName    STRING             NOT NULL,
  validated  BOOL               NOT NULL DEFAULT FALSE,
  createDate TIMESTAMP          NOT NULL,
  updateDate TIMESTAMP          NOT NULL DEFAULT CURRENT_TIMESTAMP(),
  PRIMARY KEY	(userID, id),
  UNIQUE     (userID, appName),
  UNIQUE     (appName, appUserID)
);
`
	// TODO add error column
	HistoryTblDesc = `
CREATE TABLE IF NOT EXISTS history (
  userID       SERIAL             NOT NULL REFERENCES users (id),
  id           SERIAL		  NOT NULL,
  date         TIMESTAMP          NOT NULL,
  accessMethod STRING             NOT NULL,
  successful   BOOL               NOT NULL,
  ipAddress    STRING,
  devID        STRING,
  PRIMARY KEY	(userID, id),
  INDEX         history_UserDate_indx (userID, DATE )
);
`
	AuthVerificationsTblDesc = `
CREATE TABLE IF NOT EXISTS authVerifications (
	id				CHAR		PRIMARY KEY NOT NULL,
	type			CHAR		NOT NULL,
	subjectValue	CHAR		NOT NULL,
	userID			INTEGER		NOT NULL REFERENCES users(id),
	codeHash		BYTES		NOT NULL,
	isUsed			BOOL		NOT NULL,
	issueDate		TIMESTAMPTZ	NOT NULL,
	expiryDate		TIMESTAMPTZ	NOT NULL,
	INDEX authVerifications_type_indx(type, userID)
);
`
)

var AllTableDescs = []string{UsersTBLDesc, UsernamesTblDesc,
	EmailsTblDesc, PhonesTblDesc, AppUserIDsTblDesc, HistoryTblDesc,
	AuthVerificationsTblDesc}
