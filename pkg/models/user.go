package models

import (
	"context"
	"encoding/json"
	"fmt"
	"go-crud-server/pkg/config"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

var db *gorm.DB
var ctx = context.Background()
var rdb *redis.Client

// User model
type User struct {
	ID        string    `gorm:"type:char(36);primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(100)" json:"name"`
	Email     string    `gorm:"type:varchar(100);unique" json:"email"`
	Location  string    `gorm:"type:varchar(100)" json:"location"`
	CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updated_at"`
	IsDeleted bool      `gorm:"type:boolean;default:false" json:"is_deleted"`
}

func init() {
	config.Connect()
	db = config.GetDB()     // get db
	rdb = config.GetRedis() // get redis client
	db.AutoMigrate(&User{})
}

// Create new user's id 
func (user *User) BeforeCreate(tx *gorm.DB) (err error) {
	user.ID = uuid.NewString()
	return
}

// Get User by id
func GetUserById(Id string) (*User, *gorm.DB) {

	var getUser User
	redisKey := "user:" + Id
	fmt.Println("Redis Key:", redisKey)

	cachedUser, err := rdb.Get(ctx, redisKey).Result()

	if err == redis.Nil {

		fmt.Println("User not found in cache, querying the database")

		result := db.Where("ID = ? AND is_deleted = ?", Id, false).First(&getUser) //user not in cashe find it in DB

		if result.Error != nil { //user not found
			if result.Error == gorm.ErrRecordNotFound {
				return nil, result
			}
			return nil, result
		}

		userJson, err := json.Marshal(getUser)
		if err == nil {
			rdb.Set(ctx, redisKey, userJson, 10*time.Minute) //cashing  user with 10 min expiry
		}
		return &getUser, result
	} else if err != nil {
		return nil, nil	//redis connection error
	}

	// User found in Redis, unmarshal it
	err = json.Unmarshal([]byte(cachedUser), &getUser)
	if err != nil {
		return nil, nil
	}

	fmt.Println("User details returned from cache.")
	return &getUser, db
}


// Delete User
func DeleteUser(ID string) (User, error) {
	var user User

	if err := db.First(&user, "id = ?", ID).Error; err != nil {
		return user, err //error if user is not found
	}

	user.IsDeleted = true

	defer rdb.Close()

	if err := ClearUserCache(ID); err != nil {
		log.Fatalf("Error clearing cache for user %s: %v", ID, err)
	}
	log.Printf("Cache cleared for user %s", ID)


	if err := db.Save(&user).Error; err != nil {
		return user, err //  an error if saving fails
	}
	return user, nil // Return the updated user
}

// pagination
func GetUsers(location string, page, limit int) ([]User, error) {
	var Users []User
	fmt.Println("Fetching users for location:", location)

	offset := (page - 1) * limit  //calculate offset

	result := db.Where("location = ? AND is_deleted = ?", location, false).Offset(offset).Limit(limit).Find(&Users)

	if result.Error != nil {
		return nil, result.Error
	}
	return Users, nil
}

func ClearUserCache(userID string) error {

	key := fmt.Sprintf("user:%s", userID) //construsting the key pattern for user id

	_, err := rdb.Del(context.Background(), key).Result()

	return err
}
