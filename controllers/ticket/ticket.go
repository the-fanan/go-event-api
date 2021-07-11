package ticket

import (
	"encoding/json"
	"errors"
	"fmt"
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

	type TicketValidationStruct struct {
		OwnerId           string `json:"owner_id" validate:"required,numeric"`
		OwnerType         string `json:"owner_type" validate:"required"`
		Name              string `validate:"required"`
		Description       string `validate:"required"`
		Amount            string `json:"amount" validate:"required,numeric,gte=0"`
		QuantityAvailable string `json:"quantity_available" validate:"required,numeric,gte=1"`
	}
	tivs := &TicketValidationStruct{
		OwnerId:           r.FormValue("owner_id"),
		OwnerType:         r.FormValue("owner_type"),
		Name:              r.FormValue("name"),
		Description:       r.FormValue("description"),
		Amount:            r.FormValue("amount"),
		QuantityAvailable: r.FormValue("quantity_available"),
	}
	_, vErrors := validator.Validate(tivs)

	imageErrors := validator.ValidateMultipleFormUploadedImage(formdata.File["images"])
	if imageErrors != nil {
		if vErrors == nil {
			vErrors = make(map[string][]string)
			vErrors["images"] = imageErrors
		} else {
			vErrors["images"] = imageErrors
		}
	}

	ownerId64, _ := strconv.ParseUint(r.FormValue("owner_id"), 10, 64)
	ownerId := uint(ownerId64)
	type Result struct {
		ID     int
		UserID int
	}
	var owner Result
	query := fmt.Sprintf("SELECT id FROM %s WHERE id = ?", r.FormValue("owner_type"))
	err = models.DB().Raw(query, r.FormValue("owner_id")).Scan(&owner).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		vErrors["owner_id"] = append(vErrors["owner_id"], "The record does not exist")
	} else {
		if owner.ID <= 0 {
			vErrors["owner_id"] = append(vErrors["owner_id"], "The record does not exist")
		}
	}

	if len(vErrors) != 0 {
		errorResponse := &response.Error{Status: "error", Message: "Invalid or missing input fields", Errors: vErrors}
		response.RespondBadRequest(w, errorResponse)
		return
	}

	if owner.UserID != int(user_id) {
		errorResponse := &response.Error{Status: "error", Message: "You do not have permission to add a ticket to this event"}
		response.RespondBadRequest(w, errorResponse)
		return
	}

	amount := 0.0
	quantity := 0

	if r.FormValue("quantity_available") != "" {
		quantity, _ = strconv.Atoi(r.FormValue("quantity_available"))
		if quantity < 1 {
			quantity = 1
		}
	}

	if r.FormValue("amount") != "" {
		amount, _ = strconv.ParseFloat(r.FormValue("amount"), 64)
		if amount < 0 {
			amount = 0
		}
	}

	ticket := &models.Ticket{
		UserID:            user_id,
		OwnerType:         r.FormValue("owner_type"),
		OwnerID:           ownerId,
		Name:              r.FormValue("name"),
		Description:       r.FormValue("description"),
		Amount:            amount,
		QuantityAvailable: quantity,
	}

	//get images for event if any
	ticketImages := formdata.File["images"]
	ticketImagesStructs := make([]models.Image, len(ticketImages))
	for i, _ := range ticketImages {
		file, err := ticketImages[i].Open()
		defer file.Close()
		if err != nil {
			utils.Log(err, "error")
			return
		}
		localStore := &storage.LocalStorage{
			Folder: "uploads/tickets/images",
		}
		uploadedFile := localStore.Create(file)
		image := &models.Image{
			UserID:     user_id,
			Url:        uploadedFile.Url,
			Provider:   uploadedFile.Provider,
			ProviderID: uploadedFile.ProviderID,
		}
		ticketImagesStructs[i] = *image
	}
	ticket.Images = ticketImagesStructs

	err = models.DB().Create(ticket).Error
	if err != nil {
		utils.Log(err.Error(), "error")
		errorResponse := &response.Error{Status: "error", Message: "Oops! An error occurred. Please try again later."}
		response.RespondInternalServerError(w, errorResponse)
		return
	}

	//send success response
	responseData := make(map[string]interface{})
	responseData["ticket"] = ticket
	successResponse := response.Success{
		Status:  "success",
		Message: "Ticket created successfully",
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
		Name              string `json:"name"`
		Description       string `json:"description"`
		Amount            string `json:"amount" validate:"required,numeric,gte=0"`
		QuantityAvailable string `json:"quantity_available" validate:"required,numeric,gte=1"`
		IsAvailable       int    `json:"is_available"`
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

	type TicketValidationStruct struct {
		TicketId          string `json:"ticket_id" validate:"required,exists=tickets.id"`
		Amount            string `json:"amount" validate:"numeric,gte=0"`
		QuantityAvailable string `json:"quantity_available" validate:"numeric,gte=1"`
		IsAvailable       int    `json:"is_available" validate:"numeric"`
	}
	tivs := &TicketValidationStruct{
		TicketId:          params["id"],
		Amount:            request.Amount,
		QuantityAvailable: request.QuantityAvailable,
		IsAvailable:       request.IsAvailable,
	}
	_, vErrors := validator.Validate(tivs)
	if len(vErrors) != 0 {
		errorResponse := &response.Error{Status: "error", Message: "Invalid or missing input fields", Errors: vErrors}
		response.RespondBadRequest(w, errorResponse)
		return
	}

	id, _ := strconv.ParseUint(params["id"], 10, 64)

	ticket := &models.Ticket{}
	err = models.DB().First(ticket, id).Error
	if err != nil {
		utils.Log(err.Error(), "error")
		errorResponse := &response.Error{Status: "error", Message: "Oops! An error occurred. Please try again later."}
		response.RespondInternalServerError(w, errorResponse)
		return
	}

	//validate that user has permissions to edit this event
	if ticket.UserID != user_id {
		errorResponse := &response.Error{Status: "error", Message: "You do not have permission to carry out this action"}
		response.RespondUnauthorized(w, errorResponse)
		return
	}

	err = models.DB().Model(ticket).Updates(request).Error
	if err != nil {
		utils.Log(err.Error(), "error")
		errorResponse := &response.Error{Status: "error", Message: "Oops! An error occurred. Please try again later."}
		response.RespondInternalServerError(w, errorResponse)
		return
	}
	//handle zero values not handled by Updates
	err = models.DB().Model(ticket).Update("is_available", request.IsAvailable).Error
	if err != nil {
		utils.Log(err.Error(), "error")
		errorResponse := &response.Error{Status: "error", Message: "Oops! An error occurred. Please try again later."}
		response.RespondInternalServerError(w, errorResponse)
		return
	}

	responseData := make(map[string]interface{})
	responseData["ticket"] = ticket
	successResponse := response.Success{
		Status:  "success",
		Message: "Ticket updated successfully",
		Data:    responseData,
	}

	response.RespondSuccess(w, successResponse)
}

func Get(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	type TicketValidationStruct struct {
		TicketId string `json:"ticket_id" validate:"required,exists=tickets.id"`
	}
	tivs := &TicketValidationStruct{
		TicketId: params["id"],
	}
	_, vErrors := validator.Validate(tivs)
	if len(vErrors) != 0 {
		errorResponse := &response.Error{Status: "error", Message: "Invalid or missing input fields", Errors: vErrors}
		response.RespondBadRequest(w, errorResponse)
		return
	}

	id, _ := strconv.ParseUint(params["id"], 10, 64)

	ticket := &models.Ticket{}
	err := models.DB().First(ticket, id).Error
	if err != nil {
		utils.Log(err.Error(), "error")
		errorResponse := &response.Error{Status: "error", Message: "Oops! An error occurred. Please try again later."}
		response.RespondInternalServerError(w, errorResponse)
		return
	}

	//get images
	images := make([]models.Image, 0)
	models.DB().Where("owner_type = ? AND owner_id = ?", "tickets", ticket.ID).Find(&images)
	ticket.Images = images

	responseData := make(map[string]interface{})
	responseData["ticket"] = ticket
	successResponse := response.Success{
		Status:  "success",
		Message: "Ticket retrieved successfully",
		Data:    responseData,
	}

	response.RespondSuccess(w, successResponse)
}
