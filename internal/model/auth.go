package model

type Auth struct {
	ID    string
	Roles []string // kode role, contoh: ["ADMIN", "MANAGER"]
}
