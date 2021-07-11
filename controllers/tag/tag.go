package tag

import (
	"goventy/models"
	"goventy/utils"
	"goventy/utils/response"
	"net/http"
	"strconv"

	"github.com/biezhi/gorm-paginator/pagination"
	"github.com/jinzhu/gorm"
)

func Find(w http.ResponseWriter, r *http.Request) {
	var err error
	page := 1
	limit := 20
	name := r.URL.Query().Get("name")
	pageString := r.URL.Query().Get("page")
	if pageString != "" {
		page, err = strconv.Atoi(pageString)
		if err != nil {
			utils.Log(err, "error")
			errorResponse := &response.Error{Status: "error", Message: "Oops! An error occurred. Please try again later."}
			response.RespondInternalServerError(w, errorResponse)
			return
		}
	}
	limitString := r.URL.Query().Get("limit")
	if limitString != "" {
		limit, err = strconv.Atoi(limitString)
		if err != nil {
			utils.Log(err, "error")
			errorResponse := &response.Error{Status: "error", Message: "Oops! An error occurred. Please try again later."}
			response.RespondInternalServerError(w, errorResponse)
			return
		}
	}
	//I omitted validation for event ID because if no event exists then no ticket will be returned
	tags := make([]models.Tag, 0)
	var tagPaginator *pagination.Paginator
	var paginationQuery *gorm.DB
	paginationQuery = models.DB().Where("name LIKE ?", name+"%")

	tagPaginator = pagination.Paging(&pagination.Param{
		DB:      paginationQuery,
		Page:    page,
		Limit:   limit,
		OrderBy: []string{"created_at desc"},
		ShowSQL: true,
	}, &tags)

	responseData := make(map[string]interface{})
	responseData["tags"] = tagPaginator
	successResponse := response.Success{
		Status:  "success",
		Message: "Tags retrieved successfully",
		Data:    responseData,
	}

	response.RespondSuccess(w, successResponse)
}
