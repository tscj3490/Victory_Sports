package controllers

import (
	"context"
	"fmt"
	"net/http"

	"bitbucket.org/softwarehouseio/victory/victory-frontend/db"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/models"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"
)

type OrdersAdminResource struct {
	BaseURL   string
	TableName string
	CtxKey    string
}

type orderResponse struct {
	*models.Order
}

func (t *orderResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// pre-processing before sending it out
	return nil
}

type orderRequest struct {
	*models.Order

	// TODO: Add content here
}

func (t *orderRequest) Bind(r *http.Request) error {
	// post-processing after decode

	// TODO: Add post-processing

	return nil
}

func (ur *OrdersAdminResource) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(render.SetContentType(render.ContentTypeJSON))
	// URLFormat is a middleware that parses the url extension from a request path and stores it
	// on the context as a string under the key `middleware.URLFormatCtxKey`. The middleware will
	// trim the suffix from the routing path and continue routing.
	r.Use(middleware.URLFormat)

	r.Get("/", ur.ListView)
	r.Post("/", ur.CreateView)

	r.Route("/{id}", func(in chi.Router) {
		in.Use(ur.EntityCtx)
		in.Get("/", ur.ShowView)
		in.Put("/", ur.EditView)
		in.Delete("/", ur.DeleteView)
	})

	return r
}

func (ar *OrdersAdminResource) EntityCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			order = &models.Order{}
			err   error
			objID = chi.URLParam(r, "id")
			tx    = db.GetDBFromRequestContext(r)
		)
		if objID == "" {
			render.Render(w, r, ErrNotFound)
			return
		}

		err = tx.Model(order).
			Preload("OrderItem").
			Where("id = ?", objID).
			Find(order).Error

		if err == gorm.ErrRecordNotFound {
			render.Render(w, r, ErrNotFound)
			return
		}
		if err != nil {
			fmt.Errorf("%v-Ctx Err: %v", ar.TableName, err)
			e := ErrInternalServerError(err)
			render.Render(w, r, e)
			return
		}

		ctx := context.WithValue(r.Context(), ar.CtxKey, order)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
func (ar *OrdersAdminResource) GetObjectFromCtx(r *http.Request) *models.Order {
	obj, _ := r.Context().Value(ar.CtxKey).(*models.Order)
	return obj
}

func (ar *OrdersAdminResource) ListView(w http.ResponseWriter, r *http.Request) {
	var (
		items = []models.Order{}
		tx    = db.GetDBFromRequestContext(r)
	)
	if err := tx.Table(ar.TableName).Order("created_at").Find(&items).Error; err != nil {
		e := ErrInternalServerError(err)
		render.Render(w, r, e)
		return
	}
	render.JSON(w, r, items)
	return
}

func (ar *OrdersAdminResource) CreateView(w http.ResponseWriter, r *http.Request) {
	tx := db.GetDBFromRequestContext(r)
	data := &orderRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	obj := data.Order
	if err := tx.Model(obj).Create(obj).Error; err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, obj)
}
func (ar *OrdersAdminResource) ShowView(w http.ResponseWriter, r *http.Request) {
	// positive case

	obj := ar.GetObjectFromCtx(r)
	if obj == nil {
		render.Render(w, r, ErrNotFound)
		return
	}

	if err := render.Render(w, r, &orderResponse{obj}); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
	return
}
func (ar *OrdersAdminResource) EditView(w http.ResponseWriter, r *http.Request) {

}
func (ar *OrdersAdminResource) DeleteView(w http.ResponseWriter, r *http.Request) {
	tx := db.GetDBFromRequestContext(r)
	obj := ar.GetObjectFromCtx(r)
	if obj == nil {
		render.Render(w, r, ErrNotFound)
		return
	}
	if err := tx.Table(ar.TableName).Delete(obj).Error; err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

	if err := render.Render(w, r, &orderResponse{obj}); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}
