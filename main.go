// Recipes API
//
// This is a sample recipes API. You can find out more about the API at https://github.com/ArgenisGutierrez/recipes-api-gin
//
// Schemes: http
//
// Host: localhost:8080
// BasePath: /
// Version: 1.0.0
// Contact: Argenis Gutierrez <argenis.v.gtz@gmail.com> https://github.com/ArgenisGutierrez
// License: MIT https://opensource.org/licenses/MIT
//
// Consumes:
// - application/json
//
// Produces:
// - application/json
// swagger:meta
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// swagger:parameters recipe newRecipe
type Recipe struct {
	ID           primitive.ObjectID `json:"id"           bson:"_id"`
	Name         string             `json:"name"         bson:"name"`
	Tags         []string           `json:"tags"         bson:"tags"`
	Ingredients  []string           `json:"ingredients"  bson:"ingredients"`
	Instructions []string           `json:"instructions" bson:"instructions"`
	PulishedAt   time.Time          `json:"publishedAt"  bson:"publishedAt"`
}

var recipes []Recipe
var ctx context.Context
var err error
var client *mongo.Client
var collection *mongo.Collection

func init() {
	ctx := context.Background()
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	collection = client.Database(os.Getenv(
		"MONGO_DATABASE")).Collection("recipes")
	log.Println("conected to mongo")
}

// swagger:operation POST /recipes recipes newRecipe
// Returns a NewRecipeHandler
// ---
// produces:
// - application/json
// responses:
//
//	'200':
//		description: Successful operation
//	'400':
//		description: Invalid data
func NewRecipeHandler(ctx *gin.Context) {
	var recipe Recipe
	if err := ctx.ShouldBindJSON(&recipe); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error()})
		return
	}

	recipe.ID = primitive.NewObjectID()
	recipe.PulishedAt = time.Now()
	_, err := collection.InsertOne(ctx, recipe)
	if err != nil {
		fmt.Println(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error al insertar nueva receta"})
		return
	}
	ctx.JSON(http.StatusOK, recipe)
}

// swagger:operation GET /recipes recipes listRecipes
// Returns list of recipes
// ---
// produces:
// - application/json
// responses:
// '200':
// description: Successful operation
func ListRecipesHandler(ctx *gin.Context) {
	cur, err := collection.Find(ctx, bson.M{})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cur.Close(ctx)

	recipes := make([]Recipe, 0)
	for cur.Next(ctx) {
		var recipe Recipe
		cur.Decode(&recipe)
		recipes = append(recipes, recipe)
	}

	ctx.JSON(http.StatusOK, recipes)
}

// swagger:operation PUT /recipes/{id} recipes updateRecipe
// Returns a updated recipe
// ---
// parameters:
//   - name: id
//     in: path
//     description: ID of the recipe to update
//     required: true
//     type: string
//
// produces:
//   - application/json
//
// responses:
//
//	'200':
//	  description: Operación exitosa
//	'400':
//	  description: ID inválido
//	'404':
//	  description: Receta no encontrada
func UpdateRecipeHandler(ctx *gin.Context) {
	id := ctx.Param("id")
	var recipe Recipe
	if err := ctx.ShouldBindJSON(&recipe); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	objectId, _ := primitive.ObjectIDFromHex(id)
	_, err = collection.UpdateOne(ctx, bson.M{
		"_id": objectId,
	}, bson.D{{"$set", bson.D{
		{"name", recipe.Name},
		{"instructions", recipe.Instructions},
		{"ingredients", recipe.Ingredients},
		{"tags", recipe.Tags},
	}}})

	if err != nil {
		fmt.Println(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Receta actualizada"})

}

// swagger:operation DELETE /recipes/{id} recipes deleteRecipe
// returns a deleted recipe
// ---
// parameters:
//   - name: id
//     in: path
//     description: ID de la receta
//     required: true
//     type: string
//
// produces:
//   - application/json
//
// responses:
//
//	'200':
//		description: Successful operation
//	'404':
//		description: Recipe not found
func DeleteRecipeHandler(ctx *gin.Context) {
	id := ctx.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)
	_, err := collection.DeleteOne(ctx, bson.M{"_id": objectId})
	if err != nil {
		fmt.Println(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Receta eliminada"})
}

// swagger:operation GET /recipes/search recipes searchRecipes
// returns a list of recipes
// ---
// parameters:
//   - name: tag
//     in: query
//     description: Tag para buscar
//     required: true
//     type: string
//
// responses:
//
//	'200':
//		description: Successful operation
//	'404':
//		description: No recipes found
func SearchRecipesHandler(ctx *gin.Context) {
	tag := ctx.Query("tag")
	listOfRecipes := make([]Recipe, 0)
	for i := 0; i < len(recipes); i++ {
		found := false
		for _, t := range recipes[i].Tags {
			if strings.EqualFold(t, tag) {
				found = true
			}
		}
		if found {
			listOfRecipes = append(listOfRecipes, recipes[i])
		}
	}
	if len(listOfRecipes) == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "No se encuentra ninguna receta con el tag " + tag,
		})
		return
	}

	ctx.JSON(http.StatusOK, listOfRecipes)
}

// main Ejecutar el servidor
func main() {
	router := gin.Default()
	router.POST("/recipes", NewRecipeHandler)
	router.GET("/recipes", ListRecipesHandler)
	router.PUT("/recipes/:id", UpdateRecipeHandler)
	router.DELETE("/recipes/:id", DeleteRecipeHandler)
	router.GET("/recipes/search", SearchRecipesHandler)
	router.Run()
}
