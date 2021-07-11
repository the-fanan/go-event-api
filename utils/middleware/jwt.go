package middleware

import (
	"goventy/utils"
	"goventy/utils/response"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
)

type JwtAuthMiddleware struct {
	Roles    []string //what are the roles that are permitted to pass
	Provider string   //what table are we going to use to fetch user?
}

/**
*	This method grants passage to users as long as their JWT is valid
* It does not check to see if the JWT's "role" claim mathces the roles stated in the Struct
 */
func (jam *JwtAuthMiddleware) GeneralAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		/**
		* Prepare error response before hand
		 */
		errorResponse := &response.Error{Status: "error", Message: "Token has expired"}
		/**
		* Handle JWT validation
		 */
		//get headers and manipulate
		authorization := r.Header.Get("Authorization")
		if authorization == "" {
			response.RespondUnauthorized(w, errorResponse)
			return
		}
		token, err := utils.GetJwtToken(strings.Split(authorization, " ")[1])
		if err != nil {
			response.RespondUnauthorized(w, errorResponse)
			return
		} else {
			if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				next.ServeHTTP(w, r)
			} else {
				response.RespondUnauthorized(w, errorResponse)
				return
			}
		}
	})
}
