define({ "api": [
  {
    "type": "get",
    "url": "/groups",
    "title": "Get Groups",
    "name": "GetGroups",
    "version": "0.1.1",
    "group": "Auth",
    "permission": [
      {
        "name": "^admin"
      }
    ],
    "header": {
      "fields": {
        "Header": [
          {
            "group": "Header",
            "optional": false,
            "field": "x-api-key",
            "description": "<p>the api key</p>"
          }
        ]
      }
    },
    "parameter": {
      "fields": {
        "URL Query Parameters": [
          {
            "group": "URL Query Parameters",
            "type": "String",
            "optional": false,
            "field": "token",
            "description": "<p>the JWT accessed during auth.</p>"
          },
          {
            "group": "URL Query Parameters",
            "type": "Number",
            "optional": true,
            "field": "offset",
            "defaultValue": "0",
            "description": "<p>The beginning index to fetch groups.</p>"
          },
          {
            "group": "URL Query Parameters",
            "type": "Number",
            "optional": true,
            "field": "count",
            "defaultValue": "10",
            "description": "<p>The maximum number of groups to fetch.</p>"
          }
        ]
      }
    },
    "success": {
      "fields": {
        "Success 200": [
          {
            "group": "Success 200",
            "type": "Object[]",
            "optional": false,
            "field": "json-body",
            "description": "<p>JSON array of <a href=\"#api-Objects-Group\">groups</a></p>"
          }
        ]
      }
    },
    "filename": "handler/http/handler.go",
    "groupTitle": "Auth"
  },
  {
    "type": "get",
    "url": "/users",
    "title": "Get Users",
    "name": "GetUsers",
    "version": "0.1.1",
    "group": "Auth",
    "permission": [
      {
        "name": "^admin"
      }
    ],
    "header": {
      "fields": {
        "Header": [
          {
            "group": "Header",
            "optional": false,
            "field": "x-api-key",
            "description": "<p>the api key</p>"
          }
        ]
      }
    },
    "parameter": {
      "fields": {
        "URL Query Parameters": [
          {
            "group": "URL Query Parameters",
            "type": "String",
            "optional": false,
            "field": "token",
            "description": "<p>the JWT accessed during auth.</p>"
          },
          {
            "group": "URL Query Parameters",
            "type": "Number",
            "optional": true,
            "field": "offset",
            "defaultValue": "0",
            "description": "<p>The beginning index to fetch groups.</p>"
          },
          {
            "group": "URL Query Parameters",
            "type": "Number",
            "optional": true,
            "field": "count",
            "defaultValue": "10",
            "description": "<p>The maximum number of groups to fetch.</p>"
          },
          {
            "group": "URL Query Parameters",
            "type": "String",
            "optional": true,
            "field": "group",
            "description": "<p>Filter by group name. one can have multiple groups e.g. ?group=admin&amp;group=staff, multiple group names are always filtered using the OR operator.</p>"
          },
          {
            "group": "URL Query Parameters",
            "type": "String",
            "size": "0-10",
            "allowedValues": [
              "gt_[number]",
              "lt_[number]",
              "[number]",
              "gteq_[number]",
              "lteq_[number]"
            ],
            "optional": true,
            "field": "acl",
            "description": "<p>Filter by access levels:</p> <ul> <li>gt_[number] - access level greater than number e.g. gt_5</li> <li>lt_[number] - access level less than number e.g. lt_5</li> <li>[number] - access level equal to number e.g. 5</li> <li>gteq_[number] - access level greater than or equal to number e.g. gteq_5</li> <li>lteq_[number] - access level less than or equal to  number e.g. lteq_5</li> <li>one can have multiple filters e.g. ?acl=gt_5&amp;acl=lteq_9&amp;matchAllACLs=true to get acl in (5 &lt; acl &lt;= 9)</li> </ul>"
          },
          {
            "group": "URL Query Parameters",
            "type": "String",
            "allowedValues": [
              "true",
              "false"
            ],
            "optional": true,
            "field": "matchAllACLs",
            "defaultValue": "false",
            "description": "<p>Setting this to true will force all acl's provided to be matched using the AND operator, otherwise uses the OR operator.</p>"
          },
          {
            "group": "URL Query Parameters",
            "type": "String",
            "allowedValues": [
              "true",
              "false"
            ],
            "optional": true,
            "field": "matchAll",
            "defaultValue": "false",
            "description": "<p>Setting this to true will force acl,group filters to be matched using the AND operator, otherwise uses the OR operator.</p>"
          }
        ]
      }
    },
    "success": {
      "fields": {
        "Success 200": [
          {
            "group": "Success 200",
            "type": "Object[]",
            "optional": false,
            "field": "json-body",
            "description": "<p>JSON array of <a href=\"#api-Objects-User\">users</a></p>"
          }
        ]
      }
    },
    "filename": "handler/http/handler.go",
    "groupTitle": "Auth"
  },
  {
    "type": "POST",
    "url": "/:loginType/login",
    "title": "Login",
    "description": "<p>User login. See <a href=\"#api-Auth-Register\">Register</a> for loginType options.</p>",
    "name": "Login",
    "version": "0.1.0",
    "group": "Auth",
    "header": {
      "fields": {
        "Header": [
          {
            "group": "Header",
            "optional": false,
            "field": "x-api-key",
            "description": "<p>the api key</p>"
          },
          {
            "group": "Header",
            "optional": false,
            "field": "Authorization",
            "description": "<p>Basic auth containing loginType's identifier and password in the format 'Basic: base64Of(identifier:password)'</p>"
          }
        ]
      }
    },
    "parameter": {
      "fields": {
        "URL Parameters": [
          {
            "group": "URL Parameters",
            "type": "String",
            "allowedValues": [
              "usernames",
              "emails",
              "phones",
              "facebook"
            ],
            "optional": false,
            "field": "loginType",
            "description": "<p>type of identifier in Authorization header.</p>"
          }
        ]
      }
    },
    "filename": "handler/http/handler.go",
    "groupTitle": "Auth",
    "success": {
      "fields": {
        "Success 200": [
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "ID",
            "description": "<p>Unique ID of the user (can be cast to long Integer).</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": false,
            "field": "type",
            "description": "<p>The <a href=\"#api-Objects-UserType\">UserType</a> of this user.</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": false,
            "field": "group",
            "description": "<p>The <a href=\"#api-Objects-Group\">group</a> the user belongs to.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "created",
            "description": "<p>The date the user was created.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "lastUpdated",
            "description": "<p>date the user was last updated.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": true,
            "field": "JWT",
            "description": "<p>JSON Web Token for accessing services. This is only provided during <a href=\"#api-Auth-Login\">Login</a>, <a href=\"#api-Auth-Register\">Registration</a> and  <a href=\"#api-Auth-FirstUser\">First User Registration</a>.</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": true,
            "field": "username",
            "description": "<p>The user's <a href=\"#api-Objects-Username\">username</a> (if this user has one).</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": true,
            "field": "phone",
            "description": "<p>The user's <a href=\"#api-Objects-VerifLogin\">phone</a> (if this user has one).</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": true,
            "field": "email",
            "description": "<p>The user's <a href=\"#api-Objects-VerifLogin\">email</a> (if this user has one).</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": true,
            "field": "facebook",
            "description": "<p>The user's <a href=\"#api-Objects-FacebookID\">facebook ID</a> (if this user has one).</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": true,
            "field": "device",
            "description": "<p>The <a href=\"#api-Objects-Device\">device</a> this user is attached to, if any.</p>"
          }
        ]
      }
    }
  },
  {
    "type": "put",
    "url": "/:loginType/register",
    "title": "Register",
    "description": "<p>Register new user.</p>",
    "permission": [
      {
        "name": "^admin for registering other"
      }
    ],
    "name": "Register",
    "version": "0.1.0",
    "group": "Auth",
    "header": {
      "fields": {
        "Header": [
          {
            "group": "Header",
            "optional": false,
            "field": "x-api-key",
            "description": "<p>the api key</p>"
          }
        ]
      }
    },
    "parameter": {
      "fields": {
        "URL Parameters": [
          {
            "group": "URL Parameters",
            "type": "String",
            "allowedValues": [
              "usernames",
              "emails",
              "phones",
              "facebook"
            ],
            "optional": false,
            "field": "loginType",
            "description": "<p>type of identifier in JSON Body</p>"
          }
        ],
        "URL Query Parameters": [
          {
            "group": "URL Query Parameters",
            "type": "String",
            "allowedValues": [
              "true",
              "device"
            ],
            "optional": false,
            "field": "selfReg",
            "description": "<p>Whether registering self or not:</p> <ul> <li>true for self registration</li> <li>device for self registration by unique device ID</li> <li>not-provided for admin to register any other user</li> </ul>"
          }
        ],
        "JSON Request Body": [
          {
            "group": "JSON Request Body",
            "type": "String",
            "allowedValues": [
              "individual",
              "company"
            ],
            "optional": false,
            "field": "userType",
            "description": "<p>Type of user.</p>"
          },
          {
            "group": "JSON Request Body",
            "type": "String",
            "optional": false,
            "field": "identifier",
            "description": "<p>The 'username' corresponding to loginType.</p>"
          },
          {
            "group": "JSON Request Body",
            "type": "String",
            "optional": true,
            "field": "secret",
            "description": "<p>The user's password - required when selfReg set to true or device.</p>"
          },
          {
            "group": "JSON Request Body",
            "type": "String",
            "optional": true,
            "field": "groupID",
            "description": "<p>groupID to add this user to - required when selfReg not set.</p>"
          },
          {
            "group": "JSON Request Body",
            "type": "String",
            "optional": true,
            "field": "deviceID",
            "description": "<p>the unique device ID for the user - required when selfReg=device.</p>"
          }
        ]
      }
    },
    "success": {
      "fields": {
        "Success 201": [
          {
            "group": "Success 201",
            "type": "Object",
            "optional": false,
            "field": "json-body",
            "description": "<p>See <a href=\"#api-Objects-User\">User</a> for details.</p>"
          }
        ]
      }
    },
    "filename": "handler/http/handler.go",
    "groupTitle": "Auth"
  },
  {
    "type": "POST",
    "url": "/reset_password",
    "title": "Reset password",
    "description": "<p>Reset a user's password.</p>",
    "name": "ResetPassword",
    "version": "0.1.0",
    "group": "Auth",
    "header": {
      "fields": {
        "Header": [
          {
            "group": "Header",
            "optional": false,
            "field": "x-api-key",
            "description": "<p>the api key</p>"
          }
        ]
      }
    },
    "parameter": {
      "fields": {
        "JSON Request Body": [
          {
            "group": "JSON Request Body",
            "type": "String",
            "allowedValues": [
              "usernames",
              "emails",
              "phones",
              "facebook"
            ],
            "optional": false,
            "field": "loginType",
            "description": "<p>type of identifier for which password reset is sort.</p>"
          },
          {
            "group": "JSON Request Body",
            "type": "String",
            "optional": false,
            "field": "identifier",
            "description": "<p>the loginType's unique identifier for which password reset is sort.</p>"
          },
          {
            "group": "JSON Request Body",
            "type": "String",
            "optional": false,
            "field": "OTP",
            "description": "<p>The password reset code sent to user during <a href=\"#api-Auth-SendPasswordResetOTP\">SendPasswordResetOTP</a>.</p>"
          },
          {
            "group": "JSON Request Body",
            "type": "String",
            "optional": false,
            "field": "newSecret",
            "description": "<p>The new password.</p>"
          }
        ]
      }
    },
    "filename": "handler/http/handler.go",
    "groupTitle": "Auth",
    "success": {
      "fields": {
        "Success 200": [
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "ID",
            "description": "<p>Unique ID of the verifiable login (can be cast to long Integer).</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "userID",
            "description": "<p>ID for user who owns this verifiable login.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "value",
            "description": "<p>The unique verifiable login string value.</p>"
          },
          {
            "group": "Success 200",
            "type": "Boolean",
            "optional": false,
            "field": "verified",
            "description": "<p>True if this login is verified, false otherwise.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "created",
            "description": "<p>ISO8601 date the verifiable login was created.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "lastUpdated",
            "description": "<p>ISO8601 date the verifiable login was last updated.</p>"
          }
        ]
      }
    }
  },
  {
    "type": "POST",
    "url": "/reset_password/send_otp",
    "title": "Send Password Reset OTP",
    "description": "<p>Send Password reset Code (OTP) to identifier of type loginType.</p>",
    "name": "SendPasswordResetOTP",
    "version": "0.1.0",
    "group": "Auth",
    "header": {
      "fields": {
        "Header": [
          {
            "group": "Header",
            "optional": false,
            "field": "x-api-key",
            "description": "<p>the api key</p>"
          }
        ]
      }
    },
    "parameter": {
      "fields": {
        "JSON Request Body": [
          {
            "group": "JSON Request Body",
            "type": "String",
            "allowedValues": [
              "usernames",
              "emails",
              "phones",
              "facebook"
            ],
            "optional": false,
            "field": "loginType",
            "description": "<p>type of identifier to send OTP to.</p>"
          },
          {
            "group": "JSON Request Body",
            "type": "String",
            "optional": false,
            "field": "identifier",
            "description": "<p>The loginType's unique identifier for which to send password reset code to.</p>"
          }
        ]
      }
    },
    "filename": "handler/http/handler.go",
    "groupTitle": "Auth",
    "success": {
      "fields": {
        "Success 200": [
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "obfuscatedAddress",
            "description": "<p>Obfuscated address to which OTP was sent.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "expiresAt",
            "description": "<p>ISO8601 expiry date of OTP.</p>"
          }
        ]
      }
    }
  },
  {
    "type": "POST",
    "url": "/:loginType/verify",
    "title": "Send Verification Code",
    "description": "<p>Send OTP to identifier of type loginType for purpose of verifying identifier.</p>",
    "name": "SendVerificationCode",
    "version": "0.2.0",
    "group": "Auth",
    "header": {
      "fields": {
        "Header": [
          {
            "group": "Header",
            "optional": false,
            "field": "x-api-key",
            "description": "<p>the api key</p>"
          }
        ]
      }
    },
    "parameter": {
      "fields": {
        "URL Parameters": [
          {
            "group": "URL Parameters",
            "type": "String",
            "allowedValues": [
              "usernames",
              "emails",
              "phones",
              "facebook"
            ],
            "optional": false,
            "field": "loginType",
            "description": "<p>type of identifier in JSON Body</p>"
          }
        ],
        "URL Query Parameters": [
          {
            "group": "URL Query Parameters",
            "type": "String",
            "optional": false,
            "field": "token",
            "description": "<p>the JWT provided during login.</p>"
          }
        ],
        "JSON Request Body": [
          {
            "group": "JSON Request Body",
            "type": "String",
            "optional": false,
            "field": "identifier",
            "description": "<p>The loginType's address to be verified.</p>"
          }
        ]
      }
    },
    "filename": "handler/http/handler.go",
    "groupTitle": "Auth",
    "success": {
      "fields": {
        "Success 200": [
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "obfuscatedAddress",
            "description": "<p>Obfuscated address to which OTP was sent.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "expiresAt",
            "description": "<p>ISO8601 expiry date of OTP.</p>"
          }
        ]
      }
    }
  },
  {
    "type": "POST",
    "url": "/:loginType/verify/:identifier?token=:JWT",
    "title": "Send Verification Code",
    "description": "<p>Send OTP to identifier of type loginType for purpose of verifying identifier. See <a href=\"#api-Auth-Register\">Register</a> for loginType and identifier options.</p>",
    "name": "SendVerificationCode",
    "version": "0.1.0",
    "group": "Auth",
    "header": {
      "fields": {
        "Header": [
          {
            "group": "Header",
            "optional": false,
            "field": "x-api-key",
            "description": "<p>the api key</p>"
          }
        ]
      }
    },
    "parameter": {
      "fields": {
        "Parameter": [
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "identifier",
            "description": "<p>The new 'username' corresponding to loginType</p>"
          }
        ]
      }
    },
    "success": {
      "fields": {
        "200": [
          {
            "group": "200",
            "type": "Object",
            "optional": false,
            "field": "json-body",
            "description": "<p>See <a href=\"#api-Objects-OTPStatus\">OTPStatus</a>.</p>"
          }
        ]
      }
    },
    "filename": "handler/http/_apidoc.js",
    "groupTitle": "Auth"
  },
  {
    "type": "POST",
    "url": "/users/:userID/set_group/:groupID",
    "title": "Set User's Group",
    "description": "<p>Assign group to user.</p>",
    "name": "SetUserGroup",
    "version": "0.1.0",
    "permission": [
      {
        "name": "^admin"
      }
    ],
    "group": "Auth",
    "header": {
      "fields": {
        "Header": [
          {
            "group": "Header",
            "optional": false,
            "field": "x-api-key",
            "description": "<p>the api key</p>"
          }
        ]
      }
    },
    "parameter": {
      "fields": {
        "URL Parameters": [
          {
            "group": "URL Parameters",
            "type": "String",
            "optional": false,
            "field": "userID",
            "description": "<p>The ID of the <a href=\"#api-Objects-User\">user</a> to update.</p>"
          },
          {
            "group": "URL Parameters",
            "type": "String",
            "optional": false,
            "field": "groupID",
            "description": "<p>The ID of the <a href=\"#api-Objects-Group\">group</a> to assign the user to.</p>"
          }
        ],
        "URL Query Parameters": [
          {
            "group": "URL Query Parameters",
            "type": "String",
            "optional": false,
            "field": "token",
            "description": "<p>the JWT provided during login.</p>"
          }
        ]
      }
    },
    "filename": "handler/http/handler.go",
    "groupTitle": "Auth",
    "success": {
      "fields": {
        "Success 200": [
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "ID",
            "description": "<p>Unique ID of the user (can be cast to long Integer).</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": false,
            "field": "type",
            "description": "<p>The <a href=\"#api-Objects-UserType\">UserType</a> of this user.</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": false,
            "field": "group",
            "description": "<p>The <a href=\"#api-Objects-Group\">group</a> the user belongs to.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "created",
            "description": "<p>The date the user was created.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "lastUpdated",
            "description": "<p>date the user was last updated.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": true,
            "field": "JWT",
            "description": "<p>JSON Web Token for accessing services. This is only provided during <a href=\"#api-Auth-Login\">Login</a>, <a href=\"#api-Auth-Register\">Registration</a> and  <a href=\"#api-Auth-FirstUser\">First User Registration</a>.</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": true,
            "field": "username",
            "description": "<p>The user's <a href=\"#api-Objects-Username\">username</a> (if this user has one).</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": true,
            "field": "phone",
            "description": "<p>The user's <a href=\"#api-Objects-VerifLogin\">phone</a> (if this user has one).</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": true,
            "field": "email",
            "description": "<p>The user's <a href=\"#api-Objects-VerifLogin\">email</a> (if this user has one).</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": true,
            "field": "facebook",
            "description": "<p>The user's <a href=\"#api-Objects-FacebookID\">facebook ID</a> (if this user has one).</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": true,
            "field": "device",
            "description": "<p>The <a href=\"#api-Objects-Device\">device</a> this user is attached to, if any.</p>"
          }
        ]
      }
    }
  },
  {
    "type": "get",
    "url": "/status",
    "title": "Status",
    "name": "Status",
    "version": "0.1.0",
    "group": "Auth",
    "header": {
      "fields": {
        "Header": [
          {
            "group": "Header",
            "optional": false,
            "field": "x-api-key",
            "description": "<p>the api key</p>"
          }
        ]
      }
    },
    "success": {
      "fields": {
        "Success 200": [
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "name",
            "description": "<p>Micro-service name.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "version",
            "description": "<p>Current running version.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "description",
            "description": "<p>Short description of the micro-service.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "canonicalName",
            "description": "<p>Canonical name of the micro-service.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "needRegSuper",
            "description": "<p>true if a super-user has been registered, false otherwise.</p>"
          }
        ]
      }
    },
    "filename": "handler/http/handler.go",
    "groupTitle": "Auth"
  },
  {
    "type": "POST",
    "url": "/users/:userID",
    "title": "Update Identifier",
    "description": "<p>Update (or set for first time) the identifier details for loginType.</p>",
    "name": "UpdateIdentifier",
    "version": "0.1.0",
    "group": "Auth",
    "header": {
      "fields": {
        "Header": [
          {
            "group": "Header",
            "optional": false,
            "field": "x-api-key",
            "description": "<p>the api key</p>"
          }
        ]
      }
    },
    "parameter": {
      "fields": {
        "URL Parameters": [
          {
            "group": "URL Parameters",
            "type": "String",
            "optional": false,
            "field": "userID",
            "description": "<p>The ID of the <a href=\"#api-Objects-User\">user</a> to update.</p>"
          }
        ],
        "URL Query Parameters": [
          {
            "group": "URL Query Parameters",
            "type": "String",
            "optional": false,
            "field": "token",
            "description": "<p>the JWT provided during login.</p>"
          }
        ],
        "JSON Request Body": [
          {
            "group": "JSON Request Body",
            "type": "String",
            "allowedValues": [
              "usernames",
              "emails",
              "phones",
              "facebook"
            ],
            "optional": false,
            "field": "loginType",
            "description": "<p>type of identifier in JSON Body</p>"
          },
          {
            "group": "JSON Request Body",
            "type": "String",
            "optional": false,
            "field": "identifier",
            "description": "<p>The new loginType's unique identifier.</p>"
          }
        ]
      }
    },
    "filename": "handler/http/handler.go",
    "groupTitle": "Auth",
    "success": {
      "fields": {
        "Success 200": [
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "ID",
            "description": "<p>Unique ID of the user (can be cast to long Integer).</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": false,
            "field": "type",
            "description": "<p>The <a href=\"#api-Objects-UserType\">UserType</a> of this user.</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": false,
            "field": "group",
            "description": "<p>The <a href=\"#api-Objects-Group\">group</a> the user belongs to.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "created",
            "description": "<p>The date the user was created.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "lastUpdated",
            "description": "<p>date the user was last updated.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": true,
            "field": "JWT",
            "description": "<p>JSON Web Token for accessing services. This is only provided during <a href=\"#api-Auth-Login\">Login</a>, <a href=\"#api-Auth-Register\">Registration</a> and  <a href=\"#api-Auth-FirstUser\">First User Registration</a>.</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": true,
            "field": "username",
            "description": "<p>The user's <a href=\"#api-Objects-Username\">username</a> (if this user has one).</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": true,
            "field": "phone",
            "description": "<p>The user's <a href=\"#api-Objects-VerifLogin\">phone</a> (if this user has one).</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": true,
            "field": "email",
            "description": "<p>The user's <a href=\"#api-Objects-VerifLogin\">email</a> (if this user has one).</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": true,
            "field": "facebook",
            "description": "<p>The user's <a href=\"#api-Objects-FacebookID\">facebook ID</a> (if this user has one).</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": true,
            "field": "device",
            "description": "<p>The <a href=\"#api-Objects-Device\">device</a> this user is attached to, if any.</p>"
          }
        ]
      }
    }
  },
  {
    "type": "get",
    "url": "/users/:userID",
    "title": "User Details",
    "name": "UserDetails",
    "version": "0.1.1",
    "group": "Auth",
    "permission": [
      {
        "name": "owner|^staff"
      }
    ],
    "header": {
      "fields": {
        "Header": [
          {
            "group": "Header",
            "optional": false,
            "field": "x-api-key",
            "description": "<p>the api key</p>"
          }
        ]
      }
    },
    "parameter": {
      "fields": {
        "URL Parameters": [
          {
            "group": "URL Parameters",
            "type": "String",
            "optional": false,
            "field": ":userID",
            "description": "<p>The ID of the <a href=\"#api-Objects-User\">User</a> whose details are sort.</p>"
          }
        ],
        "URL Query Parameters": [
          {
            "group": "URL Query Parameters",
            "type": "String",
            "optional": false,
            "field": "token",
            "description": "<p>The JWT provided during auth.</p>"
          }
        ]
      }
    },
    "filename": "handler/http/handler.go",
    "groupTitle": "Auth",
    "success": {
      "fields": {
        "Success 200": [
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "ID",
            "description": "<p>Unique ID of the user (can be cast to long Integer).</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": false,
            "field": "type",
            "description": "<p>The <a href=\"#api-Objects-UserType\">UserType</a> of this user.</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": false,
            "field": "group",
            "description": "<p>The <a href=\"#api-Objects-Group\">group</a> the user belongs to.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "created",
            "description": "<p>The date the user was created.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "lastUpdated",
            "description": "<p>date the user was last updated.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": true,
            "field": "JWT",
            "description": "<p>JSON Web Token for accessing services. This is only provided during <a href=\"#api-Auth-Login\">Login</a>, <a href=\"#api-Auth-Register\">Registration</a> and  <a href=\"#api-Auth-FirstUser\">First User Registration</a>.</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": true,
            "field": "username",
            "description": "<p>The user's <a href=\"#api-Objects-Username\">username</a> (if this user has one).</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": true,
            "field": "phone",
            "description": "<p>The user's <a href=\"#api-Objects-VerifLogin\">phone</a> (if this user has one).</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": true,
            "field": "email",
            "description": "<p>The user's <a href=\"#api-Objects-VerifLogin\">email</a> (if this user has one).</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": true,
            "field": "facebook",
            "description": "<p>The user's <a href=\"#api-Objects-FacebookID\">facebook ID</a> (if this user has one).</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": true,
            "field": "device",
            "description": "<p>The <a href=\"#api-Objects-Device\">device</a> this user is attached to, if any.</p>"
          }
        ]
      }
    }
  },
  {
    "type": "GET",
    "url": "/users/:userID/:loginType/verify/:OTP",
    "title": "Verify OTP",
    "description": "<p>Verify OTP sent to user's verifiable address e.g. phone, email. See <a href=\"#api-Auth-Register\">Register</a> for loginType options. extend can be set to &quot;true&quot; if intent on extending the expiry of the OTP.</p>",
    "name": "VerifyOTP",
    "version": "0.2.0",
    "group": "Auth",
    "header": {
      "fields": {
        "Header": [
          {
            "group": "Header",
            "optional": false,
            "field": "x-api-key",
            "description": "<p>the api key</p>"
          }
        ]
      }
    },
    "parameter": {
      "fields": {
        "URL Parameters": [
          {
            "group": "URL Parameters",
            "type": "String",
            "optional": false,
            "field": "userID",
            "description": "<p>The user ID of the user who needs verification.</p>"
          },
          {
            "group": "URL Parameters",
            "type": "String",
            "allowedValues": [
              "usernames",
              "emails",
              "phones",
              "facebook"
            ],
            "optional": false,
            "field": "loginType",
            "description": "<p>type of identifier to verify.</p>"
          },
          {
            "group": "URL Parameters",
            "type": "String",
            "optional": false,
            "field": "OTP",
            "description": "<p>The One Time Password sent to the user for verification.</p>"
          }
        ],
        "URL Query Parameters": [
          {
            "group": "URL Query Parameters",
            "type": "String",
            "allowedValues": [
              "true"
            ],
            "optional": false,
            "field": "extend",
            "description": "<p>set true to return an extended expiry period OTP.</p>"
          },
          {
            "group": "URL Query Parameters",
            "type": "String",
            "allowedValues": [
              "true"
            ],
            "optional": false,
            "field": "redirectToWebApp",
            "description": "<p>set true to redirect user to webApp instead of returning a JSON result.</p>"
          }
        ]
      }
    },
    "success": {
      "fields": {
        "Success 200": [
          {
            "group": "Success 200",
            "type": "String",
            "optional": true,
            "field": "OTP",
            "description": "<p>(if extending OTP) the new OTP with extended expiry</p>"
          }
        ]
      }
    },
    "filename": "handler/http/handler.go",
    "groupTitle": "Auth"
  },
  {
    "type": "GET",
    "url": "/users/:userID/verify/:OTP?loginType=:loginType&extend=:extend",
    "title": "Verify OTP",
    "description": "<p>Verify OTP. See <a href=\"#api-Auth-Register\">Register</a> for loginType and identifier options. userID is the ID of the <a href=\"#api-Objects-User\">User</a> to whom OTP was sent. extend can be set to &quot;true&quot; if intent on extending the expiry of the OTP.</p>",
    "name": "VerifyOTP",
    "version": "0.1.0",
    "group": "Auth",
    "header": {
      "fields": {
        "Header": [
          {
            "group": "Header",
            "optional": false,
            "field": "x-api-key",
            "description": "<p>the api key</p>"
          }
        ]
      }
    },
    "success": {
      "fields": {
        "200": [
          {
            "group": "200",
            "type": "String",
            "optional": false,
            "field": "OTP",
            "description": "<p>[if extending OTP] the new OTP with extended expiry</p>"
          },
          {
            "group": "200",
            "type": "Object",
            "optional": false,
            "field": "json-body",
            "description": "<p>[if not extending OTP] see <a href=\"#api-Objects-VerifLogin\">VerifLogin</a>.</p>"
          }
        ]
      }
    },
    "filename": "handler/http/_apidoc.js",
    "groupTitle": "Auth"
  },
  {
    "type": "NULL",
    "url": "Device",
    "title": "Device",
    "name": "Device",
    "version": "0.1.0",
    "group": "Objects",
    "success": {
      "fields": {
        "Success 200": [
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "ID",
            "description": "<p>Unique ID of the device (can be cast to long Integer).</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "userID",
            "description": "<p>ID for user who owns this device ID.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "deviceID",
            "description": "<p>The unique device ID string value.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "created",
            "description": "<p>ISO8601 date the device was created.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "lastUpdated",
            "description": "<p>ISO8601 date the device was last updated.</p>"
          }
        ]
      }
    },
    "filename": "handler/http/device.go",
    "groupTitle": "Objects"
  },
  {
    "type": "NULL",
    "url": "FacebookID",
    "title": "FacebookID",
    "name": "FacebookID",
    "version": "0.1.0",
    "group": "Objects",
    "success": {
      "fields": {
        "Success 200": [
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "ID",
            "description": "<p>Unique ID of the facebook ID (can be cast to long Integer).</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "userID",
            "description": "<p>ID for user who owns this facebook ID.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "facebookID",
            "description": "<p>The unique facebook ID string value.</p>"
          },
          {
            "group": "Success 200",
            "type": "Boolean",
            "optional": false,
            "field": "verified",
            "description": "<p>True if this login is verified, false otherwise.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "created",
            "description": "<p>ISO8601 date the facebook ID was inserted.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "lastUpdated",
            "description": "<p>ISO8601 date the facebook ID value was last updated.</p>"
          }
        ]
      }
    },
    "filename": "handler/http/facebook.go",
    "groupTitle": "Objects"
  },
  {
    "type": "NULL",
    "url": "Group",
    "title": "Group",
    "name": "Group",
    "version": "0.1.0",
    "group": "Objects",
    "success": {
      "fields": {
        "Success 200": [
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "ID",
            "description": "<p>Unique ID of the group (can be cast to long Integer).</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "name",
            "description": "<p>The unique group name string value.</p>"
          },
          {
            "group": "Success 200",
            "type": "Integer",
            "optional": false,
            "field": "accessLevel",
            "description": "<p>The access level for this group in (0 &gt;= accessLevel &lt;= 10)</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "created",
            "description": "<p>ISO8601 date the group was created.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "lastUpdated",
            "description": "<p>ISO8601 date the group was last updated.</p>"
          }
        ]
      }
    },
    "filename": "handler/http/group.go",
    "groupTitle": "Objects"
  },
  {
    "type": "NULL",
    "url": "OTPStatus",
    "title": "OTP Status",
    "name": "OTPStatus",
    "version": "0.1.0",
    "group": "Objects",
    "filename": "handler/http/dbt_status.go",
    "groupTitle": "Objects",
    "success": {
      "fields": {
        "Success 200": [
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "obfuscatedAddress",
            "description": "<p>Obfuscated address to which OTP was sent.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "expiresAt",
            "description": "<p>ISO8601 expiry date of OTP.</p>"
          }
        ]
      }
    }
  },
  {
    "type": "JSON",
    "url": "User",
    "title": "User",
    "name": "User",
    "version": "0.1.0",
    "group": "Objects",
    "filename": "handler/http/user.go",
    "groupTitle": "Objects",
    "success": {
      "fields": {
        "Success 200": [
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "ID",
            "description": "<p>Unique ID of the user (can be cast to long Integer).</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": false,
            "field": "type",
            "description": "<p>The <a href=\"#api-Objects-UserType\">UserType</a> of this user.</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": false,
            "field": "group",
            "description": "<p>The <a href=\"#api-Objects-Group\">group</a> the user belongs to.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "created",
            "description": "<p>The date the user was created.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "lastUpdated",
            "description": "<p>date the user was last updated.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": true,
            "field": "JWT",
            "description": "<p>JSON Web Token for accessing services. This is only provided during <a href=\"#api-Auth-Login\">Login</a>, <a href=\"#api-Auth-Register\">Registration</a> and  <a href=\"#api-Auth-FirstUser\">First User Registration</a>.</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": true,
            "field": "username",
            "description": "<p>The user's <a href=\"#api-Objects-Username\">username</a> (if this user has one).</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": true,
            "field": "phone",
            "description": "<p>The user's <a href=\"#api-Objects-VerifLogin\">phone</a> (if this user has one).</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": true,
            "field": "email",
            "description": "<p>The user's <a href=\"#api-Objects-VerifLogin\">email</a> (if this user has one).</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": true,
            "field": "facebook",
            "description": "<p>The user's <a href=\"#api-Objects-FacebookID\">facebook ID</a> (if this user has one).</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": true,
            "field": "device",
            "description": "<p>The <a href=\"#api-Objects-Device\">device</a> this user is attached to, if any.</p>"
          }
        ]
      }
    }
  },
  {
    "type": "NULL",
    "url": "UserType",
    "title": "User Type",
    "name": "UserType",
    "version": "0.1.0",
    "group": "Objects",
    "success": {
      "fields": {
        "Success 200": [
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "ID",
            "description": "<p>Unique ID of the userType (can be cast to long Integer).</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "name",
            "description": "<p>Unique name of the user type.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "created",
            "description": "<p>ISO8601 date the user type was created.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "lastUpdated",
            "description": "<p>ISO8601 date the user type was last updated.</p>"
          }
        ]
      }
    },
    "filename": "handler/http/user_type.go",
    "groupTitle": "Objects"
  },
  {
    "type": "NULL",
    "url": "Username",
    "title": "Username",
    "name": "Username",
    "version": "0.1.0",
    "group": "Objects",
    "success": {
      "fields": {
        "Success 200": [
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "ID",
            "description": "<p>Unique ID of the username (can be cast to long Integer).</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "userID",
            "description": "<p>ID for user who owns this Username.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "value",
            "description": "<p>The unique username string value.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "created",
            "description": "<p>ISO8601 date the username was created.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "lastUpdated",
            "description": "<p>ISO8601 date the username was last updated.</p>"
          }
        ]
      }
    },
    "filename": "handler/http/username.go",
    "groupTitle": "Objects"
  },
  {
    "type": "NULL",
    "url": "VerifLogin",
    "title": "Verifiable Login",
    "name": "VerifLogin",
    "version": "0.1.0",
    "group": "Objects",
    "filename": "handler/http/verifiable_login.go",
    "groupTitle": "Objects",
    "success": {
      "fields": {
        "Success 200": [
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "ID",
            "description": "<p>Unique ID of the verifiable login (can be cast to long Integer).</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "userID",
            "description": "<p>ID for user who owns this verifiable login.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "value",
            "description": "<p>The unique verifiable login string value.</p>"
          },
          {
            "group": "Success 200",
            "type": "Boolean",
            "optional": false,
            "field": "verified",
            "description": "<p>True if this login is verified, false otherwise.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "created",
            "description": "<p>ISO8601 date the verifiable login was created.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "lastUpdated",
            "description": "<p>ISO8601 date the verifiable login was last updated.</p>"
          }
        ]
      }
    }
  },
  {
    "type": "put",
    "url": "/first_user",
    "title": "First User",
    "description": "<p>Register the first super-user (super admin)</p>",
    "permission": [
      {
        "name": "anyone"
      }
    ],
    "name": "FirstUser",
    "version": "0.1.0",
    "group": "Setup",
    "header": {
      "fields": {
        "Header": [
          {
            "group": "Header",
            "optional": false,
            "field": "x-api-key",
            "description": "<p>the api key</p>"
          }
        ]
      }
    },
    "parameter": {
      "fields": {
        "JSON Request Body": [
          {
            "group": "JSON Request Body",
            "type": "String",
            "allowedValues": [
              "individual",
              "company"
            ],
            "optional": false,
            "field": "userType",
            "description": "<p>Type of user.</p>"
          },
          {
            "group": "JSON Request Body",
            "type": "String",
            "allowedValues": [
              "usernames",
              "phones",
              "emails",
              "facebook"
            ],
            "optional": false,
            "field": "loginType",
            "description": "<p>Type of identifier.</p>"
          },
          {
            "group": "JSON Request Body",
            "type": "String",
            "optional": false,
            "field": "identifier",
            "description": "<p>The user's unique loginType identifier.</p>"
          },
          {
            "group": "JSON Request Body",
            "type": "String",
            "optional": false,
            "field": "secret",
            "description": "<p>The users password</p>"
          }
        ]
      }
    },
    "success": {
      "fields": {
        "Success 201": [
          {
            "group": "Success 201",
            "type": "Object",
            "optional": false,
            "field": "json-body",
            "description": "<p>See <a href=\"#api-Objects-User\">User</a> for details.</p>"
          }
        ]
      }
    },
    "filename": "handler/http/handler.go",
    "groupTitle": "Setup"
  }
] });
