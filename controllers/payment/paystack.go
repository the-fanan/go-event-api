package payment

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"goventy/config"
	"goventy/models"
	"goventy/utils"
	"goventy/utils/emails"
	"goventy/utils/response"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/segmentio/ksuid"
)

func ProcessPaystackWebhook(w http.ResponseWriter, r *http.Request) {
	//get IP address of requester
	ip := r.Header.Get("X-FORWARDED-FOR")
	if ip == "" {
		ip = r.RemoteAddr
	}
	//get paystack x-signature
	paystackXSignature := r.Header.Get("x-paystack-signature")
	if paystackXSignature == "" {
		utils.Log(fmt.Sprintf("No Paystack X signature sent from IP [%s] that tried to send a payment request", ip), "warn")
		successResponse := response.Success{
			Status:  "success",
			Message: "Information has been queued for processing",
		}
		response.RespondSuccess(w, successResponse)
		return
	}
	//verify that the IP sending this request is from Paystack
	allowedIPs := strings.Split(config.ENV()["PAYSTACK_ALLOWED_IPS"], ",")
	for i, allowedIP := range allowedIPs {
		if ip == allowedIP {
			break
		}
		//if this is the last item then the IP is not allowed. Return success but do not process anything
		if i == len(allowedIPs)-1 {
			utils.Log(fmt.Sprintf("An invalid IP [%s] tried to send a payment request", ip), "warn")
			successResponse := response.Success{
				Status:  "success",
				Message: "Information has been queued for processing",
			}
			response.RespondSuccess(w, successResponse)
			return
		}
	}
	//get Body of request. It is a JSON string
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.Log(err, "error")
		errorResponse := &response.Error{Status: "error", Message: "Oops! An error occurred. Please try again later."}
		response.RespondInternalServerError(w, errorResponse)
		return
	}
	bodyString := string(bodyBytes)
	mac := hmac.New(sha512.New, []byte(config.ENV()["PAYSTACK_SECRET_KEY"]))
	mac.Write([]byte(bodyString))
	sha := hex.EncodeToString(mac.Sum(nil))

	if sha != paystackXSignature {
		utils.Log(fmt.Sprintf("An invalid x-paystack-signature was sent by IP [%s] that tried to send a payment request", ip), "warn")
		successResponse := response.Success{
			Status:  "success",
			Message: "Information has been queued for processing",
		}
		response.RespondSuccess(w, successResponse)
		return
	}

	//the request is from a valid source, process it
	var payload map[string]interface{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&payload)
	if err != nil {
		vErrors := make(map[string][]string)
		vErrors["json_decode"] = []string{err.Error()}
		errorResponse := &response.Error{Status: "error", Message: "Invalid data types supplied", Errors: vErrors}
		response.RespondBadRequest(w, errorResponse)
		return
	}

	switch payload["event"].(string) {
	case "charge.success":
		go paystackProcessChargeEvent(payload["data"].(map[string]interface{}))
		break
	}

	successResponse := response.Success{
		Status:  "success",
		Message: "Information has been queued for processing",
	}
	response.RespondSuccess(w, successResponse)
	return
}

func paystackProcessChargeEvent(data map[string]interface{}) {
	if data["status"].(string) != "success" {
		//no need to process the transaction as it failed
		return
	}

	//was this payment for a ticket or a quick item?
	switch data["metadata"].(map[string]interface{})["owner_type"].(string) {
	case "tickets":
		//process ticket payment
		paystackProcessTicketPayment(data)
		break

	case "quick_items":
		break
	}
}

func paystackProcessTicketPayment(data map[string]interface{}) {
	//get ticket from owner_type and owner_id
	ticket_id := data["metadata"].(map[string]interface{})["owner_id"].(string)
	user_id := data["metadata"].(map[string]interface{})["user_id"].(string)
	user_email := data["metadata"].(map[string]interface{})["user_email"].(string)
	user_name := data["metadata"].(map[string]interface{})["user_name"].(string)
	amount := data["amount"].(float64)
	ticket := &models.Ticket{}
	models.DB().First(ticket, ticket_id)
	//validate that paid amount is equal to ticket amount
	if amount < ticket.Amount {
		return
	}
	//generate code for sale
	code_id := ksuid.New()
	year, month, day := time.Now().Date()
	//FLJ - goventy
	//TK - Ticket
	code := fmt.Sprintf("FLJ-TK-%s-%d%d%d", code_id.String(), year, month, day)
	//create sale
	ticket_id_uint64, _ := strconv.ParseUint(ticket_id, 10, 64)
	sale := &models.Sale{
		OwnerType: "tickets",
		OwnerID:   uint(ticket_id_uint64),
		Code:      code,
		UserEmail: user_email,
		UserName:  user_name,
	}

	if user_id != "" {
		user_id_uint64, _ := strconv.ParseUint(user_id, 10, 64)
		sale.UserID = uint(user_id_uint64)
	}

	err := models.DB().Create(sale).Error
	if err != nil {
		utils.Log(err.Error(), "error")
		return
	}
	//create payment
	payment := &models.Payment{
		OwnerType:         "sales",
		OwnerID:           sale.ID,
		UserEmail:         user_email,
		UserName:          user_name,
		Amount:            amount,
		Details:           "Payment for Ticket",
		Provider:          "paystack",
		ProviderReference: data["reference"].(string),
	}

	if user_id != "" {
		user_id_uint64, _ := strconv.ParseUint(user_id, 10, 64)
		payment.UserID = uint(user_id_uint64)
	}

	err = models.DB().Create(payment).Error
	if err != nil {
		utils.Log(err.Error(), "error")
		return
	}
	//send email to user with details of ticket
	user_send_email := user_email
	user_send_name := user_name
	if user_id != "" {
		user := &models.User{}
		err := models.DB().First(user, user_id).Error
		if err == nil {
			user_send_email = user.Email
			user_send_name = user.Name
		}
	}
	type TemplateData struct {
		Name     string
		Code     string
		Amount   float64
		FrontUrl string
		Year     int
		TicketID string
	}
	td := TemplateData{
		Name:     strings.Title(strings.ToLower(user_send_name)),
		Code:     code,
		FrontUrl: config.ENV()["FRONTEND_URL"],
		Year:     time.Now().UTC().Year(),
		Amount:   amount,
		TicketID: ticket_id,
	}
	mailer := &emails.Mailer{
		TemplateData: td,
		TemplatePath: "/user/ticket.html",
		To:           []string{user_send_email},
		Subject:      "Welcome To goventy",
	}
	go mailer.Send()
}
