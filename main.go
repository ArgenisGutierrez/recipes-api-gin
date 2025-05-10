// Package main API para gestionar recetas
package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

// Recipe Estrutura de una receta
type Recipe struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Tags         []string  `json:"tags"`
	Ingredients  []string  `json:"ingredients"`
	Instructions []string  `json:"instructions"`
	PulishedAt   time.Time `json:"publishedAt"`
}

var recipes []Recipe

func init() {
	recipes = make([]Recipe, 0)
	file, _ := ioutil.ReadFile("recipes.json")
	_ = json.Unmarshal([]byte(file), &recipes)
}

// NewRecipeHandler crear una receta
func NewRecipeHandler(ctx *gin.Context) {
	var recipe Recipe
	if err := ctx.ShouldBindJSON(&recipe); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error()})
		return
	}
	recipe.ID = xid.New().String()
	recipe.PulishedAt = time.Now()
	recipes = append(recipes, recipe)
	ctx.JSON(http.StatusOK, recipe)
}

// ListRecipesHandler Listar recetas
func ListRecipesHandler(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, recipes)
}

func UpdateRecipeHandler(ctx *gin.Context) {
	id := ctx.Params.ByName("id")
	var recipe Recipe
	if err := ctx.ShouldBindJSON(&recipe); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	index := -1
	for i := 0; i < len(recipes); i++ {
		if recipes[i].ID == id {
			index = i
			break
		}
	}

	if index == -1 {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "No se encuentra la receta",
		})
		return
	}

	recipes[index] = recipe
	ctx.JSON(http.StatusOK, recipe)
}

func DeleteRecipeHandler(ctx *gin.Context) {
	id := ctx.Params.ByName("id")
	index := -1
	for i := 0; i < len(recipes); i++ {
		if recipes[i].ID == id {
			index = i
		}
	}

	if index == -1 {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "No se encuentra la receta",
		})
		return
	}

	recipes = append(recipes[:index], recipes[index+1:]...)
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Receta borrada",
	})
}

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
