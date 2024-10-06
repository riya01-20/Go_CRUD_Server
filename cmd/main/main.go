package main

import (
	"fmt"
	"go-crud-server/pkg/config"
	"go-crud-server/pkg/models"
	"go-crud-server/pkg/routes"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func main() {
	config.Connect()
	db := config.GetDB()
	db.AutoMigrate(&models.User{})
	r := mux.NewRouter()
	routes.UserRoutes(r)
	http.Handle("/", r)
	fmt.Println("......Serving.....")
	log.Fatal(http.ListenAndServe("localhost:3000", r))
}
