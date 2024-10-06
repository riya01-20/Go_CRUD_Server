package controllers

import (
	"encoding/json"
	"fmt"
	"go-crud-server/pkg/config"
	"go-crud-server/pkg/models"
	"go-crud-server/pkg/utils"
	"log"

	"net/http"
	"strconv"

	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)


func GetUsers(w http.ResponseWriter, r *http.Request) {

	queryParams := r.URL.Query() 

	location := queryParams.Get("location")
	if location == "" {
		location = "Bengaluru" // Default location
	}

	pageQuery := queryParams.Get("page")
	if pageQuery == "" {
		pageQuery = "1" // Default page
	}
	page, err := strconv.Atoi(pageQuery)

	if err != nil || page < 1 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Invalid or missing page parameter"}`))
		return
	}

	limitQuery := queryParams.Get("limit")
	if limitQuery == "" {
		limitQuery = "10" // Default limit
	}
	limit, err := strconv.Atoi(limitQuery)

	if err != nil || limit <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Invalid or missing limit parameter"}`))
		return
	}

	fmt.Printf("page: %d, limit: %d, location: %s\n", page, limit, location)

	users, err := models.GetUsers(location, page, limit)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(`{"error": "An error occurred: %s"}`, err.Error())))
		return
	}

	// Convert the users list to JSON and send it in the response
	res, _ := json.Marshal(users)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}

func GetUserById(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["id"]

	// Call the modified GetUserById function
	userDetails, dbResult := models.GetUserById(id)

	// Handle the case where the user is not found or is soft-deleted
	if dbResult != nil && dbResult.Error != nil {
		if dbResult.Error == gorm.ErrRecordNotFound {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "User does not exist"}`))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "An internal error occurred"}`))
		}
		return
	}

	if userDetails == nil {	//if details are nil
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "User not found"}`))
		return
	}

	res, _ := json.Marshal(userDetails)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}

func CreateUser(w http.ResponseWriter, r *http.Request) {

	CreateUser := &models.User{}

	utils.ParseBody(r, CreateUser)

	db := config.GetDB()

	err := db.Create(&CreateUser).Error		// Creating User

	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {		//check for duplicate emails
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(`{"error": "Email already exists"}`))
			return
		}

		w.WriteHeader(http.StatusInternalServerError)	//other errors
		w.Write([]byte(fmt.Sprintf(`{"error": "An error occurred: %s"}`, err.Error())))
		return
	}

	res, _ := json.Marshal(CreateUser)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}



func DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ID := vars["id"]

	user, err := models.DeleteUser(ID)	//Delete with user id

	if err != nil {
		w.Header().Set("Content-Type", "application/json")	//didn't found user or any error
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	
	res, _ := json.Marshal(user)	//store result
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	var updateUser = &models.User{}
	utils.ParseBody(r, updateUser)
	vars := mux.Vars(r)

	fmt.Println(vars)
	id := vars["id"]

	userDetails, db := models.GetUserById(id)

	// Check for database connection
	if db == nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	// Check if user was found
	if userDetails == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if updateUser.Name != "" {
		userDetails.Name = updateUser.Name
	}
	if updateUser.Email != "" {
		userDetails.Email = updateUser.Email
	}
	if updateUser.Location != "" {
		userDetails.Location = updateUser.Location
	}
	// Save updated user
	if err := db.Save(userDetails).Error; err != nil {
		http.Error(w, "Error updating user", http.StatusInternalServerError)
		return
	}

	// fmt.Println("hiiiii")
	if err := models.ClearUserCache(id); err != nil {
		log.Printf("Error clearing cache for user %s: %v", id, err)
	} else {
		log.Printf("Cache cleared for user %s", id)
	}

	// fmt.Println("User Details", userDetails)
	res, _ := json.Marshal(userDetails)
	w.Header().Set("Content-Type", "pkglication/json")
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}
