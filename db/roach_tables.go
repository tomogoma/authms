package db

const (
	Version = 1

	TblConfigurations = "configurations"
	TblUserTypes      = "userTypes"
	TblUsers          = "users"
	TblGroups         = "groups"
	TblUserGroupsJoin = "userGroupsJoin"
	TblAPIKeys        = "apiKeys"
	TblUserNameIDs    = "userNameIDs"
	TblEmailIDs       = "emailIDs"
	TblEmailTokens    = "emailTokens"
	TblPhoneIDs       = "phoneIDs"
	TblPhoneTokens    = "phoneTokens"
	TblFacebookIDs    = "facebookIDs"

	ColID          = "ID"
	ColPassword    = "password"
	ColCreateDate  = "createDate"
	ColUpdateDate  = "updateDate"
	ColName        = "name"
	ColAccessLevel = "accessLevel"
	ColUserID      = "userID"
	ColGroupID     = "groupID"
	ColUserName    = "username"
	ColEmail       = "email"
	ColPhone       = "phone"
	ColVerified    = "verified"
	ColFacebookID  = "facebookID"
	ColToken       = "token"
	ColIsUsed      = "isUsed"
	ColIssueDate   = "issueDate"
	ColExpiryDate  = "expiryDate"
	ColKey         = "key"
	ColValue       = "value"
	ColAPIKey      = "apiKey"
	ColTypeID      = "typeID"
	ColDescription = "description"

	TblDescConfigurations = `
	CREATE TABLE IF NOT EXISTS ` + TblConfigurations + ` (
		` + ColID + ` CHAR PRIMARY KEY NOT NULL CHECK (` + ColID + ` != ''),
		` + ColKey + ` CHAR UNIQUE NOT NULL CHECK (` + ColKey + ` != ''),
		` + ColValue + ` CHAR NOT NULL CHECK (` + ColValue + ` != ''),
		` + ColCreateDate + ` TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP(),
		` + ColUpdateDate + ` TIMESTAMPTZ NOT NULL
	);
	`
	TblDescUserTypes = `
	CREATE TABLE IF NOT EXISTS ` + TblUserTypes + ` (
		` + ColID + ` CHAR PRIMARY KEY NOT NULL CHECK (` + ColID + ` != ''),
		` + ColName + ` CHAR UNIQUE NOT NULL CHECK (` + ColName + ` != ''),
		` + ColCreateDate + ` TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP(),
		` + ColUpdateDate + ` TIMESTAMPTZ NOT NULL
	);
	`
	TblDescUsers = `
	CREATE TABLE IF NOT EXISTS ` + TblUsers + ` (
		` + ColID + ` CHAR PRIMARY KEY NOT NULL CHECK (` + ColID + ` != ''),
		` + ColTypeID + ` CHAR NOT NULL REFERENCES ` + TblUsers + ` (` + ColUserID + `),
		` + ColPassword + ` BYTEA NOT NULL CHECK ( LENGTH(` + ColPassword + `) >= 8 ),
		` + ColCreateDate + ` TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP(),
		` + ColUpdateDate + ` TIMESTAMPTZ NOT NULL
	);
	`
	TblDescGroups = `
	CREATE TABLE IF NOT EXISTS ` + TblGroups + ` (
		` + ColID + ` CHAR PRIMARY KEY NOT NULL CHECK (` + ColID + ` != ''),
		` + ColName + ` CHAR UNIQUE NOT NULL CHECK (` + ColName + ` != ''),
		` + ColAccessLevel + ` SMALLINT UNIQUE NOT NULL CHECK (` + ColAccessLevel + ` BETWEEN 0 AND 10),
		` + ColCreateDate + ` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(),
		` + ColUpdateDate + ` TIMESTAMP NOT NULL
	);
	`
	TblDescUsersGroupsJoin = `
	CREATE TABLE IF NOT EXISTS ` + TblUserGroupsJoin + ` (
		` + ColID + ` CHAR PRIMARY KEY NOT NULL CHECK (` + ColID + ` != ''),
		` + ColUserID + ` CHAR UNIQUE NOT NULL REFERENCES ` + TblUsers + ` (` + ColUserID + `),
		` + ColGroupID + ` CHAR NOT NULL REFERENCES ` + TblGroups + ` (` + ColGroupID + `),
		` + ColCreateDate + ` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(),
		` + ColUpdateDate + ` TIMESTAMP NOT NULL
	);
	`
	TblDescAPIKeys = `
	CREATE TABLE IF NOT EXISTS ` + TblAPIKeys + ` (
		` + ColID + ` CHAR PRIMARY KEY NOT NULL CHECK (` + ColID + ` != ''),
		` + ColUserID + ` CHAR NOT NULL REFERENCES ` + TblUsers + ` (` + ColUserID + `),
		` + ColAPIKey + ` BYTEA NOT NULL CHECK ( LENGTH(` + ColAPIKey + `) >= 56 ),
		` + ColCreateDate + ` TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP(),
		` + ColUpdateDate + ` TIMESTAMPTZ NOT NULL
	);
	`
	TblDescUserNameIDs = `
	CREATE TABLE IF NOT EXISTS ` + TblUserNameIDs + ` (
		` + ColID + ` CHAR PRIMARY KEY NOT NULL CHECK (` + ColID + ` != ''),
		` + ColUserID + ` CHAR UNIQUE NOT NULL REFERENCES ` + TblUsers + ` (` + ColUserID + `),
		` + ColUserName + ` STRING UNIQUE NOT NULL,
		` + ColCreateDate + ` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(),
		` + ColUpdateDate + ` TIMESTAMP NOT NULL,
	);
	`
	TblDescEmailIDs = `
	CREATE TABLE IF NOT EXISTS ` + TblEmailIDs + ` (
		` + ColID + ` CHAR PRIMARY KEY NOT NULL CHECK (` + ColID + ` != ''),
		` + ColUserID + ` CHAR UNIQUE NOT NULL REFERENCES ` + TblUsers + ` (` + ColUserID + `),
		` + ColEmail + ` CHAR UNIQUE NOT NULL CHECK (` + ColEmail + ` != ''),
		` + ColVerified + ` BOOL NOT NULL DEFAULT FALSE,
		` + ColCreateDate + ` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(),
		` + ColUpdateDate + ` TIMESTAMP NOT NULL
	);
	`
	TblDescEmailTokens = `
	CREATE TABLE IF NOT EXISTS ` + TblEmailTokens + ` (
		` + ColID + ` CHAR PRIMARY KEY NOT NULL CHECK (` + ColID + ` != ''),
		` + ColUserID + ` CHAR NOT NULL REFERENCES ` + TblUsers + ` (` + ColUserID + `),
		` + ColEmail + ` CHAR NOT NULL REFERENCES ` + TblEmailIDs + ` (` + ColEmail + `),
		` + ColToken + ` BYTEA NOT NULL,
		` + ColIsUsed + ` BOOL NOT NULL,
		` + ColIssueDate + ` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(),
		` + ColExpiryDate + ` TIMESTAMP NOT NULL,
		INDEX (` + ColUserID + `),
		INDEX (` + ColEmail + `)
	);
	`
	TblDescPhoneIDs = `
	CREATE TABLE IF NOT EXISTS ` + TblPhoneIDs + ` (
		` + ColID + ` CHAR PRIMARY KEY NOT NULL CHECK (` + ColID + ` != ''),
		` + ColUserID + ` CHAR UNIQUE NOT NULL REFERENCES ` + TblUsers + ` (` + ColUserID + `),
		` + ColPhone + ` CHAR UNIQUE NOT NULL CHECK (` + ColPhone + ` != ''),
		` + ColVerified + ` BOOL NOT NULL DEFAULT FALSE,
		` + ColCreateDate + ` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(),
		` + ColUpdateDate + ` TIMESTAMP NOT NULL
	);
	`
	TblDescPhoneTokens = `
	CREATE TABLE IF NOT EXISTS ` + TblPhoneTokens + ` (
		` + ColID + ` CHAR PRIMARY KEY NOT NULL CHECK (` + ColID + ` != ''),
		` + ColUserID + ` CHAR NOT NULL REFERENCES ` + TblUsers + ` (` + ColUserID + `),
		` + ColPhone + ` CHAR NOT NULL REFERENCES ` + TblPhoneIDs + ` (` + ColPhone + `),
		` + ColToken + ` BYTEA NOT NULL,
		` + ColIsUsed + ` BOOL NOT NULL,
		` + ColIssueDate + ` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(),
		` + ColExpiryDate + ` TIMESTAMP NOT NULL,
		INDEX (` + ColUserID + `),
		INDEX (` + ColEmail + `)
	);
	`
	TblDescFacebookIDs = `
	CREATE TABLE IF NOT EXISTS ` + TblFacebookIDs + ` (
		` + ColID + ` CHAR PRIMARY KEY NOT NULL CHECK (` + ColID + ` != ''),
		` + ColUserID + ` CHAR UNIQUE NOT NULL REFERENCES ` + TblUsers + ` (` + ColUserID + `),
		` + ColFacebookID + ` CHAR UNIQUE NOT NULL CHECK(` + ColFacebookID + ` != ''),
		` + ColVerified + ` BOOL NOT NULL DEFAULT FALSE,
		` + ColCreateDate + ` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(),
		` + ColUpdateDate + ` TIMESTAMP NOT NULL
	);
	`
)

var AllTableDescs = []string{
	TblDescConfigurations,
	TblDescUserTypes,
	TblDescUsers,
	TblDescGroups,
	TblDescUsersGroupsJoin,
	TblDescAPIKeys,
	TblDescUserNameIDs,
	TblDescEmailIDs,
	TblDescEmailTokens,
	TblDescPhoneIDs,
	TblDescPhoneTokens,
	TblDescFacebookIDs,
}
