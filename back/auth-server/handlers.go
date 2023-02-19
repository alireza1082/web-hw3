package main

import (
	"encoding/json"
	"net/http"
	"time"
	"log"
	"database/sql"
	"fmt"
  
	_ "github.com/lib/pq"
	"github.com/golang-jwt/jwt/v4"
	"github.com/go-redis/redis"
	"golang.org/x/crypto/bcrypt"
)


func main() {
	http.HandleFunc("/signup", SignUp)
	http.HandleFunc("/signin", Signin)
	http.HandleFunc("/info", Info)
	http.HandleFunc("/refresh-token", RefreshToken)
	http.HandleFunc("/logout", Logout)

	initRedisFromInvalidTokens()

	log.Fatal(http.ListenAndServe(":8000", nil))
}


//this is my_secret_key for jwt
var jwtKey = []byte("fklnsdflsnflsnfl")

type User struct {
	UserId int `json:"user_id"`
	PhoneNumber string  `json:"phone_number"`
	Email string  `json:"email"`
	Gender string  `json:"gender"`
	FirstName string  `json:"first_name"`
	LastName string  `json:"last_name"`
}

type SignUpUser struct {
	Id int `json:"id"`
	PhoneNumber string  `json:"phone_number"`
	Email string  `json:"email"`
	Gender string  `json:"gender"`
	FirstName string  `json:"first_name"`
	LastName string  `json:"last_name"`
	Password string  `json:"password"`

}

// Create a struct that models the structure of a user, both in the request body, and in the DB
type Credentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

type Claims struct {
	Id int  `json:"id"`
	Email string `json:"email"`
	PhoneNumber string `json:"PhoneNumber"`
	FirstName string `json:"firstName"`
	LastName string `json:"lastName"`
	Gender string `json:"gender"`
	CType string `json:"token_type"`
	jwt.RegisteredClaims
}

type LoginResponse struct {
	TokenString string `json:"token"`
	RefreshTokenString string `json:"refresh_token"`

}

type ErrorMessage struct {
	Error       string `json:"error"`
}

func SignUp(w http.ResponseWriter, r *http.Request) {
	var creds SignUpUser

	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(creds.Email) == 0 || len(creds.Password) == 0|| len(creds.PhoneNumber) == 0 || len(creds.Gender) == 0|| len(creds.FirstName) == 0|| len(creds.LastName) == 0|| len(creds.Password) == 0|| (creds.Gender != "F" && creds.Gender != "M")  {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")

		var errorResponse ErrorMessage
		errorResponse.Error = "invalid request body"
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	

	var existedUser SignUpUser
	existedUser = getUserByPhoneNumberOrEmail(creds.PhoneNumber, creds.Email)

	if len(existedUser.Email) != 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")

		var errorResponse ErrorMessage
		errorResponse.Error = "user already existed with this email or phoneNumber"
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

    creds.Password = GeneratehashPassword(creds.Password)
	creds.Id = inserUserInDataBase(creds);

	expirationTime := time.Now().Add(5 * time.Minute)
	refreshTokenExpirationTime := time.Now().Add(60 * 5 * time.Minute)

	claims := &Claims{
		Id: creds.Id,
		Email: creds.Email,
		PhoneNumber: creds.PhoneNumber,
		Gender: creds.Gender,
		FirstName: creds.FirstName,
		LastName: creds.LastName,
		CType: "accessToken",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}


	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtKey)
	if err != nil {

		w.WriteHeader(http.StatusInternalServerError)
		return
	}


	refreshClaims := &Claims{
		Id: creds.Id,
		Email: creds.Email,
		PhoneNumber: creds.PhoneNumber,
		Gender: creds.Gender,
		FirstName: creds.FirstName,
		LastName: creds.LastName,
		CType: "refreshToken",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshTokenExpirationTime),
		},
	}


	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)

	refreshTokenString, err := refreshToken.SignedString(jwtKey)
	if err != nil {

		w.WriteHeader(http.StatusInternalServerError)
		return
	}


	var responseBody LoginResponse
	responseBody.TokenString = tokenString
	responseBody.RefreshTokenString = refreshTokenString
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(responseBody)
}

func Signin(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var existedUser SignUpUser
	existedUser = getUserByPhoneNumberOrEmail(creds.Username, creds.Username)

	if len(existedUser.Email) == 0 || !CheckPasswordHash(creds.Password, existedUser.Password)  {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")

		var errorResponse ErrorMessage
		errorResponse.Error = "username or password is not correct"
		json.NewEncoder(w).Encode(errorResponse)
		return
	}


	expirationTime := time.Now().Add(5 * time.Minute)
	refreshTokenExpirationTime := time.Now().Add(60 * 5 * time.Minute)

	claims := &Claims{
		Id: existedUser.Id,
		Email: existedUser.Email,
		PhoneNumber: existedUser.PhoneNumber,
		Gender:existedUser.Gender,
		FirstName: existedUser.FirstName,
		LastName: existedUser.LastName,
		CType: "accessToken",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}


	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtKey)
	if err != nil {

		w.WriteHeader(http.StatusInternalServerError)
		return
	}


	refreshClaims  := &Claims{
		Id: existedUser.Id,
		Email: existedUser.Email,
		PhoneNumber: existedUser.PhoneNumber,
		Gender:existedUser.Gender,
		FirstName: existedUser.FirstName,
		LastName: existedUser.LastName,
		CType: "refreshToken",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshTokenExpirationTime),
		},
	}


	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)

	refreshTokenString, err := refreshToken.SignedString(jwtKey)
	if err != nil {

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var responseBody LoginResponse
	responseBody.TokenString = tokenString
	responseBody.RefreshTokenString = refreshTokenString
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(responseBody)
}

func Info(w http.ResponseWriter, r *http.Request) {

	c := r.Header.Get("Authorization")
	if c == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if readFromRedis(c) == "invalid" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	tknStr := c

	claims := &Claims{}

	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if !tkn.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if  claims.CType == "refreshToken" {
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	var userResponse User
	userResponse.UserId = claims.Id
	userResponse.Email = claims.Email
	userResponse.PhoneNumber = claims.PhoneNumber
	userResponse.Gender = claims.Gender
	userResponse.FirstName = claims.FirstName
	userResponse.LastName = claims.LastName

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(userResponse)
}

func RefreshToken(w http.ResponseWriter, r *http.Request) {

	c := r.Header.Get("refreshToken")
	if c == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if readFromRedis(c) == "invalid" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	tknStr := c

	claims := &Claims{}

	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if !tkn.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if  claims.CType != "refreshToken" {
		w.WriteHeader(http.StatusUnauthorized)

		return
	}


	expirationTime := time.Now().Add(5 * time.Minute)
	refreshTokenExpirationTime := time.Now().Add(60 * 5 * time.Minute)

	newClaims := &Claims{
		Id: claims.Id,
		Email: claims.Email,
		PhoneNumber: claims.PhoneNumber,
		Gender:claims.Gender,
		FirstName: claims.FirstName,
		LastName: claims.LastName,
		CType: "accessToken",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}


	token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)

	tokenString, err := token.SignedString(jwtKey)
	if err != nil {

		w.WriteHeader(http.StatusInternalServerError)
		return
	}


	refreshClaims  := &Claims{
		Id: claims.Id,
		Email: claims.Email,
		PhoneNumber: claims.PhoneNumber,
		Gender:claims.Gender,
		FirstName: claims.FirstName,
		LastName: claims.LastName,
		CType: "refreshToken",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshTokenExpirationTime),
		},
	}


	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)

	refreshTokenString, err := refreshToken.SignedString(jwtKey)
	if err != nil {

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var responseBody LoginResponse
	responseBody.TokenString = tokenString
	responseBody.RefreshTokenString = refreshTokenString
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(responseBody)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	c := r.Header.Get("Authorization")
	if c == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	tknStr := c

	claims := &Claims{}

	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !tkn.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	inserInvalidTokenDataBase(claims.Id, c)
	writeInRedis(c)
}

////////////database


const (
	host     = "localhost"
	port     = 5432
	user     = "username"
	password = "password"
	dbname   = "auth"
)
  
func getUserByPhoneNumberOrEmail(phoenNumber string, email string) (result SignUpUser){
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
	"password=%s dbname=%s sslmode=disable",
	host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	sqlStatement := `SELECT * FROM user_account WHERE email=$1 OR phone_number=$2;`
	var user SignUpUser
	row := db.QueryRow(sqlStatement, email, phoenNumber)
	readError :=row.Scan(&user.Id, &user.Email, &user.PhoneNumber, &user.Gender, &user.FirstName, &user.LastName, &user.Password)
	result = user
	switch readError {
		// case sql.ErrNoRows:
		// 	fmt.Println("No rows were returned!")
		// case nil:
		// 	log.Println(result)
		// default:
		// 	log.Println(readError)
	}

	return
}

func inserUserInDataBase(userInfo SignUpUser) (id int) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
	"password=%s dbname=%s sslmode=disable",
	host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	sqlStatement := `
	INSERT INTO user_account (email, phone_number, first_name, last_name, gender, password_hash)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING user_id`
	user_id := 0
	err = db.QueryRow(sqlStatement, userInfo.Email, userInfo.PhoneNumber, userInfo.FirstName, userInfo.LastName, userInfo.Gender, userInfo.Password).Scan(&user_id)
	if err != nil {
	panic(err)
	}
	fmt.Println("New record ID is:", user_id)
	id = user_id
	return
}


func inserInvalidTokenDataBase(userId int, token string) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
	"password=%s dbname=%s sslmode=disable",
	host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	sqlStatement := `
	INSERT INTO unauthorized_token (user_id, token, expiration)
	VALUES ($1, $2, $3)`
	db.QueryRow(sqlStatement, userId, token, time.Now())
}


func isInvalidToken(token string) (invalid bool){
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
	"password=%s dbname=%s sslmode=disable",
	host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	sqlStatement := `SELECT count(*) FROM unauthorized_token WHERE token=$1;`
	var count = 0
	row := db.QueryRow(sqlStatement, token)
	row.Scan(&count)
	invalid = count != 0
	return
}

func getInvalidTokens() (invalidTokens []string){
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
	"password=%s dbname=%s sslmode=disable",
	host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	sqlStatement := `SELECT token FROM unauthorized_token;`
	rows, err := db.Query(sqlStatement)
	if err != nil {
        log.Fatal(err)
    }else {
		var result []string
		for rows.Next() {
			var next string 
			rows.Scan(&next)
			result = append(result,next)
		}
	
		invalidTokens = result
	}

	return
}


/////////////////redis

func writeInRedis(token string) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		Password: "",
		DB: 0,
	})

	client.Set(token, "invalid", 0).Err()
	
}

func readFromRedis(token string) (result string){
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		Password: "",
		DB: 0,
	})

	val, err := client.Get(token).Result()
	if err != nil {
    	// fmt.Println(err)
	}

	result = val

	return
}

func initRedisFromInvalidTokens() {
	for _,v := range getInvalidTokens() {

		writeInRedis(v)
	}
}

/////////////utils


func GeneratehashPassword(password string) (string) {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes)
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}