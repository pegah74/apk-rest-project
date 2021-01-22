package main

import (
	"context"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
)

func main() {
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}
	e.Use(middleware.CORS())

	e.GET("/", homepage)
	e.POST("/api/v1/auth/register", register)

	e.Logger.Fatal(e.Start(":9000"))
}

//database ####################################################

func db() *mongo.Client {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")
	return client
}

var userCollection = db().Database("ApkProject").Collection("users")

func saveUser(user *User) (*mongo.InsertOneResult, error) {
	return userCollection.InsertOne(context.TODO(), user)
}

func findUserByMobile(mobile string) (User, error) {
	var result User
	err := userCollection.FindOne(context.TODO(), bson.D{{"mobile", mobile}}).Decode(&result)
	return result, err
}

//routes ####################################################
func homepage(c echo.Context) error {
	return c.String(http.StatusOK, "APK Rest Service is up!")
}

func register(c echo.Context) error {

	u := new(User)
	if err := c.Bind(u); err != nil {
		return err
	}

	if err := c.Validate(u); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	userInDB, userError := findUserByMobile(u.Mobile)
	if userError == nil {
		return c.JSON(http.StatusConflict, "mobile "+userInDB.Mobile+" already exists")
	}

	//hash password
	u.Password = hashAndSalt([]byte(u.Password))

	//save db
	_, err := saveUser(u)
	if err != nil {
		log.Println(err.Error())
		return c.JSON(http.StatusFailedDependency, "user can not be saved on db. ")
	}
	return c.JSON(http.StatusCreated, "user saved on db")
}

//utils ####################################################

func hashAndSalt(pwd []byte) string {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}
	return string(hash)
}
func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

//structs ####################################################

type (
	User struct {
		Firstname string `json:"firstname"  form:"firstname" query:"firstname" `
		Lastname  string `json:"lastname"  form:"lastname" query:"lastname" `
		Mobile    string `json:"mobile"  form:"mobile" query:"mobile" validate:"required,len=11"`
		Birthdate int32  `json:"birthdate"  form:"birthdate" query:"birthdate"`
		Gender    string `json:"gender"  form:"gender" query:"gender"`
		Password  string `json:"password"  form:"password" query:"password" validate:"required,min=4"`
	}
	CustomValidator struct {
		validator *validator.Validate
	}
)
