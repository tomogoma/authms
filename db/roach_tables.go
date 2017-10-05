package db

const (
	// Database definition version
	Version = 1

	// Table names
	TblConfigurations = "configurations"
	TblAPIKeys        = "apiKeys"
	TblUserTypes      = "userTypes"
	TblGroups         = "groups"
	TblUsers          = "users"
	TblUserGroupsJoin = "userGroupsJoin"
	TblDeviceIDs      = "deviceIDs"
	TblUserNameIDs    = "userNameIDs"
	TblEmailIDs       = "emailIDs"
	TblEmailTokens    = "emailTokens"
	TblPhoneIDs       = "phoneIDs"
	TblPhoneTokens    = "phoneTokens"
	TblFacebookIDs    = "facebookIDs"
	TblRefreshTokens  = "refreshTokens"

	// DB Table Columns
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
	ColDevID       = "deviceID"
	ColIsRevoked   = "isRevoked"
	ColAPIKeyID    = "apiKeyID"

	// CREATE TABLE DESCRIPTIONS
	TblDescConfigurations = `
	CREATE TABLE IF NOT EXISTS ` + TblConfigurations + ` (
		` + ColKey + ` CHAR PRIMARY KEY NOT NULL CHECK (` + ColKey + ` != ''),
		` + ColValue + ` BYTEA NOT NULL CHECK (` + ColValue + ` != ''),
		` + ColCreateDate + ` TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP(),
		` + ColUpdateDate + ` TIMESTAMPTZ NOT NULL
	);
	`
	TblDescUserTypes = `
	CREATE TABLE IF NOT EXISTS ` + TblUserTypes + ` (
		` + ColID + ` SERIAL PRIMARY KEY NOT NULL CHECK (` + ColID + `>0),
		` + ColName + ` CHAR UNIQUE NOT NULL CHECK (` + ColName + ` != ''),
		` + ColCreateDate + ` TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP(),
		` + ColUpdateDate + ` TIMESTAMPTZ NOT NULL
	);
	`
	TblDescGroups = `
	CREATE TABLE IF NOT EXISTS ` + TblGroups + ` (
		` + ColID + ` SERIAL PRIMARY KEY NOT NULL CHECK (` + ColID + `>0),
		` + ColName + ` CHAR UNIQUE NOT NULL CHECK (` + ColName + ` != ''),
		` + ColAccessLevel + ` FLOAT UNIQUE NOT NULL CHECK (` + ColAccessLevel + ` BETWEEN 0 AND 10),
		` + ColCreateDate + ` TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP(),
		` + ColUpdateDate + ` TIMESTAMPTZ NOT NULL
	);
	`
	TblDescUsers = `
	CREATE TABLE IF NOT EXISTS ` + TblUsers + ` (
		` + ColID + ` SERIAL PRIMARY KEY NOT NULL CHECK (` + ColID + `>0),
		` + ColTypeID + ` INTEGER NOT NULL REFERENCES ` + TblUserTypes + ` (` + ColID + `),
		` + ColPassword + ` BYTEA NOT NULL CHECK ( LENGTH(` + ColPassword + `) >= 8 ),
		` + ColCreateDate + ` TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP(),
		` + ColUpdateDate + ` TIMESTAMPTZ NOT NULL
	);
	`
	TblDescUsersGroupsJoin = `
	CREATE TABLE IF NOT EXISTS ` + TblUserGroupsJoin + ` (
		` + ColID + ` SERIAL PRIMARY KEY NOT NULL CHECK (` + ColID + `>0),
		` + ColUserID + ` INTEGER NOT NULL REFERENCES ` + TblUsers + ` (` + ColID + `),
		` + ColGroupID + ` INTEGER NOT NULL REFERENCES ` + TblGroups + ` (` + ColID + `),
		` + ColCreateDate + ` TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP(),
		` + ColUpdateDate + ` TIMESTAMPTZ NOT NULL,
		UNIQUE (` + ColUserID + `,` + ColGroupID + `)
	);
	`
	TblDescDeviceIDs = `
	CREATE TABLE IF NOT EXISTS ` + TblDeviceIDs + ` (
		` + ColID + ` SERIAL PRIMARY KEY NOT NULL CHECK (` + ColID + `>0),
		` + ColDevID + ` CHAR UNIQUE NOT NULL CHECK (` + ColDevID + ` != ''),
		` + ColUserID + ` INTEGER NOT NULL REFERENCES ` + TblUsers + ` (` + ColID + `),
		` + ColCreateDate + ` TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP(),
		` + ColUpdateDate + ` TIMESTAMPTZ NOT NULL
	);
	`
	TblDescAPIKeys = `
	CREATE TABLE IF NOT EXISTS ` + TblAPIKeys + ` (
		` + ColID + ` SERIAL PRIMARY KEY NOT NULL CHECK (` + ColID + `>0),
		` + ColUserID + ` INTEGER NOT NULL REFERENCES ` + TblUsers + ` (` + ColID + `),
		` + ColAPIKey + ` BYTEA NOT NULL CHECK ( LENGTH(` + ColAPIKey + `) >= 56 ),
		` + ColCreateDate + ` TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP(),
		` + ColUpdateDate + ` TIMESTAMPTZ NOT NULL
	);
	`
	TblDescUserNameIDs = `
	CREATE TABLE IF NOT EXISTS ` + TblUserNameIDs + ` (
		` + ColID + ` SERIAL PRIMARY KEY NOT NULL CHECK (` + ColID + `>0),
		` + ColUserName + ` STRING UNIQUE NOT NULL,
		` + ColUserID + ` INTEGER UNIQUE NOT NULL REFERENCES ` + TblUsers + ` (` + ColID + `),
		` + ColCreateDate + ` TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP(),
		` + ColUpdateDate + ` TIMESTAMPTZ NOT NULL
	);
	`
	TblDescEmailIDs = `
	CREATE TABLE IF NOT EXISTS ` + TblEmailIDs + ` (
		` + ColID + ` SERIAL PRIMARY KEY NOT NULL CHECK (` + ColID + `>0),
		` + ColEmail + ` CHAR UNIQUE NOT NULL CHECK (` + ColEmail + ` != ''),
		` + ColUserID + ` INTEGER UNIQUE NOT NULL REFERENCES ` + TblUsers + ` (` + ColID + `),
		` + ColVerified + ` BOOL NOT NULL DEFAULT FALSE,
		` + ColCreateDate + ` TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP(),
		` + ColUpdateDate + ` TIMESTAMPTZ NOT NULL
	);
	`
	TblDescEmailTokens = `
	CREATE TABLE IF NOT EXISTS ` + TblEmailTokens + ` (
		` + ColID + ` SERIAL PRIMARY KEY NOT NULL CHECK (` + ColID + `>0),
		` + ColUserID + ` INTEGER NOT NULL REFERENCES ` + TblUsers + ` (` + ColID + `),
		` + ColEmail + ` CHAR NOT NULL REFERENCES ` + TblEmailIDs + ` (` + ColEmail + `),
		` + ColToken + ` BYTEA NOT NULL CHECK (LENGTH(` + ColToken + `)>0),
		` + ColIsUsed + ` BOOL NOT NULL,
		` + ColIssueDate + ` TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP(),
		` + ColExpiryDate + ` TIMESTAMPTZ NOT NULL
	);
	`
	TblDescPhoneIDs = `
	CREATE TABLE IF NOT EXISTS ` + TblPhoneIDs + ` (
		` + ColID + ` SERIAL PRIMARY KEY NOT NULL CHECK (` + ColID + `>0),
		` + ColPhone + ` CHAR UNIQUE NOT NULL CHECK (` + ColPhone + ` != ''),
		` + ColUserID + ` INTEGER UNIQUE NOT NULL REFERENCES ` + TblUsers + ` (` + ColID + `),
		` + ColVerified + ` BOOL NOT NULL DEFAULT FALSE,
		` + ColCreateDate + ` TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP(),
		` + ColUpdateDate + ` TIMESTAMPTZ NOT NULL
	);
	`
	TblDescPhoneTokens = `
	CREATE TABLE IF NOT EXISTS ` + TblPhoneTokens + ` (
		` + ColID + ` SERIAL PRIMARY KEY NOT NULL CHECK (` + ColID + `>0),
		` + ColUserID + ` INTEGER NOT NULL REFERENCES ` + TblUsers + ` (` + ColID + `),
		` + ColPhone + ` CHAR NOT NULL REFERENCES ` + TblPhoneIDs + ` (` + ColPhone + `),
		` + ColToken + ` BYTEA NOT NULL CHECK (LENGTH(` + ColToken + `)>0),
		` + ColIsUsed + ` BOOL NOT NULL,
		` + ColIssueDate + ` TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP(),
		` + ColExpiryDate + ` TIMESTAMPTZ NOT NULL
	);
	`
	TblDescFacebookIDs = `
	CREATE TABLE IF NOT EXISTS ` + TblFacebookIDs + ` (
		` + ColID + ` SERIAL PRIMARY KEY NOT NULL CHECK (` + ColID + `>0),
		` + ColFacebookID + ` CHAR UNIQUE NOT NULL CHECK(` + ColFacebookID + ` != ''),
		` + ColUserID + ` INTEGER UNIQUE NOT NULL REFERENCES ` + TblUsers + ` (` + ColID + `),
		` + ColVerified + ` BOOL NOT NULL DEFAULT FALSE,
		` + ColCreateDate + ` TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP(),
		` + ColUpdateDate + ` TIMESTAMPTZ NOT NULL
	);
	`
	TblDescRefreshTokens = `
	CREATE TABLE IF NOT EXISTS ` + TblRefreshTokens + ` (
		` + ColID + ` SERIAL PRIMARY KEY NOT NULL CHECK (` + ColID + `>0),
		` + ColUserID + ` INTEGER NOT NULL REFERENCES ` + TblUsers + ` (` + ColID + `),
		` + ColAPIKeyID + ` INTEGER NOT NULL REFERENCES ` + TblAPIKeys + ` (` + ColID + `),
		` + ColToken + ` BYTEA NOT NULL,
		` + ColIsRevoked + ` BOOL NOT NULL,
		` + ColIssueDate + ` TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP(),
		` + ColExpiryDate + ` TIMESTAMPTZ NOT NULL
	);
	`
)

// AllTableDescs lists all CREATE TABLE DESCRIPTIONS in order of dependency
// (tables with foreign key references listed after parent table descriptions).
var AllTableDescs = []string{
	TblDescConfigurations,
	TblDescUserTypes,
	TblDescGroups,
	TblDescUsers,
	TblDescUsersGroupsJoin,
	TblDescAPIKeys,
	TblDescDeviceIDs,
	TblDescUserNameIDs,
	TblDescEmailIDs,
	TblDescEmailTokens,
	TblDescPhoneIDs,
	TblDescPhoneTokens,
	TblDescFacebookIDs,
	TblDescRefreshTokens,
}

// AllTableNames lists all table names in order of dependency
// (tables with foreign key references listed after parent table descriptions).
var AllTableNames = []string{
	TblConfigurations,
	TblUserTypes,
	TblGroups,
	TblUsers,
	TblUserGroupsJoin,
	TblAPIKeys,
	TblDeviceIDs,
	TblUserNameIDs,
	TblEmailIDs,
	TblEmailTokens,
	TblPhoneIDs,
	TblPhoneTokens,
	TblFacebookIDs,
	TblRefreshTokens,
}
