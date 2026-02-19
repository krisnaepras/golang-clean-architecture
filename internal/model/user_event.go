package model

type UserRegisteredEvent struct {
	ID        string `json:"id"`
	FullName  string `json:"full_name"`
	Email     string `json:"email"`
	CreatedAt int64  `json:"created_at"`
}

func (u *UserRegisteredEvent) GetId() string {
	return u.ID
}
