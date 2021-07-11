package image

import (
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

	_ "github.com/biezhi/gorm-paginator/pagination"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

func Update(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(30 * 1024 * 1024) // grab the multipart form
	if err != nil {
		errorResponse := &response.Error{Status: "error", Message: "Request body is too large"}
		response.RespondRequestTooLarge(w, errorResponse)
		return
	}

	token, err := utils.GetJwtToken(strings.Split(r.Header.Get("Authorization"), " ")[1])
	if err != nil {
		utils.Log(err, "error")
	}

	claims, _ := token.Claims.(jwt.MapClaims)
	user_id := uint(claims["client_id"].(float64))

	params := mux.Vars(r)

	formdata := r.MultipartForm

	type ImageValidationStruct struct {
		ID       string `json:"event_id" validate:"required,exists=images.id"`
		Priority string `validate:"numberic,gte=0"`
		Owner    string `validate:"required"`
	}

	ivs := &ImageValidationStruct{
		ID:       params["id"],
		Priority: r.FormValue("priority"),
		Owner:    r.FormValue("owner_type"),
	}

	_, vErrors := validator.Validate(ivs)
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

	imageId64, _ := strconv.ParseUint(params["id"], 10, 64)
	imageId := uint(imageId64)

	image := &models.Image{}

	err = models.DB().First(image, imageId).Error
	if err != nil {
		utils.Log(err.Error(), "error")
		errorResponse := &response.Error{Status: "error", Message: "Oops! An error occurred. Please try again later."}
		response.RespondInternalServerError(w, errorResponse)
		return
	}

	if image.UserID != user_id {
		errorResponse := &response.Error{Status: "error", Message: "You do not have permission to carry out this action."}
		response.RespondUnauthorized(w, errorResponse)
		return
	}

	type UpdateParams struct {
		Priority string
		Url      string
	}

	updateParams := &UpdateParams{
		Priority: r.FormValue("priority"),
	}

	var url string
	for i, _ := range formdata.File["images"] {
		if i > 0 {
			//we only take the first image
			break
		}
		file, err := formdata.File["images"][i].Open()
		defer file.Close()
		if err != nil {
			utils.Log(err, "error")
			return
		}
		localStore := &storage.LocalStorage{
			Folder: fmt.Sprintf("uploads/%s/images", r.FormValue("owner_type")),
		}
		uploadedFile := localStore.Create(file)
		url = uploadedFile.Url
	}
	if url != "" {
		updateParams.Url = url
		//delete old file
		ls := &storage.LocalStorage{
			FileName: image.Url,
		}
		ls.Delete()
	}

	err = models.DB().Model(image).Updates(updateParams).Error
	if err != nil {
		utils.Log(err.Error(), "error")
		errorResponse := &response.Error{Status: "error", Message: "Oops! An error occurred. Please try again later."}
		response.RespondInternalServerError(w, errorResponse)
		return
	}

	responseData := make(map[string]interface{})
	responseData["image"] = image
	successResponse := response.Success{
		Status:  "success",
		Message: "Image updated successfully",
		Data:    responseData,
	}

	response.RespondSuccess(w, successResponse)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	token, err := utils.GetJwtToken(strings.Split(r.Header.Get("Authorization"), " ")[1])
	if err != nil {
		utils.Log(err, "error")
		errorResponse := &response.Error{Status: "error", Message: "Oops! An error occurred. Please try again later."}
		response.RespondInternalServerError(w, errorResponse)
		return
	}

	claims, _ := token.Claims.(jwt.MapClaims)
	user_id := uint(claims["client_id"].(float64))

	params := mux.Vars(r)

	type ImageValidationStruct struct {
		ID string `json:"event_id" validate:"required,exists=images.id"`
	}

	ivs := &ImageValidationStruct{
		ID: params["id"],
	}

	_, vErrors := validator.Validate(ivs)

	if len(vErrors) != 0 {
		errorResponse := &response.Error{Status: "error", Message: "Invalid or missing input fields", Errors: vErrors}
		response.RespondBadRequest(w, errorResponse)
		return
	}

	imageId64, _ := strconv.ParseUint(params["id"], 10, 64)
	imageId := uint(imageId64)

	image := &models.Image{}

	err = models.DB().First(image, imageId).Error
	if err != nil {
		utils.Log(err.Error(), "error")
		errorResponse := &response.Error{Status: "error", Message: "Oops! An error occurred. Please try again later."}
		response.RespondInternalServerError(w, errorResponse)
		return
	}

	if image.UserID != user_id {
		errorResponse := &response.Error{Status: "error", Message: "You do not have permission to carry out this action."}
		response.RespondUnauthorized(w, errorResponse)
		return
	}

	ls := &storage.LocalStorage{
		FileName: image.Url,
	}
	ls.Delete()

	err = models.DB().Delete(image).Error
	if err != nil {
		utils.Log(err.Error(), "error")
		errorResponse := &response.Error{Status: "error", Message: "Oops! An error occurred. Please try again later."}
		response.RespondInternalServerError(w, errorResponse)
		return
	}

	responseData := make(map[string]interface{})
	responseData["image"] = image
	successResponse := response.Success{
		Status:  "success",
		Message: "Image deleted successfully",
		Data:    responseData,
	}

	response.RespondSuccess(w, successResponse)
}

func Append(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(30 * 1024 * 1024) // grab the multipart form
	if err != nil {
		errorResponse := &response.Error{Status: "error", Message: "Request body is too large"}
		response.RespondRequestTooLarge(w, errorResponse)
		return
	}

	token, err := utils.GetJwtToken(strings.Split(r.Header.Get("Authorization"), " ")[1])
	if err != nil {
		utils.Log(err, "error")
	}

	claims, _ := token.Claims.(jwt.MapClaims)
	user_id := uint(claims["client_id"].(float64))

	formdata := r.MultipartForm

	type ImageValidationStruct struct {
		Owner   string `validate:"required"`
		OwnerID string `validate:"required,numeric"`
	}

	ivs := &ImageValidationStruct{
		Owner:   r.FormValue("owner_type"),
		OwnerID: r.FormValue("owner_id"),
	}

	_, vErrors := validator.Validate(ivs)
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

	//validate that the parent supplied is a valid one
	type ParentResult struct {
		ID     int
		UserID int
	}
	var parent ParentResult

	query := fmt.Sprintf("SELECT id, user_id FROM %s WHERE id = ?", r.FormValue("owner_type"))
	err = models.DB().Raw(query, r.FormValue("owner_id")).Scan(&parent).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		errorResponse := &response.Error{Status: "error", Message: fmt.Sprintf("The parent entity  [%s] does not exist", r.FormValue("owner_type"))}
		response.RespondBadRequest(w, errorResponse)
		return
	}

	//validate that the user has permission to modify the parent
	if user_id != uint(parent.UserID) {
		errorResponse := &response.Error{Status: "error", Message: "You do not have permission to carry out this action."}
		response.RespondUnauthorized(w, errorResponse)
		return
	}

	//append images
	imagesStructs := make([]models.Image, len(formdata.File["images"]))
	for i, _ := range formdata.File["images"] {
		file, err := formdata.File["images"][i].Open()
		defer file.Close()
		if err != nil {
			utils.Log(err, "error")
			return
		}
		localStore := &storage.LocalStorage{
			Folder: fmt.Sprintf("uploads/%s/images", r.FormValue("owner_type")),
		}
		uploadedFile := localStore.Create(file)
		image := &models.Image{
			OwnerType:  r.FormValue("owner_type"),
			OwnerID:    uint(parent.ID),
			UserID:     user_id,
			Url:        uploadedFile.Url,
			Provider:   uploadedFile.Provider,
			ProviderID: uploadedFile.ProviderID,
		}
		models.DB().Create(image)
		imagesStructs[i] = *image
	}

	responseData := make(map[string]interface{})
	responseData["images"] = imagesStructs
	successResponse := response.Success{
		Status:  "success",
		Message: "Images appended successfully",
		Data:    responseData,
	}

	response.RespondSuccess(w, successResponse)
}
