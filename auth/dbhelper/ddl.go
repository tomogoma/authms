package dbhelper

const (
	users = `
CREATE TABLE IF NOT EXISTS users (
  id         SERIAL PRIMARY KEY NOT NULL,
  password   BYTES              NOT NULL,
  createDate TIMESTAMP          NOT NULL,
  updateDate TIMESTAMP          NOT NULL DEFAULT CURRENT_TIMESTAMP()
);
`
	usernames = `
CREATE TABLE IF NOT EXISTS userNames (
  userID     SERIAL             NOT NULL REFERENCES users (id),
  id         SERIAL		NOT NULL,
  userName   STRING UNIQUE      NOT NULL,
  createDate TIMESTAMP          NOT NULL,
  updateDate TIMESTAMP          NOT NULL DEFAULT CURRENT_TIMESTAMP(),
  PRIMARY KEY	(userID, id)
);
`
	emails = `
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
	phones = `
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
	appUserIDs = `
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
	// TODO enforce that devID, userID [, and appID??] as unique
	tokens = `
CREATE TABLE IF NOT EXISTS tokens (
  userID SERIAL             NOT NULL REFERENCES users (id),
  id     SERIAL		    NOT NULL,
  devID  STRING             NOT NULL,
  token  STRING UNIQUE      NOT NULL,
  issued TIMESTAMP          NOT NULL,
  expiry TIMESTAMP          NOT NULL,
  PRIMARY KEY	(userID, id)
);
`
	// TODO add error column
	history = `
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
)
