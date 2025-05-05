package auth

type User struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

var MockUsers = map[string]User{
	"admin": {
		UserID:   "user123",
		Username: "admin",
		Password: "password",
	},
	"user": {
		UserID:   "user456",
		Username: "user",
		Password: "password",
	},
	"guest": {
		UserID:   "user789",
		Username: "guest",
		Password: "password",
	},
	"test": {
		UserID:   "user101",
		Username: "test",
		Password: "password",
	},
}
