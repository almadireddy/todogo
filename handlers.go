package main

import (
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"time"
)

func ListItems(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	params := mux.Vars(request)
	limit, err := strconv.Atoi(params["limit"])
	if err != nil {
		limit = 10
	}

	c, err := request.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			response.WriteHeader(http.StatusUnauthorized)
			_, _ = response.Write([]byte( `{ "message3": "You are not signed in."}` ))
			return
		}
		response.WriteHeader(http.StatusBadRequest)
		return
	}

	tokenString := c.Value
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			response.WriteHeader(http.StatusUnauthorized)
			return
		}
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	if !token.Valid {
		response.WriteHeader(http.StatusUnauthorized)
		_, _ = response.Write([]byte( `{ "message3": "Token not valid"}` ))
		return
	}

	items, err := todoItemStore.List(limit, claims.UserId)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		_, _ = response.Write([]byte( `{ "message3": '` + err.Error() + `'}` ))
		return
	}
	_ = json.NewEncoder(response).Encode(&items)
}

func AddItem(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	var item TodoItem

	c, err := request.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			response.WriteHeader(http.StatusUnauthorized)
			_, _ = response.Write([]byte( `{ "message3": "You are not signed in."}` ))
			return
		}
		response.WriteHeader(http.StatusBadRequest)
		return
	}

	tokenString := c.Value
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			response.WriteHeader(http.StatusUnauthorized)
			return
		}
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	if !token.Valid {
		response.WriteHeader(http.StatusUnauthorized)
		_, _ = response.Write([]byte( `{ "message3": "Token not valid"}` ))
		return
	}

	err = json.NewDecoder(request.Body).Decode(&item)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		_, _ = response.Write([]byte( `{ "message": '` + err.Error() + `'}` ))
		return
	}

	item.UserID = claims.UserId
	insertedItem, err := todoItemStore.Create(&item)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		_, _ = response.Write([]byte( `{ "message": '` + err.Error() + `'}` ))
		return
	}

	_ = json.NewEncoder(response).Encode(insertedItem)
}


func SignUp(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	var user User

	err := json.NewDecoder(request.Body).Decode(&user)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		_, _ = response.Write([]byte( `{ "message": '` + err.Error() + `'}` ))
		return
	}

	newUser, err := userStore.Create(&user, string(user.Password))
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		_, _ = response.Write([]byte( `{ "message": '` + err.Error() + `'}` ))
		return
	}

	issueToken(response, newUser)

	_ = json.NewEncoder(response).Encode(newUser)
}

func issueToken(response http.ResponseWriter, user *User) {
	expirationTime := time.Now().Add(5*time.Minute)

	claims := &Claims{
		Username: user.UserName,
		UserId: user.UserID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		fmt.Println("Here, ", tokenString, token)
		response.WriteHeader(http.StatusInternalServerError)
		_, _ = response.Write([]byte( `{ "message": "` + err.Error() + `"}` ))
		return
	}

	http.SetCookie(response, &http.Cookie{
		Name: "token",
		Value: tokenString,
		Expires: expirationTime,
	})
}


type Claims struct {
	Username string `json:"username"`
	UserId int `json:"userId"`
	jwt.StandardClaims
}

func SignIn(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	var user User

	err := json.NewDecoder(request.Body).Decode(&user)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		_, _ = response.Write([]byte( `{ "message": '` + err.Error() + `'}` ))
		return
	}

	valid, err := userStore.Validate(string(user.Password), user.UserName)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		_, _ = response.Write([]byte( `{ "message": '` + err.Error() + `'}` ))
		return
	}

	if !valid {
		response.WriteHeader(http.StatusUnauthorized)
		_, _ = response.Write([]byte( `{ "message": "Your username or password is incorrect"` ))
		return
	}

	dbUser, err := userStore.GetByUsername(user.UserName)
	if err != nil {
		response.WriteHeader(http.StatusUnauthorized)
		_, _ = response.Write([]byte( `{ "message": "Error getting user from db" }` ))
		return
	}

	issueToken(response, dbUser)

	_ = json.NewEncoder(response).Encode(dbUser)
}

func Refresh(response http.ResponseWriter, request *http.Request) {
	c, err := request.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			response.WriteHeader(http.StatusUnauthorized)
			return
		}
		response.WriteHeader(http.StatusBadRequest)
		return
	}

	tokenString := c.Value
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			response.WriteHeader(http.StatusUnauthorized)
			return
		}
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	if !token.Valid {
		response.WriteHeader(http.StatusUnauthorized)
		return
	}

	if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) > 30*time.Second {
		response.WriteHeader(http.StatusBadRequest)
		return
	}

	expirationTime := time.Now().Add(5 * time.Minute)
	claims.ExpiresAt = expirationTime.Unix()
	token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString(jwtKey)

	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(response, &http.Cookie{
		Name: "token",
		Value: tokenString,
		Expires: expirationTime,
	})
}