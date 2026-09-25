package adminrepository

const (
	// No deleted_at filter on getUserDataSQL / getUserDetailsByIDSQL: the admin view intentionally lists soft-deleted users.
	getUserDataSQL = `
		SELECT
			u.id AS user_id,
			up.first_name,
			up.last_name,
			up.phone_number,
			up.country,
			up.city,
			up.date_of_birth,
			up.gender,
			up.avatar_url,
			up.bio,
			u.created_at,
			u.deleted_at,
			u.last_login_at,
			u.status,
			u.role
		FROM users u
		LEFT JOIN user_profiles up ON up.user_id = u.id
		ORDER BY u.created_at DESC
		LIMIT $1 OFFSET $2
	`
	existsUserSQL = `
		SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)
	`
	countUsersSQL = `
		SELECT COUNT(*) FROM users
	`
	revokeUserRoleSQL = `
		UPDATE users
		SET role  = 'user',
		updated_at = now()
		WHERE id = $1 AND role = 'subadmin'
	`
	assignUserRoleSQL = `
		UPDATE users
		SET role  = 'subadmin',
		updated_at = now()
		WHERE id = $1 AND role = 'user'
	`

	restoreUserSQL = `
		UPDATE users
		SET status = 'active',
		deleted_at = NULL,
		updated_at = now()
		WHERE id = $1 AND status = 'deleted' AND role <> 'admin'
	`
	deleteUserSQL = `
		UPDATE users
		SET status  = 'deleted',
		deleted_at = now(),
		updated_at = now()
		WHERE id = $1 AND status <> 'deleted' AND role <> 'admin'
	`

	blockUserSQL = `
		UPDATE users
		SET status  = 'blocked',
		updated_at = now()
		WHERE id = $1 AND status = 'active' AND role <> 'admin'
	`
	unblockUserSQL = `
		UPDATE users
		SET status = 'active',
		updated_at = now()
		WHERE id = $1 AND status = 'blocked' AND role <> 'admin'
	`

	getUserDetailsByIDSQL = `
		SELECT
			u.id AS user_id,
			up.first_name,
			up.last_name,
			up.phone_number,
			up.country,
			up.city,
			up.date_of_birth,
			up.gender,
			up.avatar_url,
			up.bio,
			u.created_at,
			u.deleted_at,
			u.last_login_at,
			u.status,
			u.role
		FROM users u
		LEFT JOIN user_profiles up ON up.user_id = u.id
		WHERE u.id = $1
	`
)
