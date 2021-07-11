package validator

import (
	"errors"
	"fmt"
	"goventy/models"
	"goventy/utils"
	"io"
	"mime/multipart"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
	validator10 "github.com/go-playground/validator/v10"
	"github.com/jinzhu/gorm"
)

var validate *validator10.Validate

func init() {
	validate = validator10.New()
	_ = validate.RegisterValidation("passwd", func(fl validator10.FieldLevel) bool {
		return len(fl.Field().String()) > 6
	})

	_ = validate.RegisterValidation("fullname", func(fl validator10.FieldLevel) bool {
		validName := regexp.MustCompile("^[a-zA-Z-_\\s]*$")
		if !validName.MatchString(fl.Field().String()) {
			return false
		} else {
			return true
		}
	})

	_ = validate.RegisterValidation("username", func(fl validator10.FieldLevel) bool {
		validUsername := regexp.MustCompile("^[a-zA-Z0-9_\\s]*$")
		if !validUsername.MatchString(fl.Field().String()) {
			return false
		} else {
			return true
		}
	})
	/**
	* Validate that a value exists in table,column
	 */
	type Result struct {
		ID int
	}
	var result Result
	_ = validate.RegisterValidation("exists", func(fl validator10.FieldLevel) bool {
		params := strings.Split(fl.Param(), ".")
		table := params[0]
		column := params[1]
		query := fmt.Sprintf("SELECT id FROM %s WHERE %s = ?", table, column)
		err := models.DB().Raw(query, fl.Field().String()).Scan(&result).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false
		} else {
			if result.ID > 0 {
				return true
			}
			return false
		}
	})

	_ = validate.RegisterValidation("unique", func(fl validator10.FieldLevel) bool {
		params := strings.Split(fl.Param(), ".")
		table := params[0]
		column := params[1]
		query := fmt.Sprintf("SELECT id FROM %s WHERE %s = ?", table, column)
		err := models.DB().Raw(query, fl.Field().String()).Scan(&result).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true
		} else {
			if result.ID > 0 {
				return false
			}
			return true
		}
	}, false)

	_ = validate.RegisterValidation("timeInterval", func(fl validator10.FieldLevel) bool {
		validTimeInterval := regexp.MustCompile("^([0-9]|0[0-9]|1[0-9]|2[0-3]):[0-5][0-9] ([0-9]|0[0-9]|1[0-9]|2[0-3]):[0-5][0-9]$")
		if !validTimeInterval.MatchString(fl.Field().String()) {
			return false
		} else {
			return true
		}
	})

	//date comes after or equal to another date field in parent struct
	_ = validate.RegisterValidation("aftdtfield", func(fl validator10.FieldLevel) bool {
		if fl.Field().String() == "" {
			return true
		}
		dateToCompareTo, _ := time.Parse("2006-01-02", fl.Parent().Elem().FieldByName(fl.Param()).String())
		date, _ := time.Parse("2006-01-02", fl.Field().String())

		if date.After(dateToCompareTo) {
			return true
		}

		if date.Day() == dateToCompareTo.Day() && date.Month() == dateToCompareTo.Month() && date.Year() == dateToCompareTo.Year() {
			return true
		}

		return false
	}, false)

	_ = validate.RegisterValidation("date", func(fl validator10.FieldLevel) bool {
		if fl.Field().String() == "" {
			//we can do this as validation shouldn't occur if field is nil
			return true
		}
		//first validate that the format is correct YYYY-MM-DD
		validDate := regexp.MustCompile("((19|20)\\d\\d)-(0?[1-9]|1[012])-(0?[1-9]|[12][0-9]|3[01])")
		d30 := []int{4, 6, 9, 11}
		d31 := []int{1, 3, 5, 7, 8, 10, 12}
		if !validDate.MatchString(fl.Field().String()) {
			return false
		} else {
			dateParts := strings.Split(fl.Field().String(), "-")
			year, _ := strconv.Atoi(dateParts[0])
			month, _ := strconv.Atoi(dateParts[1])
			day, _ := strconv.Atoi(dateParts[2])

			monthMaxDay := 0
			for _, m := range d30 {
				if month == m {
					monthMaxDay = 30
				}
			}

			if monthMaxDay == 0 {
				for _, m := range d31 {
					if month == m {
						monthMaxDay = 31
					}
				}
			}

			if monthMaxDay == 0 {
				if month == 2 {
					//set default
					monthMaxDay = 28
					//detemine if this is a leapyear or not
					if year%4 == 0 {
						if year%100 == 0 {
							if year%400 == 0 {
								monthMaxDay = 29
							} else {
								monthMaxDay = 28
							}
						} else {
							monthMaxDay = 29
						}
					} else {
						monthMaxDay = 28
					}
				}
			}

			if monthMaxDay == 0 {
				//month was not found in any of the lists so it must be invalid
				return false
			}

			if day > monthMaxDay {
				return false
			}
			return true
		}
	}, false)
}

func ValidateFormUploadedImage(fileHeader *multipart.FileHeader) string {
	allowedMimeTypes := []string{"image/png", "image/jpeg", "image/jp2", "image/jpx", "image/jpm", "image/gif", "image/webp"}
	file, err := fileHeader.Open()
	defer file.Close()
	if err != nil {
		utils.Log(err, "error")
	}
	mime, err := mimetype.DetectReader(file)
	if err != nil {
		utils.Log(err, "error")
	}
	_, err = file.Seek(0, io.SeekStart)

	for _, value := range allowedMimeTypes {
		if mime.String() == value {
			return ""
		}
	}
	return "File is not a valid image. The file uploaded was of type " + mime.String()
}

func ValidateMultipleFormUploadedImage(images []*multipart.FileHeader) []string {
	for _, header := range images {
		errorMessage := ValidateFormUploadedImage(header)
		if errorMessage != "" {
			return []string{errorMessage}
		}
	}
	return nil
}

func ValidateFormUploadedFileMimeType(fileHeader *multipart.FileHeader, mimes []string) string {
	file, err := fileHeader.Open()
	defer file.Close()
	if err != nil {
		utils.Log(err, "error")
	}
	mime, err := mimetype.DetectReader(file)
	if err != nil {
		utils.Log(err, "error")
	}
	_, err = file.Seek(0, io.SeekStart)

	for _, value := range mimes {
		if mime.String() == value {
			return ""
		}
	}
	return "File is not valid. The file uploaded was of type " + mime.String()
}

func ValidateMultipleFormUploadedFileMimeType(files []*multipart.FileHeader, mimes []string) []string {
	for _, header := range files {
		errorMessage := ValidateFormUploadedFileMimeType(header, mimes)
		if errorMessage != "" {
			return []string{errorMessage}
		}
	}
	return nil
}

func Validate(fields interface{}) (error, map[string][]string) {
	err := validate.Struct(fields)
	if err != nil {

		if _, ok := err.(*validator10.InvalidValidationError); ok {
			return errors.New("Invalid Validation Syntax"), nil
		}

		validationErrors := make(map[string][]string)
		for _, err := range err.(validator10.ValidationErrors) {
			field, _ := reflect.TypeOf(fields).Elem().FieldByName(err.StructField())
			var name string
			//If json tag doesn't exist, use lower case of name
			if name = field.Tag.Get("json"); name == "" {
				name = strings.ToLower(err.StructField())
			}
			switch err.Tag() {
			case "required":
				validationErrors[name] = append(validationErrors[name], "The "+name+" is required")
				break
			case "email":
				validationErrors[name] = append(validationErrors[name], "The "+name+" should be a valid email")
				break
			case "eqfield":
				validationErrors[name] = append(validationErrors[name], "The "+name+" should be equal to the "+err.Param())
				break
			case "passwd":
				validationErrors[name] = append(validationErrors[name], "The "+name+" is not secure enough. Password must be at least 6 characters long")
				break
			case "fullname":
				validationErrors[name] = append(validationErrors[name], "Name must be letters, no special keys or numbers allowed")
				break
			case "username":
				validationErrors[name] = append(validationErrors[name], "Only alphabets, numbers, and underscore allowed for username")
				break
			case "unique":
				validationErrors[name] = append(validationErrors[name], "The "+name+" already exists")
				break
			case "exists":
				validationErrors[name] = append(validationErrors[name], "The "+name+" must exist. No record was found for the value supplied.")
				break
			case "timeInterval":
				validationErrors[name] = append(validationErrors[name], "The "+name+" time interval is invalid. Time interval must be in format [HH:MM HH:MM]")
				break
			case "date":
				validationErrors[name] = append(validationErrors[name], "The "+name+" is invalid. Date must be in format [YYYY-MM-DD]")
				break
			case "aftdtfield":
				relfield := err.Param()
				var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
				var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")
				snake := matchFirstCap.ReplaceAllString(relfield, "${1}_${2}")
				snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
				tfn := strings.ToLower(snake)
				validationErrors[name] = append(validationErrors[name], "The "+name+" is invalid. It must be a date after "+tfn)
				break
			default:
				validationErrors[name] = append(validationErrors[name], "The "+name+" is invalid")
				break
			}
		}
		return err, validationErrors
	}

	return nil, nil
}
