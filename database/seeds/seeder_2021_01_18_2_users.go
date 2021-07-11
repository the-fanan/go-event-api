package seeds

import (
	"goventy/models"

	"github.com/segmentio/ksuid"
	"golang.org/x/crypto/bcrypt"
)

func Seeder_2021_01_18_2_users() {
	users := []map[string]string{
		{"name": "Test User", "email": "test@user.com", "username": "testUser1", "password": "test123"},
	}

	for i := 0; i < len(users); i++ {
		passwordHash, _ := bcrypt.GenerateFromPassword([]byte(users[i]["password"]), bcrypt.DefaultCost)
		verificationToken := ksuid.New()
		user := models.User{
			Name:                   users[i]["name"],
			Username:               users[i]["username"],
			Email:                  users[i]["email"],
			Password:               string(passwordHash),
			EmailVerified:          1,
			EmailVerificationToken: verificationToken.String(),
		}
		models.DB().Create(&user)
	}
}
