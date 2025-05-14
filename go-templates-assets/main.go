package main

import (
	"embed"
	"encoding/json"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type Recipe struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Ingredients []Ingredient `json:"ingredients"`
	Steps       []string     `json:"steps"`
	Picture     string       `json:"imageURL"`
}
type Ingredient struct {
	Quantity string `json:"quantity"`
	Name     string `json:"name"`
	Type     string `json:"type"`
}

//go:embed assets/* templates/* 404.html recipes.json
var f embed.FS

var recipes []Recipe

func init() {
	recipes = make([]Recipe, 0)
	data, _ := f.ReadFile("recipes.json")
	json.Unmarshal(data, &recipes)
}

func IndexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"recipes": recipes,
	})
}

func RecipeHandler(c *gin.Context) {
	for _, recipe := range recipes {
		if recipe.ID == c.Param("id") {
			c.HTML(http.StatusOK, "recipe.tmpl", gin.H{
				"recipe": recipe,
			})
			return
		}
	}
	c.File("404.html")
}

func loadTemplate() (*template.Template, error) {
	t := template.New("")
	for name, file := range Assets.Files {
		if file.IsDir() || !strings.HasSuffix(name, ".tmpl") {
			continue
		}
		h, err := io.ReadAll(file)
		if err != nil {
			return nil, err
		}
		t, err = t.New(name).Parse(string(h))
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}

func StaticHandler(c *gin.Context) {
	filepath := c.Param("filepath")
	data := Assets.Files["/assets"+filepath].Data
	c.Writer.Write(data)
}

func main() {
	templ := template.Must(template.New("").ParseFS(f, "templates/*.tmpl"))
	fsys, err := fs.Sub(f, "assets")
	if err != nil {
		panic(err)
	}
	router := gin.Default()
	router.SetHTMLTemplate(templ)
	router.StaticFS("/assets", http.FS(fsys))
	router.GET("/", IndexHandler)
	router.GET("/recipes/:id", RecipeHandler)
	router.Run()
}
