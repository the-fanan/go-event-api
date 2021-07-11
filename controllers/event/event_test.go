package event

import (
	_ "fmt"
	"goventy/config"
	"goventy/models"
	"goventy/test"
	"goventy/utils/storage"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateEventWithInvalidValuesFails(t *testing.T) {
	req, err := test.CreateFormDataUploadRequest("/events", "POST", map[string][]string{
		"times": []string{"12:00 13:00"},
	},
		map[string][]string{},
		map[string]string{
			"Authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRfaWQiOjEsImV4cCI6MTYxMDA5NTA1M30.ukUXujoRWV2XBTIb_eOY6v-4VeUxB7VFHwx6cFp_32g",
		})
	if err != nil {
		t.Error(err)
	}
	res := httptest.NewRecorder()
	Create(res, req)
	if status := res.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	expected := "{\"status\":\"error\",\"message\":\"Invalid or missing input fields\",\"errors\":{\"description\":[\"The description is required\"],\"end_date\":[\"The end_date is required\"],\"name\":[\"The name is required\"],\"start_date\":[\"The start_date is required\"]}}"
	if res.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			res.Body.String(), expected)
	}
}

func TestCreateEventWithValidValuesWorks(t *testing.T) {
	params := map[string][]string{
		"name":        []string{"Test Event"},
		"description": []string{"Description of a test event"},
		"address":     []string{"localhost"},
		"start_date":  []string{"2020-06-26"},
		"end_date":    []string{"2020-06-27"},
		"times":       []string{"12:00 13:00"},
	}
	req, err := test.CreateFormDataUploadRequest("/events", "POST", params,
		map[string][]string{
			"images": []string{
				filepath.Join(os.Getenv("goventy_ROOT"), "test/mocks/images/t1.jpg"),
				filepath.Join(os.Getenv("goventy_ROOT"), "test/mocks/images/t1.jpg"),
			},
		},
		map[string]string{
			"Authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRfaWQiOjEsImV4cCI6MTYxMDA5NTA1M30.ukUXujoRWV2XBTIb_eOY6v-4VeUxB7VFHwx6cFp_32g",
		})
	if err != nil {
		t.Error(err)
	}
	res := httptest.NewRecorder()
	Create(res, req)
	if status := res.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	//validate event and image exist in database
	event := &models.Event{}
	images := make([]models.Image, 0)

	err = models.DB().Where(" is_available = ? AND name = ? AND description = ?", 0, params["name"][0], params["description"][0]).First(event).Error
	if err != nil {
		t.Error(err)
	}
	if event.ID <= 0 {
		t.Error("Event was not created")
	}
	err = models.DB().Where("owner_type = ? AND owner_id = ?", "events", event.ID).Find(&images).Error
	if err != nil {
		t.Error(err)
	}
	//validate that the images created exist
	for _, image := range images {
		_, err := os.Stat(filepath.Join(os.Getenv("goventy_ROOT"), filepath.Join(config.ENV()["PUBLIC_ASSETS"], image.Url)))
		if err != nil {
			t.Error(err)
		}
		//delete the files
		ls := &storage.LocalStorage{
			FileName: image.Url,
		}
		ls.Delete()
	}

	models.DB().Exec("SET FOREIGN_KEY_CHECKS = 0")
	models.DB().Exec("TRUNCATE table goventy.events")
	models.DB().Exec("TRUNCATE table goventy.images")
	models.DB().Exec("SET FOREIGN_KEY_CHECKS = 1")

	/*expected := "{\"status\":\"success\",\"message\":\"Event created successfully\",\"data\":{\"event\":{\"id\":1,\"created_at\":\"2020-09-30T11:13:49.451661004+01:00\",\"updated_at\":\"2020-09-30T11:13:49.451661004+01:00\",\"deleted_at\":null,\"user_id\":1,\"name\":\"Test Event\",\"description\":\"Description of a test event\",\"address\":\"localhost\",\"start_date\":\"2020-06-26\",\"end_date\":\"2020-06-27\",\"times\":\"[\\\"12:00 13:00\\\"]\",\"is_available\":0,\"tickets\":null,\"presenters\":null,\"abuse_reports\":null,\"images\":[{\"id\":1,\"created_at\":\"2020-09-30T11:13:49.454990499+01:00\",\"updated_at\":\"2020-09-30T11:13:49.454990499+01:00\",\"deleted_at\":null,\"user_id\":1,\"url\":\"uploads/events/images/1iE7K07eeZzlsSCmq7kGcr0gG5u.jpg\",\"owner_type\":\"events\",\"owner_id\":1,\"provider\":\"local\",\"provider_id\":\"1iE7K07eeZzlsSCmq7kGcr0gG5u\",\"priority\":0,\"dimension\":\"2D\"},{\"id\":2,\"created_at\":\"2020-09-30T11:13:49.458139512+01:00\",\"updated_at\":\"2020-09-30T11:13:49.458139512+01:00\",\"deleted_at\":null,\"user_id\":1,\"url\":\"uploads/events/images/1iE7JyB1eemhmxckHVAaZnAZPgL.jpg\",\"owner_type\":\"events\",\"owner_id\":1,\"provider\":\"local\",\"provider_id\":\"1iE7JyB1eemhmxckHVAaZnAZPgL\",\"priority\":0,\"dimension\":\"2D\"}],\"ratings\":null,\"reviews\":null,\"tags\":null}}}"
		if res.Body.String() != expected {
	        t.Errorf("handler returned unexpected body: got %v want %v",
			res.Body.String(), expected)
		}*/
}
