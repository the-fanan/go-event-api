package seeds

import (
	"goventy/models"
)

func Seeder_2021_01_18_1_tags() {
	tags := [][]string{
		{"cruise", "Spans multiple days and many places"},
		{"chop-life", "Eat plenty till you drop"},
		{"arts", "Plenty of art galleries to visit"},
		{"date", "For your and your pertner"},
		{"tech", "Geeks and techies"},
		{"software", "build tech skill"},
	}

	for i := 0; i < len(tags); i++ {
		tag := models.Tag{}
		tag.Name = tags[i][0]
		tag.Description = tags[i][1]
		models.DB().Create(&tag)
	}
}
