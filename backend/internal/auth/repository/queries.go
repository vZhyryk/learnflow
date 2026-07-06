package authrepository

const (
	sessionColumns = `
    id, user_id, refresh_hash, user_agent, ip, expires_at, revoked_at, revoke_reason, revoked_by_user_id,
	created_at, failed_attempt_count, last_attempt_at, locked_until, token_version, previous_refresh_hash,
	last_seen_at, last_seen_ip `

	/* User Session */
	createUserSessionSQL = `
		INSERT INTO user_sessions (user_id, refresh_hash, user_agent, ip, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING` + sessionColumns

	getActiveUserSessionSQL = `
		SELECT` + sessionColumns + `
		FROM user_sessions WHERE user_id = $1 AND revoked_at IS NULL
		ORDER BY created_at DESC
		LIMIT 25
	`

	getAllUserSessionSQL = `
		SELECT` + sessionColumns + `
		FROM user_sessions WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 25
	`

	getSessionByTokenSQL = `
		SELECT` + sessionColumns + `
		FROM user_sessions WHERE refresh_hash = $1 AND revoked_at IS NULL AND expires_at > now()
		FOR UPDATE NOWAIT
	`

	// Intentionally no revoked_at IS NULL filter — must find even revoked sessions
	// to detect token reuse attacks (sliding window refresh pattern).
	getSessionByPrevHashSQL = `
		SELECT` + sessionColumns + `
		FROM user_sessions WHERE previous_refresh_hash = $1
	`

	revokeAllUserSessionsSQL = `
		UPDATE user_sessions
		SET revoked_at = now(), revoke_reason = $1, revoked_by_user_id = $2
		WHERE user_id = $3 AND revoked_at IS NULL
	`
	revokeUserSessionSQL = `
		UPDATE user_sessions
		SET revoked_at = now(), revoke_reason = $1, revoked_by_user_id = $2
		WHERE id = $3 AND revoked_at IS NULL
	`
	updateSessionTokenSQL = `
		UPDATE user_sessions
		SET previous_refresh_hash = refresh_hash,
		refresh_hash = $2,
		user_agent = $3,
		token_version = token_version + 1,
		last_attempt_at = now(),
		last_seen_at = now(),
		last_seen_ip = $4
		WHERE id = $1 AND revoked_at IS NULL AND expires_at > now()
	`

	updateFailedLoginAttemptsSQL = `
		UPDATE user_sessions
		SET failed_attempt_count = failed_attempt_count + 1,
		last_attempt_at = now(),
		locked_until = CASE
			WHEN failed_attempt_count + 1 >= $2
			THEN now() + $3::interval
			ELSE locked_until
		END
		WHERE id = $1
	`

	/* Token */
	tokenColumns = ` id, user_id, token_hash, expires_at, created_at, used_at, invalidated_at, invalidated_by_user_id `
	// Create Token
	createAccountRecoveryTokenSQL = `
		INSERT INTO account_recovery_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING` + tokenColumns

	createEmailChangeTokenSQL = `
		INSERT INTO email_change_tokens (user_id, token_hash, expires_at, new_email)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, new_email, token_hash, expires_at, created_at, used_at, invalidated_at, invalidated_by_user_id
	`
	createEmailVerificationTokenSQL = `
		INSERT INTO email_verification_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING` + tokenColumns

	createPasswordResetTokenSQL = `
		INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING` + tokenColumns

	// Get Token by Hash
	getAccountRecoveryTokenByHashSQL = `
		SELECT` + tokenColumns + `FROM account_recovery_tokens WHERE token_hash = $1 AND used_at IS NULL AND expires_at > now() AND invalidated_at IS NULL FOR UPDATE NOWAIT`
	getEmailChangeTokenByHashSQL = `
		SELECT id, user_id, new_email, token_hash, expires_at, created_at, used_at, invalidated_at, invalidated_by_user_id FROM email_change_tokens WHERE token_hash = $1 AND used_at IS NULL AND expires_at > now() AND invalidated_at IS NULL FOR UPDATE NOWAIT`
	getEmailVerificationTokenByHashSQL = `
		SELECT` + tokenColumns + `FROM email_verification_tokens WHERE token_hash = $1 AND used_at IS NULL AND expires_at > now() AND invalidated_at IS NULL FOR UPDATE NOWAIT`
	getPasswordResetTokenByHashSQL = `
		SELECT` + tokenColumns + `FROM password_reset_tokens WHERE token_hash = $1 AND used_at IS NULL AND expires_at > now() AND invalidated_at IS NULL FOR UPDATE NOWAIT`

	// Mark Token as Used
	markAccountRecoveryTokenUsedSQL = `
		UPDATE account_recovery_tokens
		SET used_at = now()
		WHERE token_hash = $1 AND used_at IS NULL
	`
	markEmailChangeTokenUsedSQL = `
		UPDATE email_change_tokens
		SET used_at = now()
		WHERE token_hash = $1 AND used_at IS NULL
	`
	markEmailVerificationTokenUsedSQL = `
		UPDATE email_verification_tokens
		SET used_at = now()
		WHERE token_hash = $1 AND used_at IS NULL
	`
	markPasswordResetTokenUsedSQL = `
		UPDATE password_reset_tokens
		SET used_at = now()
		WHERE token_hash = $1 AND used_at IS NULL
	`

	deleteExpiredEmailVerificationTokensSQL = `DELETE FROM email_verification_tokens WHERE expires_at <= now()`
	deleteExpiredPasswordResetTokensSQL     = `DELETE FROM password_reset_tokens WHERE expires_at <= now()`
	deleteExpiredEmailChangeTokensSQL       = `DELETE FROM email_change_tokens WHERE expires_at <= now()`
	deleteExpiredAccountRecoveryTokensSQL   = `DELETE FROM account_recovery_tokens WHERE expires_at <= now()`

	/* User */
	createUserSQL = `
		INSERT INTO users (email, password_hash, role)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	createUserProfileSQL = `
		INSERT INTO user_profiles (
			user_id,
			first_name,
			last_name,
  			phone_number,
    		country,
    		city,
    		date_of_birth,
    		gender,
    		ui_language,
    		avatar_url,
    		timezone,
    		bio
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	getUserProfileByUserIDSQL = `
		SELECT
			user_id,
			first_name,
			last_name,
			phone_number,
			country,
			city,
			date_of_birth,
			gender,
			ui_language,
			avatar_url,
			timezone,
			bio,
			created_at,
			updated_at
		FROM user_profiles WHERE user_id = $1
	`

	deleteUserSQL = `
		UPDATE users
		SET status = 'deleted',
		deleted_at = now(),
		updated_at = now()
		WHERE id = $1
	`

	userColumns = `
	id, email, password_hash, role, status, email_verified_at, last_login_at, deleted_at, created_at, updated_at,
	password_changed_at, email_changed_at, failed_login_count, last_failed_login_at, login_locked_until `

	getUserByEmailSQL = `
		SELECT` + userColumns + `FROM users WHERE LOWER(email) = LOWER($1) AND deleted_at IS NULL`

	getDeletedUserByEmailSQL = `
		SELECT` + userColumns + `FROM users WHERE LOWER(email) = LOWER($1) AND deleted_at IS NOT NULL`

	getUserByIDSQL = `
		SELECT ` + userColumns + ` FROM users WHERE id = $1 AND deleted_at IS NULL`

	getDeletedUserByIDSQL = `
		SELECT` + userColumns + `FROM users WHERE id = $1 AND deleted_at IS NOT NULL`

	updateLastLoginSQL = `
		UPDATE users
		SET last_login_at  = now()
		WHERE id = $1
	`
	updateEmailSQL = `
		UPDATE users
		SET email  = $1,
		email_changed_at = now(),
		updated_at = now()
		WHERE id = $2
	`
	updatePasswordSQL = `
		UPDATE users
		SET password_hash  = $1,
		password_changed_at = now(),
		updated_at = now()
		WHERE id = $2
	`
	updateUserRoleSQL = `
		UPDATE users
		SET role  = $1,
		updated_at = now()
		WHERE id = $2
	`
	updateUserStatusSQL = `
		UPDATE users
		SET status  = $1,
		updated_at = now()
		WHERE id = $2
	`

	restoreUserSQL = `
		UPDATE users
		SET status = 'active',
		deleted_at = NULL,
		updated_at = now()
		WHERE id = $1
	`

	updateEmailVerifiedAtSQL = `
		UPDATE users
		SET email_verified_at = now(),
		updated_at = now()
		WHERE id = $1
	`

	incrementFailedLoginSQL = `
		UPDATE users
		SET
			failed_login_count   = failed_login_count + 1,
			last_failed_login_at = now(),
			login_locked_until   = CASE
				WHEN failed_login_count + 1 >= $2
				THEN now() + $3::interval
				ELSE login_locked_until
			END
		WHERE id = $1
	`

	resetFailedLoginSQL = `
		UPDATE users
		SET
			failed_login_count   = 0,
			last_failed_login_at = NULL,
			login_locked_until   = NULL
		WHERE id = $1
	`
)
