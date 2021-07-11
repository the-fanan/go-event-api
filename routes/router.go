package routes

import (
	"goventy/config"
	"goventy/controllers"
	AuthController "goventy/controllers/auth"
	EventController "goventy/controllers/event"
	ImageController "goventy/controllers/image"
	PaymentController "goventy/controllers/payment"
	PresenterController "goventy/controllers/presenter"
	TagController "goventy/controllers/tag"
	TicketController "goventy/controllers/ticket"
	UserController "goventy/controllers/user"
	"goventy/utils/middleware"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
)

var router *mux.Router

/**
* This is where route paths are defined
 */
func init() {
	router = mux.NewRouter().StrictSlash(true)
	/**
	* Static route pah for assets (CSS, JS, Images and fonts)
	 */
	router.PathPrefix("/assets/").Handler(
		http.StripPrefix(
			"/assets/",
			http.FileServer(
				http.Dir(
					filepath.Join(os.Getenv("goventy_ROOT"), config.ENV()["PUBLIC_ASSETS"])))))

	/**
	* Index
	 */
	router.HandleFunc("/", controllers.Home)

	/**
	* Authentication
	 */
	auth := router.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/register", AuthController.Register).Methods("POST")
	auth.HandleFunc("/login", AuthController.Login).Methods("POST")

	/**
	* The following routes require a valid JWT token to get access to them
	 */
	gJam := &middleware.JwtAuthMiddleware{}

	/**
	* Event routes
	 */
	eventGaurded := router.PathPrefix("/events").Subrouter()
	eventUngaurded := router.PathPrefix("/events").Subrouter()
	//event routes that require authentication
	eventGaurded.Use(gJam.GeneralAuthentication)
	eventGaurded.HandleFunc("", EventController.Create).Methods("POST")
	eventGaurded.HandleFunc("/{id}", EventController.Update).Methods("PUT")
	//event routes that do not require authentication
	eventUngaurded.HandleFunc("", EventController.Find).Methods("GET")
	eventUngaurded.HandleFunc("/{id}", EventController.Get).Methods("GET")
	eventUngaurded.HandleFunc("/{event_id}/tickets", EventController.Tickets).Methods("GET")
	eventUngaurded.HandleFunc("/{event_id}/presenters", EventController.Presenters).Methods("GET")

	/**
	* Ticket routes
	 */
	ticketGaurded := router.PathPrefix("/tickets").Subrouter()
	ticketUngaurded := router.PathPrefix("/tickets").Subrouter()
	//ticket routes that require authentication
	ticketGaurded.Use(gJam.GeneralAuthentication)
	ticketGaurded.HandleFunc("", TicketController.Create).Methods("POST")
	ticketGaurded.HandleFunc("/{id}", TicketController.Update).Methods("PUT")
	//ticket routes that do not require authentication
	ticketUngaurded.HandleFunc("/{id}", TicketController.Get).Methods("GET")
	/**
	* Presenter routes
	 */
	presenterGaurded := router.PathPrefix("/presenters").Subrouter()
	presenterUngaurded := router.PathPrefix("/presenters").Subrouter()
	//presenter routes that require authentication
	presenterGaurded.Use(gJam.GeneralAuthentication)
	presenterGaurded.HandleFunc("", PresenterController.Create).Methods("POST")
	presenterGaurded.HandleFunc("/{id}", PresenterController.Update).Methods("PUT")
	//presenter routes that do not require authentication
	presenterUngaurded.HandleFunc("/{id}", PresenterController.Get).Methods("GET")

	/**
	* Image routes
	 */
	imageGuarded := router.PathPrefix("/images").Subrouter()
	imageGuarded.Use(gJam.GeneralAuthentication)
	imageGuarded.HandleFunc("", ImageController.Append).Methods("POST") //appends image to the parent
	imageGuarded.HandleFunc("/{id}", ImageController.Update).Methods("PATCH")
	imageGuarded.HandleFunc("/{id}", ImageController.Delete).Methods("DELETE")

	/**
	* Tag routes
	 */
	tagUnGuarded := router.PathPrefix("/tags").Subrouter()
	tagUnGuarded.HandleFunc("", TagController.Find).Methods("GET")

	/**
	* User routes
	 */
	userUnGuarded := router.PathPrefix("/users").Subrouter()
	//userGuarded := router.PathPrefix("/users").Subrouter()
	//userGuarded.Use(gJam.GeneralAuthentication)
	userUnGuarded.HandleFunc("/{id}/events", UserController.Events).Methods("GET")

	/**
	* Payment routes
	 */
	paymentUnGuarded := router.PathPrefix("/payments").Subrouter()
	paymentUnGuarded.HandleFunc("/webhooks/paystack", PaymentController.ProcessPaystackWebhook).Methods("POST")
}

func MakeRouter() *mux.Router {
	return router
}
