package main

import (
	"fmt"
	"crypto/rand"
	"context"
	"strings"
	"regexp"
	"encoding/json"
	"github.com/JeanServices/myAPI-Components"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// User is the base struct for general users
type User struct {
	Token		string 		`bson:"token" json:"token"`
	ID			string		`bson:"id" json:"id"`
	Name		string		`bson:"name" json:"name"`
	Username	string 		`bson:"username" json:"username"`
	Mail		string		`bson:"mail" json:"mail"`
	Password	string		`bson:"password" json:"password"`
}

type GetMe struct {
	Token	string		`json:"token"`
}

func main() {
	API(jeanservices.MyMongo("your-mongo-uri"))
}

// API is the main function to initialize the server and register the routes
func API(client *mongo.Client) {
	server := fiber.New()
	server.Use(cors.New())
	
	//jeanservices.RegisterRoutes(server, client)
	RegisterRoutes(server, client)

	server.Listen(":5000")
}

// RegisterRoutes register all routes
func RegisterRoutes(server *fiber.App, client *mongo.Client) {
	server.Use("/api/v1", func(c *fiber.Ctx) error {
		c.Accepts("application/json")
		return c.Next()
	})

	server.Get("/api/v1", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": 200,
			"message": "ok",
			"data": nil,
			"exited_code": 0,
		})
	})
	
	server.Get("/api/v1/get/public/user/:id", func (c *fiber.Ctx) error {
		data := make(map[string]interface{})
		err := client.Database("myAPI").Collection("users").FindOne(context.TODO(), bson.D{{"id", c.Params("id")}}).Decode(&data)
		if err != nil {
			return c.JSON(fiber.Map{
				"status": 500,
				"message": err.Error(),
				"data": nil,
				"exited_code": 1,
			})
		}

		data["mail"] = nil
		data["password"] = nil

		return c.JSON(fiber.Map{
			"status": 200,
			"message": "obtained",
			"data": data,
			"exited_code": 0,
		})
	})

	server.Post("/api/v1/get/me/user/:id", func (c *fiber.Ctx) error {
		body := make(map[string]interface{})
		err := json.Unmarshal([]byte(string(c.Body())), &body)
		if err != nil {
			return c.JSON(fiber.Map{
				"status": 500,
				"message": err.Error(),
				"data": nil,
				"exited_code": 1,
			})
		}
		
		if body["token"] == nil {
			return c.JSON(fiber.Map{
				"status": 500,
				"message": "the raw body need to have: token",
				"data": nil,
				"exited_code": 1,
			})
		}

		var data = make(map[string]interface{})
		err = client.Database("myAPI").Collection("users").FindOne(context.TODO(), bson.D{{"id", c.Params("id")}}).Decode(&data)
		if err != nil {
			return c.JSON(fiber.Map{
				"status": 500,
				"message": err.Error(),
				"data": nil,
				"exited_code": 1,
			})
		}

		if data["token"] == nil {
			return c.JSON(fiber.Map{
				"status": 500,
				"message": "user document dont have token",
				"data": nil,
				"exited_code": 1,
			})
		}

		if strings.Compare(data["token"].(string), body["token"].(string)) != 0 {
			return c.JSON(fiber.Map{
				"status": 500,
				"message": "invalid token",
				"data": nil,
				"exited_code": 1,
			})
		}

		return c.JSON(fiber.Map{
			"status": 200,
			"message": "the id and token match, and the data has been sent",
			"data": data,
			"exited_code": 0,
		})
	})

	server.Post("/api/v1/post/user/create", func(c *fiber.Ctx) error {
		body := make(map[string]interface{})
		err := json.Unmarshal([]byte(string(c.Body())), &body)
		if err != nil {
			return c.JSON(fiber.Map{
				"status": 500,
				"message": err.Error(),
				"data": nil,
				"exited_code": 1,
			})
		}

		if body["username"] == nil || body["name"] == nil || body["mail"] == nil && body["password"] == nil && body["confirm_password"] == nil {
			return c.JSON(fiber.Map{
				"status": 500,
				"message": "the raw body need to have: username, mail, password and confirm_password",
				"data": nil,
				"exited_code": 0,
			})
		}

		if len(body["username"].(string)) >= 25 || len(body["username"].(string)) <= 4  {
			return c.JSON(fiber.Map{
				"status": 500,
				"message": "the username need to be lenght more than 4 and less of 25 characters",
				"data": nil,
				"exited_code": 0,
			})
		}

		userData := make(map[string]interface{})
		err = client.Database("myAPI").Collection("users").FindOne(context.TODO(), bson.D{{"username", body["username"]}}).Decode(&userData)
		if userData["username"] != nil {
			return c.JSON(fiber.Map{
				"status": 500,
				"message": "existing user with that username",
				"data": nil,
				"exited_code": 0,
			})
		}

		if len(body["name"].(string)) >= 100 || len(body["name"].(string)) <= 1  {
			return c.JSON(fiber.Map{
				"status": 500,
				"message": "the username need to be lenght more than 1 and less of 100 characters",
				"data": nil,
				"exited_code": 0,
			})
		}

		var mailRegexp = regexp.MustCompile(`^(([^<>()[\]\.,;:\s@\"]+(\.[^<>()[\]\.,;:\s@\"]+)*)|(\".+\"))@(([^<>()[\]\.,;:\s@\"]+\.)+[^<>()[\]\.,;:\s@\"]{2,})$`)
		if !mailRegexp.MatchString(body["mail"].(string)) {
			return c.JSON(fiber.Map{
				"status": 500,
				"message": "not valid mail",
				"data": nil,
				"exited_code": 0,
			})
		}

		err = client.Database("myAPI").Collection("users").FindOne(context.TODO(), bson.D{{"mail", body["mail"]}}).Decode(&userData)
		if userData["username"] != nil {
			return c.JSON(fiber.Map{
				"status": 500,
				"message": "existing user with that mail",
				"data": nil,
				"exited_code": 0,
			})
		}

		if len(body["password"].(string)) <= 8 || len(body["password"].(string)) >= 1000 {
			return c.JSON(fiber.Map{
				"status": 500,
				"message": "the password need to be lenght more than 4 and less of 1000 characters",
				"data": nil,
				"exited_code": 0,
			})
		}

		if strings.Compare(body["password"].(string), body["confirm_password"].(string)) != 0 {
			return c.JSON(fiber.Map{
				"status": 500,
				"message": "the password doesn't match with confirm password",
				"data": nil,
				"exited_code": 0,
			})
		}


		token := tokenGenerator(100)
		id := tokenGenerator(20)
		user := User{
			Token: token,
			ID: id,
			Name: body["name"].(string),
			Username: body["username"].(string),
			Mail: body["mail"].(string),
			Password: body["password"].(string),
		}

		_, err = client.Database("myAPI").Collection("users").InsertOne(context.TODO(), user)
		if err != nil {
			return c.JSON(fiber.Map{
				"status": 500,
				"message": err.Error(),
				"data": nil,
				"exited_code": 0,
			})
		}

		return c.JSON(fiber.Map{
			"status": 200,
			"message": "ok",
			"data": fiber.Map{
				"token": token,
				"id": id,
			},
			"exited_code": 0,
		})
	})
}

func tokenGenerator(count int) string {
	b := make([]byte, count)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
