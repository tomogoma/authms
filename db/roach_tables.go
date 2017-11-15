package db

const (
	// Database definition version
	Version = 1

	// Table names
	TblConfigurations = "configurations"
	TblUserTypes      = "userTypes"
	TblGroups         = "groups"
	TblUsers          = "users"
	TblUserGroupsJoin = "userGroupsJoin"
	TblAPIKeys        = "apiKeys"
	TblDeviceIDs      = "deviceIDs"
	TblUserNames      = "userNames"
	TblEmails         = "emails"
	TblEmailTokens    = "emailTokens"
	TblPhones         = "phones"
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
	ColTypeID      = "typeID"
	ColDevID       = "deviceID"
	ColIsRevoked   = "isRevoked"
	ColAPIKeyID    = "apiKeyID"

	// CREATE TABLE DESCRIPTIONS
	TblDescConfigurations = `
	CREATE TABLE IF NOT EXISTS ` + TblConfigurations + ` (
		` + ColKey + ` VARCHAR(56) PRIMARY KEY NOT NULL CHECK (` + ColKey + ` != ''),
		` + ColValue + ` BYTEA NOT NULL CHECK (` + ColValue + ` != ''),
		` + ColCreateDate + ` TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		` + ColUpdateDate + ` TIMESTAMPTZ NOT NULL
	);
	`
	TblDescUserTypes = `
	CREATE TABLE IF NOT EXISTS ` + TblUserTypes + ` (
		` + ColID + ` BIGSERIAL PRIMARY KEY NOT NULL CHECK (` + ColID + `>0),
		` + ColName + ` VARCHAR(56) UNIQUE NOT NULL CHECK (` + ColName + ` != ''),
		` + ColCreateDate + ` TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		` + ColUpdateDate + ` TIMESTAMPTZ NOT NULL
	);
	`
	TblDescGroups = `
	CREATE TABLE IF NOT EXISTS ` + TblGroups + ` (
		` + ColID + ` BIGSERIAL PRIMARY KEY NOT NULL CHECK (` + ColID + `>0),
		` + ColName + ` VARCHAR(56) UNIQUE NOT NULL CHECK (` + ColName + ` != ''),
		` + ColAccessLevel + ` FLOAT NOT NULL CHECK (` + ColAccessLevel + ` BETWEEN 0 AND 10),
		` + ColCreateDate + ` TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		` + ColUpdateDate + ` TIMESTAMPTZ NOT NULL
	);
	`
	TblDescUsers = `
	CREATE TABLE IF NOT EXISTS ` + TblUsers + ` (
		` + ColID + ` BIGSERIAL PRIMARY KEY NOT NULL CHECK (` + ColID + `>0),
		` + ColTypeID + ` BIGINT NOT NULL REFERENCES ` + TblUserTypes + ` (` + ColID + `),
		` + ColPassword + ` BYTEA NOT NULL CHECK ( LENGTH(` + ColPassword + `) >= 8 ),
		` + ColCreateDate + ` TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		` + ColUpdateDate + ` TIMESTAMPTZ NOT NULL
	);
	`
	TblDescUsersGroupsJoin = `
	CREATE TABLE IF NOT EXISTS ` + TblUserGroupsJoin + ` (
		` + ColID + ` BIGSERIAL PRIMARY KEY NOT NULL CHECK (` + ColID + `>0),
		` + ColUserID + ` BIGINT NOT NULL REFERENCES ` + TblUsers + ` (` + ColID + `),
		` + ColGroupID + ` BIGINT NOT NULL REFERENCES ` + TblGroups + ` (` + ColID + `),
		` + ColCreateDate + ` TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		` + ColUpdateDate + ` TIMESTAMPTZ NOT NULL,
		UNIQUE (` + ColUserID + `,` + ColGroupID + `)
	);
	`
	TblDescAPIKeys = `
	CREATE TABLE IF NOT EXISTS ` + TblAPIKeys + ` (
		` + ColID + ` BIGSERIAL PRIMARY KEY NOT NULL CHECK (` + ColID + `>0),
		` + ColUserID + ` BIGINT NOT NULL REFERENCES ` + TblUsers + ` (` + ColID + `),
		` + ColKey + ` VARCHAR(256) NOT NULL CHECK ( LENGTH(` + ColKey + `) >= 56 ),
		` + ColCreateDate + ` TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		` + ColUpdateDate + ` TIMESTAMPTZ NOT NULL
	);
	`
	TblDescDeviceIDs = `
	CREATE TABLE IF NOT EXISTS ` + TblDeviceIDs + ` (
		` + ColID + ` BIGSERIAL PRIMARY KEY NOT NULL CHECK (` + ColID + `>0),
		` + ColDevID + ` VARCHAR(256) UNIQUE NOT NULL CHECK (` + ColDevID + ` != ''),
		` + ColUserID + ` BIGINT NOT NULL REFERENCES ` + TblUsers + ` (` + ColID + `),
		` + ColCreateDate + ` TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		` + ColUpdateDate + ` TIMESTAMPTZ NOT NULL
	);
	`
	TblDescUserNames = `
	CREATE TABLE IF NOT EXISTS ` + TblUserNames + ` (
		` + ColID + ` BIGSERIAL PRIMARY KEY NOT NULL CHECK (` + ColID + `>0),
		` + ColUserName + ` VARCHAR(56) UNIQUE NOT NULL,
		` + ColUserID + ` BIGINT UNIQUE NOT NULL REFERENCES ` + TblUsers + ` (` + ColID + `),
		` + ColCreateDate + ` TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		` + ColUpdateDate + ` TIMESTAMPTZ NOT NULL
	);
	`
	TblDescEmails = `
	CREATE TABLE IF NOT EXISTS ` + TblEmails + ` (
		` + ColID + ` BIGSERIAL PRIMARY KEY NOT NULL CHECK (` + ColID + `>0),
		` + ColEmail + ` VARCHAR(128) UNIQUE NOT NULL CHECK (` + ColEmail + ` != ''),
		` + ColUserID + ` BIGINT UNIQUE NOT NULL REFERENCES ` + TblUsers + ` (` + ColID + `),
		` + ColVerified + ` BOOL NOT NULL DEFAULT FALSE,
		` + ColCreateDate + ` TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		` + ColUpdateDate + ` TIMESTAMPTZ NOT NULL
	);
	`
	TblDescEmailTokens = `
	CREATE TABLE IF NOT EXISTS ` + TblEmailTokens + ` (
		` + ColID + ` BIGSERIAL PRIMARY KEY NOT NULL CHECK (` + ColID + `>0),
		` + ColUserID + ` BIGINT NOT NULL REFERENCES ` + TblUsers + ` (` + ColID + `),
		` + ColEmail + ` VARCHAR(128) NOT NULL REFERENCES ` + TblEmails + ` (` + ColEmail + `),
		` + ColToken + ` BYTEA NOT NULL CHECK (LENGTH(` + ColToken + `)>0),
		` + ColIsUsed + ` BOOL NOT NULL,
		` + ColIssueDate + ` TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		` + ColExpiryDate + ` TIMESTAMPTZ NOT NULL
	);
	`
	TblDescPhones = `
	CREATE TABLE IF NOT EXISTS ` + TblPhones + ` (
		` + ColID + ` BIGSERIAL PRIMARY KEY NOT NULL CHECK (` + ColID + `>0),
		` + ColPhone + ` VARCHAR(56) UNIQUE NOT NULL CHECK (` + ColPhone + ` != ''),
		` + ColUserID + ` BIGINT UNIQUE NOT NULL REFERENCES ` + TblUsers + ` (` + ColID + `),
		` + ColVerified + ` BOOL NOT NULL DEFAULT FALSE,
		` + ColCreateDate + ` TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		` + ColUpdateDate + ` TIMESTAMPTZ NOT NULL
	);
	`
	TblDescPhoneTokens = `
	CREATE TABLE IF NOT EXISTS ` + TblPhoneTokens + ` (
		` + ColID + ` BIGSERIAL PRIMARY KEY NOT NULL CHECK (` + ColID + `>0),
		` + ColUserID + ` BIGINT NOT NULL REFERENCES ` + TblUsers + ` (` + ColID + `),
		` + ColPhone + ` VARCHAR(56) NOT NULL REFERENCES ` + TblPhones + ` (` + ColPhone + `),
		` + ColToken + ` BYTEA NOT NULL CHECK (LENGTH(` + ColToken + `)>0),
		` + ColIsUsed + ` BOOL NOT NULL,
		` + ColIssueDate + ` TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		` + ColExpiryDate + ` TIMESTAMPTZ NOT NULL
	);
	`
	TblDescFacebookIDs = `
	CREATE TABLE IF NOT EXISTS ` + TblFacebookIDs + ` (
		` + ColID + ` BIGSERIAL PRIMARY KEY NOT NULL CHECK (` + ColID + `>0),
		` + ColFacebookID + ` VARCHAR(512) UNIQUE NOT NULL CHECK(` + ColFacebookID + ` != ''),
		` + ColUserID + ` BIGINT UNIQUE NOT NULL REFERENCES ` + TblUsers + ` (` + ColID + `),
		` + ColVerified + ` BOOL NOT NULL DEFAULT FALSE,
		` + ColCreateDate + ` TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		` + ColUpdateDate + ` TIMESTAMPTZ NOT NULL
	);
	`
	TblDescRefreshTokens = `
	CREATE TABLE IF NOT EXISTS ` + TblRefreshTokens + ` (
		` + ColID + ` BIGSERIAL PRIMARY KEY NOT NULL CHECK (` + ColID + `>0),
		` + ColUserID + ` BIGINT NOT NULL REFERENCES ` + TblUsers + ` (` + ColID + `),
		` + ColAPIKeyID + ` BIGINT NOT NULL REFERENCES ` + TblAPIKeys + ` (` + ColID + `),
		` + ColToken + ` BYTEA NOT NULL,
		` + ColIsRevoked + ` BOOL NOT NULL,
		` + ColIssueDate + ` TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
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
	TblDescUserNames,
	TblDescEmails,
	TblDescEmailTokens,
	TblDescPhones,
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
	TblUserNames,
	TblEmails,
	TblEmailTokens,
	TblPhones,
	TblPhoneTokens,
	TblFacebookIDs,
	TblRefreshTokens,
}
