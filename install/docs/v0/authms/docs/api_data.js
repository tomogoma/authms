define({ "api": [
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
            "description": "<p>Basic auth containing identifier/secret, both provided during <a href=\"#api-Auth-Register\">Registration</a></p>"
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
            "description": "<p>See <a href=\"#api-Objects-User\">User</a>.</p>"
          }
        ]
      }
    },
    "filename": "handler/http/handler.go",
    "groupTitle": "Auth"
  },
  {
    "type": "put",
    "url": "/:loginType/register?selfReg=:selfReg",
    "title": "Register",
    "description": "<p>Register new user. Registration can be:</p> <ul> <li>self registration - provide URL param selfReg=true</li> <li>self registration by unique device ID - provide URL param selfReg=device</li> <li>or other user (by admin) - don't provide URL params</li> </ul> <p>loginType is what the user will be logging in by, can be one of:</p> <ul> <li>usernames</li> <li>emails</li> <li>phones</li> <li>facebook</li> </ul>",
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
        "Parameter": [
          {
            "group": "Parameter",
            "type": "Enum",
            "optional": false,
            "field": "userType",
            "description": "<p>Type of user [individual|company]</p>"
          },
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "identifier",
            "description": "<p>The 'username' corresponding to loginType</p>"
          },
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "secret",
            "description": "<p>The users password</p>"
          },
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "groupID",
            "description": "<p>[only if selfReg not set or false] groupID to add this user to</p>"
          },
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "deviceID",
            "description": "<p>[only if selfReg set to device] the unique device ID for the user</p>"
          }
        ]
      }
    },
    "success": {
      "fields": {
        "201": [
          {
            "group": "201",
            "type": "Object",
            "optional": false,
            "field": "json-body",
            "description": "<p>See <a href=\"#api-Objects-User\">User</a>.</p>"
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
    "description": "<p>Send Password reset Code (OTP) to identifier of type loginType. See <a href=\"#api-Auth-Register\">Register</a> for loginType and identifier options.</p>",
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
        "Parameter": [
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "loginType",
            "description": "<p>See <a href=\"#api-Auth-Register\">Register</a> for loginType options.</p>"
          },
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "identifier",
            "description": "<p>See <a href=\"#api-Auth-Register\">Register</a> for identifier options.</p>"
          },
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "OTP",
            "description": "<p>The password reset code sent to user during <a href=\"#api-Auth-SendPasswordResetOTP\">SendPasswordResetOTP</a>.</p>"
          },
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "newSecret",
            "description": "<p>The new password.</p>"
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
            "description": "<p>See <a href=\"#api-Objects-VerifLogin\">VerifLogin</a>.</p>"
          }
        ]
      }
    },
    "filename": "handler/http/handler.go",
    "groupTitle": "Auth"
  },
  {
    "type": "POST",
    "url": "/reset_password/send_otp",
    "title": "Send Password Reset OTP",
    "description": "<p>Send Password reset Code (OTP) to identifier of type loginType. See <a href=\"#api-Auth-Register\">Register</a> for loginType and identifier options.</p>",
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
        "Parameter": [
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "loginType",
            "description": "<p>See <a href=\"#api-Auth-Register\">Register</a> for loginType options.</p>"
          },
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "identifier",
            "description": "<p>See <a href=\"#api-Auth-Register\">Register</a> for identifier options.</p>"
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
    "filename": "handler/http/handler.go",
    "groupTitle": "Auth"
  },
  {
    "type": "POST",
    "url": "/:loginType/verify?token=:JWT",
    "title": "Send Verification Code",
    "description": "<p>Send OTP to identifier of type loginType for purpose of verifying identifier. See <a href=\"#api-Auth-Register\">Register</a> for loginType options.</p>",
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
        "Parameter": [
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "identifier",
            "description": "<p>The loginType's address to be verified.</p>"
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
    "filename": "handler/http/handler.go",
    "groupTitle": "Auth"
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
        "200": [
          {
            "group": "200",
            "type": "String",
            "optional": false,
            "field": "name",
            "description": "<p>Micro-service name.</p>"
          },
          {
            "group": "200",
            "type": "String",
            "optional": false,
            "field": "version",
            "description": "<p>Current running version.</p>"
          },
          {
            "group": "200",
            "type": "String",
            "optional": false,
            "field": "description",
            "description": "<p>Short description of the micro-service.</p>"
          },
          {
            "group": "200",
            "type": "String",
            "optional": false,
            "field": "canonicalName",
            "description": "<p>Canonical name of the micro-service.</p>"
          },
          {
            "group": "200",
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
    "url": "/:loginType/update?token=:JWT",
    "title": "Update Identifier",
    "description": "<p>Update (or set for first time) the identifier details for loginType. See <a href=\"#api-Auth-Register\">Register</a> for loginType. See <a href=\"#api-Objects-User\">User</a> for how to access the JWT.</p>",
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
            "description": "<p>See <a href=\"#api-Objects-User\">User</a>.</p>"
          }
        ]
      }
    },
    "filename": "handler/http/handler.go",
    "groupTitle": "Auth"
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
        "URL Path Component": [
          {
            "group": "URL Path Component",
            "type": "String",
            "optional": false,
            "field": ":userID",
            "description": "<p>The ID of the <a href=\"#api-Objects-User\">User</a> whose details are sort.</p>"
          }
        ],
        "URL Query": [
          {
            "group": "URL Query",
            "type": "String",
            "optional": false,
            "field": "token",
            "description": "<p>The JWT provided during auth.</p>"
          }
        ]
      }
    },
    "success": {
      "fields": {
        "200": [
          {
            "group": "200",
            "type": "JSON",
            "optional": false,
            "field": "body",
            "description": "<p>The details of the <a href=\"#api-Objects-User\">User</a> in JSON format.</p>"
          }
        ]
      }
    },
    "filename": "handler/http/handler.go",
    "groupTitle": "Auth"
  },
  {
    "type": "POST",
    "url": "/:loginType/verify/:OTP?extend=:extend",
    "title": "Verify OTP",
    "description": "<p>Verify OTP. See <a href=\"#api-Auth-Register\">Register</a> for loginType options. extend can be set to &quot;true&quot; if intent on extending the expiry of the OTP.</p>",
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
        "Parameter": [
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "identifier",
            "description": "<p>The loginType's address to whom the OTP was sent.</p>"
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
    },
    "filename": "handler/http/dbt_status.go",
    "groupTitle": "Objects"
  },
  {
    "type": "NULL",
    "url": "User",
    "title": "User",
    "name": "User",
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
            "description": "<p>Unique ID of the user (can be cast to long Integer).</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "JWT",
            "description": "<p>JSON Web Token for accessing services. This is only provided during <a href=\"#api-Auth-Login\">Login</a>.</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": false,
            "field": "username",
            "description": "<p>See <a href=\"#api-Objects-Username\">Username</a>.</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": false,
            "field": "phone",
            "description": "<p>See <a href=\"#api-Objects-VerifLogin\">VerifLogin</a>.</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": false,
            "field": "email",
            "description": "<p>See <a href=\"#api-Objects-VerifLogin\">VerifLogin</a>.</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": false,
            "field": "facebook",
            "description": "<p>See <a href=\"#api-Objects-FacebookID\">FacebookID</a>.</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": false,
            "field": "group",
            "description": "<p>See <a href=\"#api-Objects-Group\">Group</a>.</p>"
          },
          {
            "group": "Success 200",
            "type": "Object",
            "optional": false,
            "field": "device",
            "description": "<p>See <a href=\"#api-Objects-Device\">Device</a>.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "created",
            "description": "<p>Date the user was created.</p>"
          },
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "lastUpdated",
            "description": "<p>date the user was last updated.</p>"
          }
        ]
      }
    },
    "filename": "handler/http/user.go",
    "groupTitle": "Objects"
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
    },
    "filename": "handler/http/verifiable_login.go",
    "groupTitle": "Objects"
  },
  {
    "type": "put",
    "url": "/first_user",
    "title": "First User",
    "description": "<p>Register the first super-user (super admin)</p>",
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
        "Parameter": [
          {
            "group": "Parameter",
            "type": "Enum",
            "optional": false,
            "field": "userType",
            "description": "<p>Type of user [individual|company]</p>"
          },
          {
            "group": "Parameter",
            "type": "Enum",
            "optional": false,
            "field": "loginType",
            "description": "<p>Type of identifier [usernames|emails|phones|facebook]</p>"
          },
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "identifier",
            "description": "<p>The 'username' corresponding to loginType</p>"
          },
          {
            "group": "Parameter",
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
        "201": [
          {
            "group": "201",
            "type": "Object",
            "optional": false,
            "field": "json-body",
            "description": "<p>See <a href=\"#api-Objects-User\">User</a>.</p>"
          }
        ]
      }
    },
    "filename": "handler/http/handler.go",
    "groupTitle": "Setup"
  }
] });
