package postgres

// SQL statements for users
const (
	insertUser = `INSERT INTO users(name, age, email) VALUES ($1, $2, $3) RETURNING id`

	updateUser = `UPDATE users SET name = $1, age = $2, email = $3 WHERE id = $4`

	selectUser = `SELECT id, name, age, email FROM users WHERE id = $1`
)

// SQL statements for message relay
const (
	selectPurchaseMessages = `
		SELECT
			id,
			topic,
			partition_key,
			headers,
			"value"
		FROM outbox_messages
		WHERE
			delivered_at IS NULL
			AND
			deleted_at IS NULL
		ORDER BY created_at ASC
		LIMIT $1
	`
)
