// ------------------------------------------------------------------------------------------
// General apiDoc documentation blocks and old history blocks.
// ------------------------------------------------------------------------------------------
//
// Copy all old API docs here for archival and historical retrieval


/**
 * @api {GET} /users/:userID/verify/:OTP?loginType=:loginType&extend=:extend Verify OTP
 * @apiDescription Verify OTP.
 * See <a href="#api-Auth-Register">Register</a> for loginType and identifier options.
 * userID is the ID of the <a href="#api-Objects-User">User</a> to whom OTP was sent.
 * extend can be set to "true" if intent on extending the expiry of the OTP.
 * @apiName VerifyOTP
 * @apiVersion 0.1.0
 * @apiGroup Auth
 *
 * @apiHeader x-api-key the api key
 *
 * @apiSuccess (200) {String} OTP [if extending OTP] the new OTP with extended expiry
 *
 * @apiSuccess (200) {Object} json-body [if not extending OTP] see <a href="#api-Objects-VerifLogin">VerifLogin</a>.
 *
 */

/**
 * @api {POST} /:loginType/verify/:identifier?token=:JWT Send Verification Code
 * @apiDescription Send OTP to identifier of type loginType for purpose of verifying identifier.
 * See <a href="#api-Auth-Register">Register</a> for loginType and identifier options.
 * @apiName SendVerificationCode
 * @apiVersion 0.1.0
 * @apiGroup Auth
 *
 * @apiHeader x-api-key the api key
 *
 * @apiParam {String} identifier The new 'username' corresponding to loginType
 *
 * @apiSuccess (200) {Object} json-body See <a href="#api-Objects-OTPStatus">OTPStatus</a>.
 *
 */