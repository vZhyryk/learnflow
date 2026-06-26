package usersrepository

const (
	getProfileByUserIDSQL = `
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
		FROM user_profiles
		WHERE user_id = $1
  `

	updateProfileSQL = `
		UPDATE user_profiles
		SET
			first_name = $2,
			last_name = $3,
			phone_number = $4,
			country = $5,
			city = $6,
			date_of_birth = $7,
			gender = $8,
			ui_language = $9,
			avatar_url = $10,
			timezone = $11,
			bio = $12,
			updated_at = now()
		WHERE user_id = $1
  `
)
