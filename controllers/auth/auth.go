package auth

import (
	"encoding/json"
	"fmt"
	"goventy/config"
	"goventy/constants"
	"goventy/models"
	"goventy/utils"
	"goventy/utils/emails"
	"goventy/utils/response"
	"goventy/utils/validator"
	"net/http"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	redis "github.com/go-redis/redis/v7"
	"github.com/segmentio/ksuid"
	"golang.org/x/crypto/bcrypt"
)

func Register(w http.ResponseWriter, r *http.Request) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.ENV()["REDIS_HOST"], config.ENV()["REDIS_PORT"]),
		Password: config.ENV()["REDIS_PASSWORD"],
		DB:       0,
	})
	//get request paramters
	type Request struct {
		Email                string `json:"email"`
		Name                 string `json:"name"`
		Username             string `json:"username"`
		Password             string `json:"password"`
		PasswordConfirmation string `json:"password_confirmation"`
	}
	decoder := json.NewDecoder(r.Body)
	var request Request
	err := decoder.Decode(&request)
	if err != nil {
		utils.Log(err, "error")
	}
	name := request.Name
	username := request.Username
	email := request.Email
	password := request.Password
	password_confirmation := request.PasswordConfirmation
	//validate
	type RegistrationValidationStruct struct {
		Name                 string `validate:"required,fullname"`
		Username             string `validate:"required,username,unique=users.username"`
		Email                string `validate:"required,email,unique=users.email"`
		Password             string `validate:"required,passwd"`
		PasswordConfirmation string `json:"password_confirmation" validate:"required,eqcsfield=Password"`
	}

	registrationValidator := &RegistrationValidationStruct{
		Name:                 name,
		Username:             username,
		Email:                email,
		Password:             password,
		PasswordConfirmation: password_confirmation,
	}
	_, vErrors := validator.Validate(registrationValidator)

	if len(vErrors) != 0 {
		errorResponse := &response.Error{Status: "error", Message: "Invalid or missing input fields", Errors: vErrors}
		response.RespondBadRequest(w, errorResponse)
		return
	}

	user := models.User{}

	//hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		utils.Log(err, "error")
		errorResponse := &response.Error{Status: "error", Message: "Oops! An error occurred. Please try again later."}
		response.RespondInternalServerError(w, errorResponse)
		return
	}
	//insert into user table
	verificationToken := ksuid.New()
	user = models.User{
		Name:                   strings.Title(strings.ToLower(name)),
		Username:               username,
		Email:                  strings.ToLower(email),
		Password:               string(passwordHash),
		EmailVerificationToken: verificationToken.String(),
	}

	err = models.DB().Create(&user).Error
	if err != nil {
		utils.Log(err, "error")
		errorResponse := &response.Error{Status: "error", Message: "Oops! An error occurred. Please try again later."}
		response.RespondInternalServerError(w, errorResponse)
		return
	}

	marshalledUser, err := json.Marshal(user)
	if err != nil {
		fmt.Println(err)
	}
	//Cache non security data in Redis
	//any updates made to the user must trigger this cache keys to update
	redisUserKey := fmt.Sprintf("%s%d", constants.REDIS_USER_MARSHALLED_KEY_PREFIX, user.ID)
	redisErr := redisClient.Set(redisUserKey, marshalledUser, 0).Err()
	if redisErr != nil {
		fmt.Println(redisErr)
	}

	//we are storing only the ID to conserve memory usage
	redisErr = redisClient.HSet(constants.REDIS_LOOK_UP_KEY, user.Email, user.ID).Err()
	if redisErr != nil {
		fmt.Println(redisErr)
	}
	//create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"client_id": user.ID,
		"exp":       time.Now().Local().Add(constants.JWT_ACCESS_TOKEN_LIFETIME).Unix(),
	})

	tokenString, err := token.SignedString([]byte(config.ENV()["JWT_SECRET_KEY"]))
	if err != nil {
		utils.Log(err, "error")
		errorResponse := &response.Error{Status: "error", Message: "Oops! An error occurred. Please try again later."}
		response.RespondInternalServerError(w, errorResponse)
		return
	}
	//send user email
	type TemplateData struct {
		Name     string
		Token    string
		FrontUrl string
		Year     int
	}
	td := TemplateData{
		Name:     strings.Title(strings.ToLower(user.Name)),
		Token:    user.EmailVerificationToken,
		FrontUrl: config.ENV()["FRONTEND_URL"],
		Year:     time.Now().UTC().Year(),
	}
	mailer := &emails.Mailer{
		TemplateData: td,
		TemplatePath: "/user/welcome.html",
		To:           []string{user.Email},
		Subject:      "Welcome To goventy",
	}
	go mailer.Send()

	//send created response
	responseData := make(map[string]interface{})
	responseData["user"] = user
	responseData["token"] = tokenString

	successResponse := response.Success{
		Status:  "success",
		Message: "User registered successfully",
		Data:    responseData,
	}
	response.RespondSuccess(w, successResponse)
}

func Login(w http.ResponseWriter, r *http.Request) {
	//get request paramters
	type Request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(r.Body)
	var request Request
	errR := decoder.Decode(&request)
	if errR != nil {
		utils.Log(errR, "error")
		errorResponse := &response.Error{Status: "error", Message: "Oops! An error occurred. Please try again later."}
		response.RespondInternalServerError(w, errorResponse)
		return
	}
	email := request.Email
	password := request.Password
	type LoginValidationStruct struct {
		Email    string `validate:"required"`
		Password string `validate:"required"`
	}

	loginValidator := &LoginValidationStruct{
		Email:    email,
		Password: password,
	}
	_, vErrors := validator.Validate(loginValidator)

	if len(vErrors) != 0 {
		errorResponse := &response.Error{Status: "error", Message: "Invalid or missing input fields", Errors: vErrors}
		response.RespondBadRequest(w, errorResponse)
		return
	}

	//check if user exists
	user := models.User{}
	models.DB().Where("email = ? OR username = ?", email, email).First(&user)
	vErrors = make(map[string][]string)
	if user.Email == "" {
		vErrors["email"] = append(vErrors["email"], "The credentials you supplied do not match our records")
		errorResponse := &response.Error{Status: "error", Message: "Invalid or missing input fields", Errors: vErrors}
		response.RespondBadRequest(w, errorResponse)
		return
	}
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		//passwords do not match
		vErrors["password"] = append(vErrors["password"], "The credentials you supplied do not match our records")
		errorResponse := &response.Error{Status: "error", Message: "Invalid or missing input fields", Errors: vErrors}
		response.RespondBadRequest(w, errorResponse)
		return
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.ENV()["REDIS_HOST"], config.ENV()["REDIS_PORT"]),
		Password: config.ENV()["REDIS_PASSWORD"],
		DB:       0,
	})

	//user exists and credentials are valid
	marshalledUser, err := json.Marshal(user)
	if err != nil {
		utils.Log(err, "error")
	}
	//Cache non security data in Redis
	//any updates made to the user must trigger this cache keys to update
	redisUserKey := fmt.Sprintf("%s%d", constants.REDIS_USER_MARSHALLED_KEY_PREFIX, user.ID)
	redisErr := redisClient.Set(redisUserKey, marshalledUser, 0).Err()
	if redisErr != nil {
		utils.Log(redisErr, "error")
	}

	//we are storing only the ID to conserve memory usage
	redisErr = redisClient.HSet(constants.REDIS_LOOK_UP_KEY, user.Email, user.ID).Err()
	if redisErr != nil {
		utils.Log(redisErr, "error")
	}
	//create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"client_id": user.ID,
		"exp":       time.Now().Local().Add(constants.JWT_ACCESS_TOKEN_LIFETIME).Unix(),
	})

	tokenString, err := token.SignedString([]byte(config.ENV()["JWT_SECRET_KEY"]))
	if err != nil {
		utils.Log(err, "error")
	}
	responseData := make(map[string]interface{})
	responseData["user"] = user
	responseData["token"] = tokenString

	successResponse := response.Success{
		Status:  "success",
		Message: "User logged in successfully",
		Data:    responseData,
	}

	response.RespondSuccess(w, successResponse)
}
