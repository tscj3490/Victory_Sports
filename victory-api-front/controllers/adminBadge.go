package controllers

import (
	"github.com/jinzhu/gorm"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/models"
	"net/http"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/db"
	"context"
	"fmt"
	"github.com/go-chi/chi/middleware"
)

type BadgeAdminResource struct {
	BaseURL string
	TableName string
	CtxKey string
}

type badgeResponse struct {
	*models.Badge
}
func (t *badgeResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// pre-processing before sending it out
	return nil
}

type badgeRequest struct {
	*models.Badge
}
func (t *badgeRequest) Bind(r *http.Request) error {
	// post-processing after decode
	return nil
}

func (ur *BadgeAdminResource) Routes() chi.Router {
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


func (ar *BadgeAdminResource) EntityCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			item = &models.Badge{}
			err   error
			objID = chi.URLParam(r, "id")
			tx = db.GetDBFromRequestContext(r)
		)
		if objID == "" {
			render.Render(w, r, ErrNotFound)
			return
		}

		err = tx.Model(item).
			Where("id = ?", objID).
			Find(item).Error

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

		ctx := context.WithValue(r.Context(), ar.CtxKey, item)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
func (ar *BadgeAdminResource) GetObjectFromCtx(r *http.Request) *models.Badge{
	obj, _ := r.Context().Value(ar.CtxKey).(*models.Badge)
	return obj
}

func (ar *BadgeAdminResource) ListView(w http.ResponseWriter, r *http.Request) {
	var (
		items = []models.Badge{}
		tx = db.GetDBFromRequestContext(r)
	)
	if err := tx.Table(ar.TableName).Order("created_at").Find(&items).Error; err != nil {
		e := ErrInternalServerError(err)
		render.Render(w, r, e)
		return
	}
	render.JSON(w, r, items)
	return
}

func (ar *BadgeAdminResource) CreateView(w http.ResponseWriter, r *http.Request) {
	tx := db.GetDBFromRequestContext(r)
	data := &badgeRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	obj := data
	if err := tx.Model(obj).Create(obj).Error; err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, obj)
}
func (ar *BadgeAdminResource) ShowView(w http.ResponseWriter, r *http.Request) {
	// positive case

	obj := ar.GetObjectFromCtx(r);
	if obj == nil {
		render.Render(w, r, ErrNotFound)
		return
	}

	if err := render.Render(w, r, &badgeResponse{obj}); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
	return
}
func (ar *BadgeAdminResource) EditView(w http.ResponseWriter, r *http.Request) {

}
func (ar *BadgeAdminResource) DeleteView(w http.ResponseWriter, r *http.Request) {
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

	if err := render.Render(w, r, &badgeResponse{obj, }); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}
