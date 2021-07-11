package response

type Error struct {
	Status string `json:"status"`
	Message string `json:"message"`
	Errors interface{} `json:"errors"`
}