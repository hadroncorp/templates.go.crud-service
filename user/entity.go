package user

type User struct {
	id       string
	fullName string
}

func (u User) ID() string {
	return u.id
}

func (u User) FullName() string {
	return u.fullName
}
