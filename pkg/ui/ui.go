package ui

import (
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Init() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		// Pass the header HTML to the template
		t, err := template.ParseFiles("./ui/views/index.html")
		r.Static("/css", "./ui/assets/css")
		r.Static("/js", "./ui/assets/js")

		if err != nil {
			c.String(http.StatusInternalServerError, "Error loading template: %v", err)
			return
		}

		// // Render the header content
		// headerHTML := RenderHeaderTemplate("My Website") // Custom function to render the header

		// // Create a data map to pass variables to the template
		// data := map[string]interface{}{
		// 	"Header": template.HTML(headerHTML), // Pass the header HTML as safe HTML
		// }

		// Render the template with the data
		err = t.Execute(c.Writer, "")

		if err != nil {
			c.String(http.StatusInternalServerError, "Error rendering template: %v", err)
			return
		}
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
