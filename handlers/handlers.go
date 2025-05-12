package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ArgenisGutierrez/recipes-api/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type RecipesHandler struct {
	collection *mongo.Collection
	ctx        context.Context
}

func NewRecipesHandler(ctx context.Context, collection *mongo.
	Collection) *RecipesHandler {
	return &RecipesHandler{
		collection: collection,
		ctx:        ctx,
	}
}

// swagger:operation GET /recipes recipes listRecipes
// Returns list of recipes
// ---
// produces:
// - application/json
// responses:
// '200':
// description: Successful operation
func (handler *RecipesHandler) ListRecipesHandler(c *gin.
	Context) {
	cur, err := handler.collection.Find(handler.ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}
	defer cur.Close(handler.ctx)
	recipes := make([]models.Recipe, 0)
	for cur.Next(handler.ctx) {
		var recipe models.Recipe
		cur.Decode(&recipe)
		recipes = append(recipes, recipe)
	}
	c.JSON(http.StatusOK, recipes)
}

func (handler *RecipesHandler) NewRecipeHandler(ctx *gin.Context) {
	var recipe models.Recipe
	if err := ctx.ShouldBindJSON(&recipe); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error()})
		return
	}
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
func (handler *RecipesHandler) UpdateRecipeHandler(ctx *gin.Context) {
	id := ctx.Param("id")
	var recipe models.Recipe
	if err := ctx.ShouldBindJSON(&recipe); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	objectId, _ := primitive.ObjectIDFromHex(id)
	_, err := handler.collection.UpdateOne(ctx, bson.M{
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
func (handler *RecipesHandler) DeleteRecipeHandler(ctx *gin.Context) {
	id := ctx.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)
	_, err := handler.collection.DeleteOne(ctx, bson.M{"_id": objectId})
	if err != nil {
		fmt.Println(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Receta eliminada"})
}
