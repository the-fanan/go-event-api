package presenter

import (
	"encoding/json"
	"goventy/models"
	"goventy/utils"
	"goventy/utils/response"
	"goventy/utils/storage"
	"goventy/utils/validator"
	"net/http"
	"strconv"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
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

	type PresenterValidationStruct struct {
		EventId string `json:"event_id" validate:"required,exists=events.id"`
		Name    string `validate:"required"`
		Topic   string `validate:"required"`
	}

	pvs := &PresenterValidationStruct{
		EventId: r.FormValue("event_id"),
		Name:    r.FormValue("name"),
		Topic:   r.FormValue("description"),
	}

	_, vErrors := validator.Validate(pvs)

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

	eventId64, _ := strconv.ParseUint(r.FormValue("event_id"), 10, 64)
	eventId := uint(eventId64)

	event := &models.Event{}

	err = models.DB().First(event, eventId).Error
	if err != nil {
		utils.Log(err.Error(), "error")
		errorResponse := &response.Error{Status: "error", Message: "Oops! An error occurred. Please try again later."}
		response.RespondInternalServerError(w, errorResponse)
		return
	}

	if event.UserID != user_id {
		errorResponse := &response.Error{Status: "error", Message: "You do not have permission to add a presenter to this event"}
		response.RespondBadRequest(w, errorResponse)
		return
	}

	presenter := &models.Presenter{
		UserID:  user_id,
		EventID: eventId,
		Name:    r.FormValue("name"),
		Topic:   r.FormValue("description"),
	}
	presenterImages := formdata.File["images"]
	presenterImagesStructs := make([]models.Image, len(presenterImages))
	for i, _ := range presenterImages {
		file, err := presenterImages[i].Open()
		defer file.Close()
		if err != nil {
			utils.Log(err, "error")
			return
		}
		localStore := &storage.LocalStorage{
			Folder: "uploads/presenters/images",
		}
		uploadedFile := localStore.Create(file)
		image := &models.Image{
			UserID:     user_id,
			Url:        uploadedFile.Url,
			Provider:   uploadedFile.Provider,
			ProviderID: uploadedFile.ProviderID,
		}
		presenterImagesStructs[i] = *image
	}
	//ad images to ticket
	presenter.Images = presenterImagesStructs

	err = models.DB().Create(presenter).Error
	if err != nil {
		utils.Log(err.Error(), "error")
		errorResponse := &response.Error{Status: "error", Message: "Oops! An error occurred. Please try again later."}
		response.RespondInternalServerError(w, errorResponse)
		return
	}

	//send success response
	responseData := make(map[string]interface{})
	responseData["presenter"] = presenter
	successResponse := response.Success{
		Status:  "success",
		Message: "Presenter created successfully",
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
		Name        string `json:"name"`
		Topic       string `json:"topic"`
		IsAvailable int    `json:"is_available"`
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

	type PresenterValidationStruct struct {
		PresenterId string `json:"presenter_id" validate:"required,exists=presenters.id"`
		IsAvailable int    `json:"is_available" validate:"numeric"`
	}
	tivs := &PresenterValidationStruct{
		PresenterId: params["id"],
		IsAvailable: request.IsAvailable,
	}
	_, vErrors := validator.Validate(tivs)
	if len(vErrors) != 0 {
		errorResponse := &response.Error{Status: "error", Message: "Invalid or missing input fields", Errors: vErrors}
		response.RespondBadRequest(w, errorResponse)
		return
	}

	id, _ := strconv.ParseUint(params["id"], 10, 64)

	presenter := &models.Presenter{}
	err = models.DB().First(presenter, id).Error
	if err != nil {
		utils.Log(err.Error(), "error")
		errorResponse := &response.Error{Status: "error", Message: "Oops! An error occurred. Please try again later."}
		response.RespondInternalServerError(w, errorResponse)
		return
	}

	if presenter.UserID != user_id {
		errorResponse := &response.Error{Status: "error", Message: "You do not have permission to carry out this action"}
		response.RespondUnauthorized(w, errorResponse)
		return
	}

	err = models.DB().Model(presenter).Updates(request).Error
	if err != nil {
		utils.Log(err.Error(), "error")
		errorResponse := &response.Error{Status: "error", Message: "Oops! An error occurred. Please try again later."}
		response.RespondInternalServerError(w, errorResponse)
		return
	}
	//handle zero values not handled by Updates
	err = models.DB().Model(presenter).Update("is_available", request.IsAvailable).Error
	if err != nil {
		utils.Log(err.Error(), "error")
		errorResponse := &response.Error{Status: "error", Message: "Oops! An error occurred. Please try again later."}
		response.RespondInternalServerError(w, errorResponse)
		return
	}

	responseData := make(map[string]interface{})
	responseData["presenter"] = presenter
	successResponse := response.Success{
		Status:  "success",
		Message: "Presenter updated successfully",
		Data:    responseData,
	}

	response.RespondSuccess(w, successResponse)
}

func Get(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	type PresenterValidationStruct struct {
		PresenterId string `json:"presenter_id" validate:"required,exists=presenters.id"`
	}
	tivs := &PresenterValidationStruct{
		PresenterId: params["id"],
	}
	_, vErrors := validator.Validate(tivs)
	if len(vErrors) != 0 {
		errorResponse := &response.Error{Status: "error", Message: "Invalid or missing input fields", Errors: vErrors}
		response.RespondBadRequest(w, errorResponse)
		return
	}

	id, _ := strconv.ParseUint(params["id"], 10, 64)

	presenter := &models.Presenter{}
	err := models.DB().First(presenter, id).Error
	if err != nil {
		utils.Log(err.Error(), "error")
		errorResponse := &response.Error{Status: "error", Message: "Oops! An error occurred. Please try again later."}
		response.RespondInternalServerError(w, errorResponse)
		return
	}

	//get images
	images := make([]models.Image, 0)
	models.DB().Where("owner_type = ? AND owner_id = ?", "presenters", presenter.ID).Find(&images)
	presenter.Images = images

	responseData := make(map[string]interface{})
	responseData["presenter"] = presenter
	successResponse := response.Success{
		Status:  "success",
		Message: "Presenter retrieved successfully",
		Data:    responseData,
	}

	response.RespondSuccess(w, successResponse)
}
