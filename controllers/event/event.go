package event

import (
	"encoding/json"
	_ "goventy/config"
	"goventy/models"
	"goventy/utils"
	"goventy/utils/response"
	"goventy/utils/storage"
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

func Create(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(30 * 1024 * 1024) // grab the multipart form
	if err != nil {
		errorResponse := &response.Error{Status: "error", Message: "Request body is too large"}
		response.RespondRequestTooLarge(w, errorResponse)
		return
	}

	token, err := utils.GetJwtToken(strings.Split(r.Header.Get("Authorization"), " ")[1])
	if err != nil {
		utils.Log(err, "error")
		errorResponse := &response.Error{Status: "error", Message: "Oops! An error occurred. Please try again later."}
		response.RespondInternalServerError(w, errorResponse)
		return
	}
	claims, _ := token.Claims.(jwt.MapClaims)
	user_id := uint(claims["client_id"].(float64))

	formdata := r.MultipartForm

	type TimeValidationStruct struct {
		Time string `validate:"required,timeInterval"`
	}

	type EventValidationStruct struct {
		Name        string                  `validate:"required"`
		Description string                  `validate:"required"`
		StartDate   string                  `json:"start_date" validate:"required,date"`
		EndDate     string                  `json:"end_date" validate:"required,date,aftdtfield=StartDate"`
		Times       []*TimeValidationStruct `validate:"required,dive,required"`
	}

	tvsa := make([]*TimeValidationStruct, 0)
	for _, time := range formdata.Value["times"] {
		tvsa = append(tvsa, &TimeValidationStruct{Time: time})
	}
	evs := &EventValidationStruct{
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
		StartDate:   r.FormValue("start_date"),
		EndDate:     r.FormValue("end_date"),
		Times:       tvsa,
	}
	_, vErrors := validator.Validate(evs)
	imageErrors := validator.ValidateMultipleFormUploadedImage(formdata.File["images"])
	if imageErrors != nil {
		if vErrors == nil {
			vErrors = make(map[string][]string)
			vErrors["images"] = imageErrors
		} else {
			vErrors["images"] = imageErrors
		}
	}

	if len(vErrors) != 0 {
		errorResponse := &response.Error{Status: "error", Message: "Invalid or missing input fields", Errors: vErrors}
		response.RespondBadRequest(w, errorResponse)
		return
	}

	eventTimes, err := json.Marshal(formdata.Value["times"])
	if err != nil {
		errorResponse := &response.Error{Status: "error", Message: "Invalid or missing input fields", Errors: err}
		response.RespondBadRequest(w, errorResponse)
		return
	}
	event := &models.Event{
		UserID:      user_id,
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
		Address:     r.FormValue("address"),
		StartDate:   r.FormValue("start_date"),
		EndDate:     r.FormValue("end_date"),
		Times:       string(eventTimes),
	}

	//get images for event if any
	eventImages := formdata.File["images"]
	eventImagesStructs := make([]models.Image, len(eventImages))
	for i, _ := range eventImages {
		file, err := eventImages[i].Open()
		defer file.Close()
		if err != nil {
			utils.Log(err, "error")
			return
		}
		localStore := &storage.LocalStorage{
			Folder: "uploads/events/images",
		}
		uploadedFile := localStore.Create(file)
		image := &models.Image{
			UserID:     user_id,
			Url:        uploadedFile.Url,
			Provider:   uploadedFile.Provider,
			ProviderID: uploadedFile.ProviderID,
		}
		eventImagesStructs[i] = *image
	}
	event.Images = eventImagesStructs
	//add tags to event
	tags := formdata.Value["tags"]
	taggables := make([]models.Taggable, len(tags))
	for index, tag := range tags {
		tagId, _ := strconv.ParseUint(tag, 10, 64)
		taggable := &models.Taggable{
			TagID: uint(tagId),
		}
		taggables[index] = *taggable
	}
	event.Taggables = taggables
	//create event
	err = models.DB().Create(event).Error
	if err != nil {
		utils.Log(err.Error(), "error")
		errorResponse := &response.Error{Status: "error", Message: "Oops! An error occurred. Please try again later."}
		response.RespondInternalServerError(w, errorResponse)
		return
	}

	//send success response
	responseData := make(map[string]interface{})
	responseData["event"] = event
	successResponse := response.Success{
		Status:  "success",
		Message: "Event created successfully",
		Data:    responseData,
	}
	response.RespondSuccess(w, successResponse)
}

func Find(w http.ResponseWriter, r *http.Request) {
	var err error
	var page int
	var limit int
	name := r.URL.Query().Get("name")
	pageString := r.URL.Query().Get("page")
	if pageString == "" {
		page = 1
	} else {
		page, err = strconv.Atoi(pageString)
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
		limit, err = strconv.Atoi(pageString)
		if err != nil {
			utils.Log(err, "error")
			errorResponse := &response.Error{Status: "error", Message: "Oops! An error occurred. Please try again later."}
			response.RespondInternalServerError(w, errorResponse)
			return
		}
	}

	events := make([]models.Event, 0)

	var eventPaginator *pagination.Paginator
	var paginationQuery *gorm.DB
	if name == "" {
		paginationQuery = models.DB().Preload("Images").Preload("Creator").Where(" is_available = ?", 1)
	} else {
		paginationQuery = models.DB().Preload("Images").Preload("Creator").Where("name LIKE ? AND is_available = ?", name+"%", 1)
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

func Get(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	vErrors := make(map[string][]string)

	id, err := strconv.ParseUint(params["id"], 10, 64)
	if err != nil {
		vErrors["id"] = []string{"Invalid ID passed"}
	}

	event := &models.Event{}
	err = models.DB().Preload("Images").Preload("Creator").Preload("Tickets.Images").Preload("Presenters.Images").First(event, id).Error
	if err != nil {
		utils.Log(err.Error(), "error")
		errorResponse := &response.Error{Status: "error", Message: "Oops! An error occurred. Please try again later."}
		response.RespondInternalServerError(w, errorResponse)
		return
	}
	errorResponse := &response.Error{Status: "error", Message: "Invalid or missing input fields", Errors: vErrors}

	if len(vErrors) != 0 {
		response.RespondBadRequest(w, errorResponse)
		return
	}

	responseData := make(map[string]interface{})
	responseData["event"] = event
	successResponse := response.Success{
		Status:  "success",
		Message: "Event retrieved successfully",
		Data:    responseData,
	}

	response.RespondSuccess(w, successResponse)
}

func Update(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	token, err := utils.GetJwtToken(strings.Split(r.Header.Get("Authorization"), " ")[1])
	if err != nil {
		utils.Log(err, "error")
		errorResponse := &response.Error{Status: "error", Message: "Oops! An error occurred. Please try again later."}
		response.RespondInternalServerError(w, errorResponse)
		return
	}
	claims, _ := token.Claims.(jwt.MapClaims)
	user_id := uint(claims["client_id"].(float64))

	type Request struct {
		Name        string   `json:"name"`
		Times       []string `json:"times"`
		Description string   `json:"description"`
		Address     string   `json:"address"`
		StartDate   string   `json:"start_date"`
		EndDate     string   `json:"end_date"`
		IsAvailable int      `json:"is_available"`
	}

	decoder := json.NewDecoder(r.Body)
	var request Request
	err = decoder.Decode(&request)
	if err != nil {
		vErrors := make(map[string][]string)
		vErrors["json_decode"] = []string{err.Error()}
		errorResponse := &response.Error{Status: "error", Message: "Invalid data types supplied", Errors: vErrors}
		response.RespondBadRequest(w, errorResponse)
		return
	}

	type TimeValidationStruct struct {
		Time string `validate:"required,timeInterval"`
	}

	type EventValidationStruct struct {
		EventId     string                  `validate:"required,exists=events.id"`
		StartDate   string                  `json:"start_date" validate:"required_with=EndDate,date"`
		EndDate     string                  `json:"end_date" validate:"required_with=StartDate,date,aftdtfield=StartDate"`
		IsAvailable int                     `json:"is_available" validate:"numeric"`
		Times       []*TimeValidationStruct `validate:"required,dive,required"`
	}

	tvsa := make([]*TimeValidationStruct, 0)
	for _, time := range request.Times {
		tvsa = append(tvsa, &TimeValidationStruct{Time: time})
	}
	evs := &EventValidationStruct{
		EventId:     params["id"],
		StartDate:   request.StartDate,
		EndDate:     request.EndDate,
		IsAvailable: request.IsAvailable,
		Times:       tvsa,
	}
	_, vErrors := validator.Validate(evs)
	//validate IsVailable is a number
	if len(vErrors) != 0 {
		errorResponse := &response.Error{Status: "error", Message: "Invalid or missing input fields", Errors: vErrors}
		response.RespondBadRequest(w, errorResponse)
		return
	}

	id, err := strconv.ParseUint(params["id"], 10, 64)
	if err != nil {
		vErrors["id"] = []string{"Invalid ID passed"}
	}

	event := &models.Event{}
	err = models.DB().Preload("Images").Preload("Creator").Preload("Tickets.Images").Preload("Presenters.Images").First(event, id).Error
	if err != nil {
		utils.Log(err.Error(), "error")
		errorResponse := &response.Error{Status: "error", Message: "Oops! An error occurred. Please try again later."}
		response.RespondInternalServerError(w, errorResponse)
		return
	}

	//validate that user has permissions to edit this event
	if event.UserID != user_id {
		errorResponse := &response.Error{Status: "error", Message: "You do not have permission to carry out this action"}
		response.RespondUnauthorized(w, errorResponse)
		return
	}

	//make sure time is correct for database
	type RequestDB struct {
		Name        string `json:"name"`
		Times       string `json:"times"`
		Description string `json:"description"`
		Address     string `json:"address"`
		StartDate   string `json:"start_date"`
		EndDate     string `json:"end_date"`
		IsAvailable int    `json:"is_available"`
	}
	eventTimes, _ := json.Marshal(request.Times)
	requestDB := &RequestDB{
		Name:        request.Name,
		Description: request.Description,
		Address:     request.Address,
		StartDate:   request.StartDate,
		EndDate:     request.EndDate,
		IsAvailable: request.IsAvailable,
	}
	requestDB.Times = string(eventTimes)
	//update values
	err = models.DB().Model(event).Updates(requestDB).Error
	if err != nil {
		utils.Log(err.Error(), "error")
		errorResponse := &response.Error{Status: "error", Message: "Oops! An error occurred. Please try again later."}
		response.RespondInternalServerError(w, errorResponse)
		return
	}
	//handle zero values not handled by Updates
	err = models.DB().Model(event).Update("is_available", request.IsAvailable).Error
	if err != nil {
		utils.Log(err.Error(), "error")
		errorResponse := &response.Error{Status: "error", Message: "Oops! An error occurred. Please try again later."}
		response.RespondInternalServerError(w, errorResponse)
		return
	}

	responseData := make(map[string]interface{})
	responseData["event"] = event
	successResponse := response.Success{
		Status:  "success",
		Message: "Event updated successfully",
		Data:    responseData,
	}

	response.RespondSuccess(w, successResponse)
}

func Tickets(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var err error
	var page int
	var limit int
	name := r.URL.Query().Get("name")
	pageString := r.URL.Query().Get("page")
	if pageString == "" {
		page = 1
	} else {
		page, err = strconv.Atoi(pageString)
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
		limit, err = strconv.Atoi(limitString)
		if err != nil {
			utils.Log(err, "error")
			errorResponse := &response.Error{Status: "error", Message: "Oops! An error occurred. Please try again later."}
			response.RespondInternalServerError(w, errorResponse)
			return
		}
	}
	//I omitted validation for event ID because if no event exists then no ticket will be returned
	tickets := make([]models.Ticket, 0)
	var ticketPaginator *pagination.Paginator
	var paginationQuery *gorm.DB
	if name == "" {
		paginationQuery = models.DB().Preload("Images").Where(" is_available = ? AND owner_id = ? AND owner_type = ?", 1, params["event_id"], "events")
	} else {
		paginationQuery = models.DB().Preload("Images").Where("name LIKE ? AND is_available = ? AND owner_id = ? AND owner_type = ?", name+"%", 1, params["event_id"], "events")
	}

	ticketPaginator = pagination.Paging(&pagination.Param{
		DB:      paginationQuery,
		Page:    page,
		Limit:   limit,
		OrderBy: []string{"created_at desc"},
		ShowSQL: true,
	}, &tickets)

	responseData := make(map[string]interface{})
	responseData["tickets"] = ticketPaginator
	successResponse := response.Success{
		Status:  "success",
		Message: "Tickets retrieved successfully",
		Data:    responseData,
	}

	response.RespondSuccess(w, successResponse)
}

func Presenters(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var err error
	var page int
	var limit int
	name := r.URL.Query().Get("name")
	pageString := r.URL.Query().Get("page")
	if pageString == "" {
		page = 1
	} else {
		page, err = strconv.Atoi(pageString)
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
		limit, err = strconv.Atoi(limitString)
		if err != nil {
			utils.Log(err, "error")
			errorResponse := &response.Error{Status: "error", Message: "Oops! An error occurred. Please try again later."}
			response.RespondInternalServerError(w, errorResponse)
			return
		}
	}

	presenters := make([]models.Presenter, 0)
	var presenterPaginator *pagination.Paginator
	var paginationQuery *gorm.DB
	if name == "" {
		paginationQuery = models.DB().Preload("Images").Where(" is_available = ? AND event_id", 1, params["event_id"])
	} else {
		paginationQuery = models.DB().Preload("Images").Where("name LIKE ? AND is_available = ? AND event_id", name+"%", 1, params["event_id"])
	}

	presenterPaginator = pagination.Paging(&pagination.Param{
		DB:      paginationQuery,
		Page:    page,
		Limit:   limit,
		OrderBy: []string{"created_at desc"},
		ShowSQL: true,
	}, &presenters)

	responseData := make(map[string]interface{})
	responseData["presenters"] = presenterPaginator
	successResponse := response.Success{
		Status:  "success",
		Message: "Presenters retrieved successfully",
		Data:    responseData,
	}

	response.RespondSuccess(w, successResponse)
}
