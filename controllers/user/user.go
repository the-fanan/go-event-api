package user

import (
	_ "encoding/json"
	"fmt"
	_ "goventy/config"
	"goventy/models"
	"goventy/utils"
	"goventy/utils/response"
	_ "goventy/utils/storage"
	"goventy/utils/validator"
	_ "mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/biezhi/gorm-paginator/pagination"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

func Events(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	user_id_from_token := ""
	authorization := r.Header.Get("Authorization")
	if authorization != "" {
		token, err := utils.GetJwtToken(strings.Split(authorization, " ")[1])
		if err != nil {
			errorResponse := &response.Error{Status: "error", Message: "Oops! Token has expired."}
			response.RespondUnauthorized(w, errorResponse)
			return
		}
		claims, _ := token.Claims.(jwt.MapClaims)
		user_id_from_token = fmt.Sprintf("%d", int(claims["client_id"].(float64)))
	}

	var page int
	var limit int
	name := r.URL.Query().Get("name")
	pageString := r.URL.Query().Get("page")
	if pageString == "" {
		page = 1
	} else {
		lpage, err := strconv.Atoi(pageString)
		page = lpage
		if err != nil {
			utils.Log(err, "error")
			errorResponse := &response.Error{Status: "error", Message: "Oops! An error occurred. Please try again later."}
			response.RespondInternalServerError(w, errorResponse)
			return
		}
	}
	limitString := r.URL.Query().Get("limit")
	if limitString == "" {
		limit = 20
	} else {
		llimit, err := strconv.Atoi(limitString)
		limit = llimit
		if err != nil {
			utils.Log(err, "error")
			errorResponse := &response.Error{Status: "error", Message: "Oops! An error occurred. Please try again later."}
			response.RespondInternalServerError(w, errorResponse)
			return
		}
	}

	type RequesrValidationStruct struct {
		UserId string `validate:"required,numeric,exists=users.id"`
	}

	uvs := &RequesrValidationStruct{
		UserId: params["id"],
	}

	_, vErrors := validator.Validate(uvs)
	//validate IsVailable is a number
	if len(vErrors) != 0 {
		errorResponse := &response.Error{Status: "error", Message: "Invalid or missing input fields", Errors: vErrors}
		response.RespondBadRequest(w, errorResponse)
		return
	}

	events := make([]models.Event, 0)
	var eventPaginator *pagination.Paginator
	var paginationQuery *gorm.DB
	//if it's the user querying their own events, show them even ones that are not available so they'll be edited
	//else, only show available events
	if params["id"] != user_id_from_token {
		paginationQuery = models.DB().Preload("Images").Preload("Creator").Where("name LIKE ? AND is_available = ? AND user_id = ?", name+"%", 1, params["id"])
	} else {
		paginationQuery = models.DB().Preload("Images").Preload("Creator").Where("name LIKE ? AND user_id = ?", name+"%", params["id"])
	}

	eventPaginator = pagination.Paging(&pagination.Param{
		DB:      paginationQuery,
		Page:    page,
		Limit:   limit,
		OrderBy: []string{"created_at desc"},
		ShowSQL: true,
	}, &events)

	responseData := make(map[string]interface{})
	responseData["events"] = eventPaginator
	successResponse := response.Success{
		Status:  "success",
		Message: "Events retrieved successfully",
		Data:    responseData,
	}

	response.RespondSuccess(w, successResponse)
}
