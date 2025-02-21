package user

type postgresUser struct {
	UserID   string `db:"user_id"`
	FullName string `db:"full_name"`
}

func (u postgresUser) toEntity() User {
	return User{
		id:       u.UserID,
		fullName: u.FullName,
	}
}
