package controllers

import (
	"fmt"
	"log"
	"net/http"

	"bitbucket.org/softwarehouseio/victory/victory-frontend/config"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/db"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/models"
	"github.com/flosch/pongo2"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

type AdminEmail struct {
	BaseURL string
}

func (s *AdminEmail) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/{templateName}/", s.Preview)
	return r
}

func (s *AdminEmail) Preview(w http.ResponseWriter, r *http.Request) {

	var (
		tplContext = WebResource{}.GetTplContext(r)
		tx         = db.GetDBFromRequestContext(r)
	)
	tplKey := chi.URLParam(r, "templateName")

	tmplt, ok := config.EmailTemplates[tplKey]
	if !ok {
		log.Println("Tpl.ExecuteWriter Failed")
		e := ErrInternalServerError(nil)
		render.Render(w, r, e)
		return
	}

	tpl := pongo2.Must(pongo2.FromFile(tmplt))

	order := models.Order{}
	tx.Model(&order).
		Preload("ShippingAddress").
		Preload("OrderItems").
		Preload("OrderItems.ProductVariation").
		Preload("OrderItems.ProductVariation.Size").
		Preload("OrderItems.ProductVariation.Product").
		First(&order)

	tplContext.Update(pongo2.Context{
		"order": &order,
	})

	if err := tpl.ExecuteWriter(tplContext, w); err != nil {
		fmt.Printf("Tpl.ExecuteWriter Failed %v", err)
		e := ErrInternalServerError(err)
		render.Render(w, r, e)
		return
	}
}
