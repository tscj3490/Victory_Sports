package controllers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"

	"bitbucket.org/softwarehouseio/victory/victory-frontend/config"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/config/i18n/l10n"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/db"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/models"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"
	"github.com/kjk/betterguid"
)

type ProductAdminResource struct {
	BaseURL         string
	TableName       string
	CtxKey          string
	ItemQueryInject func(tx *gorm.DB) *gorm.DB
}
type resourceModel models.Product

type resourceResponse struct {
	*resourceModel
	TeamL10N map[string]interface{} `json:"Team"`
}

func (t *resourceResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// pre-processing before sending it out
	var teamTmp models.Team
	teamTmp = *t.Team
	t.TeamL10N = l10n.Unpack(l10n.L10NModel(teamTmp))
	return nil
}

type resourceRequest struct {
	*models.Product
	TeamTMP        map[string]interface{} `json:"Team"`
	BrandTmp       string                 `json:"Brand"`
	CollectionsTmp []string               `json:"Collections"`
}

func (t *resourceRequest) Bind(r *http.Request) error {
	// post-processing after decode
	tx := db.GetDBFromRequestContext(r)

	brand := models.Brand{}

	if err := tx.Find(&brand, "name = ?", t.BrandTmp).Error; err != nil {
		log.Printf("AP.CreateResource.Bind cant find brands. %v", err)
	}
	t.Product.Brand = &brand

	collections := []models.Collection{}
	if err := tx.Model(models.Collection{}).
		Where("code in (?)", t.CollectionsTmp).
		Find(&collections).Error; err != nil {

		log.Printf("AP.CreateResource.Bind cant find collections. %v", err)
	}
	t.Product.Collections = collections
	return nil
}
func (t *resourceRequest) Get() *models.Product {
	return t.Product
}

type createResourceRequest struct {
	*models.Product
	SizesTmp          []string `json:"Sizes"`
	BrandTmp          string   `json:"Brand"`
	CollectionsTmp    []string `json:"Collections"`
	AvailableQuantity uint
}

func (t *createResourceRequest) Bind(r *http.Request) error {
	// post-processing after decode
	// the create request sends the combined product and product variation together

	tx := db.GetDBFromRequestContext(r)
	sizes := []models.ProductSize{}
	if err := tx.Model(sizes).Where("name in (?)", t.SizesTmp).Find(&sizes).Error; err != nil {
		log.Printf("AP.CreateResource.Bind Failed to fetch sizes: %v", err)
	}
	t.Product.Sizes = sizes

	brand := models.Brand{}

	if err := tx.Find(&brand, "name = ?", t.BrandTmp).Error; err != nil {
		log.Printf("AP.CreateResource.Bind cant find brands. %v", err)
	}
	t.Product.Brand = &brand

	collections := []models.Collection{}
	if err := tx.Model(models.Collection{}).
		Where("code in (?)", t.CollectionsTmp).
		Find(&collections).Error; err != nil {

		log.Printf("AP.CreateResource.Bind cant find collections. %v", err)
	}
	t.Product.Collections = collections

	for _, size := range sizes {
		variation := models.ProductVariation{
			Size:              size,
			SKU:               fmt.Sprintf("%v-%v", t.Name, size.Name),
			AvailableQuantity: t.AvailableQuantity,
		}
		t.Variations = append(t.Variations, variation)
	}

	log.Printf("Got %v", t.Product)

	return nil
}
func (t *createResourceRequest) Get() *models.Product {
	return t.Product
}

func (ur *ProductAdminResource) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(render.SetContentType(render.ContentTypeJSON))
	// URLFormat is a middleware that parses the url extension from a request path and stores it
	// on the context as a string under the key `middleware.URLFormatCtxKey`. The middleware will
	// trim the suffix from the routing path and continue routing.
	r.Use(middleware.URLFormat)

	r.Get("/", ur.ListView)
	r.Post("/", ur.CreateView)
	r.Post("/image-upload", ur.ImageUpload)

	r.Route("/{id}", func(in chi.Router) {
		in.Use(ur.EntityCtx)
		in.Get("/", ur.ShowView)
		in.Put("/", ur.EditView)
		in.Delete("/", ur.DeleteView)
	})
	return r
}
func (ar *ProductAdminResource) EntityCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			item  = &resourceModel{}
			err   error
			objID = chi.URLParam(r, "id")
			tx    = db.GetDBFromRequestContext(r)
		)
		if objID == "" {
			render.Render(w, r, ErrNotFound)
			return
		}

		//err = ar.ItemQueryInject(tx.Table(ar.TableName).Where("id = ?", objID)).Find(item).Error
		product := &models.Product{}
		err = tx.Model(product).
			Preload("Variations").
			Preload("Variations.Size").
			Preload("Variations.Badge").
			Preload("Collections").
			Preload("Sizes").
			Preload("Category").
			Preload("Team").
			Preload("Brand").
			Where("id = ?", objID).Find(product).Error
		item = (*resourceModel)(product)

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
func (ar *ProductAdminResource) GetObjectFromCtx(r *http.Request) *resourceModel {
	obj, _ := r.Context().Value(ar.CtxKey).(*resourceModel)
	return obj
}
func (ar *ProductAdminResource) ListView(w http.ResponseWriter, r *http.Request) {
	var (
		items = []resourceModel{}
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

func (ar *ProductAdminResource) CreateView(w http.ResponseWriter, r *http.Request) {
	tx := db.GetDBFromRequestContext(r)
	data := &createResourceRequest{}
	if err := render.Bind(r, data); err != nil {
		e := ErrInternalServerError(err)
		render.Render(w, r, e)
		return
	}

	obj := data.Get()
	if err := tx.Model(obj).Create(obj).Error; err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, obj)
}
func (ar *ProductAdminResource) ShowView(w http.ResponseWriter, r *http.Request) {
	// positive case

	obj := ar.GetObjectFromCtx(r)
	if obj == nil {
		render.Render(w, r, ErrNotFound)
		return
	}

	if err := render.Render(w, r, &resourceResponse{obj, nil}); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
	return
}
func (ar *ProductAdminResource) EditView(w http.ResponseWriter, r *http.Request) {

	tx := db.GetDBFromRequestContext(r)
	obj := ar.GetObjectFromCtx(r)
	if obj == nil {
		render.Render(w, r, ErrNotFound)
		return
	}

	data := &resourceRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	objNew := data.Get()

	if err := tx.Model(objNew).Where("id = ?", obj.ID).Update(objNew).Error; err != nil {
		log.Printf("PAR.EditView failed: %v", err)
		render.Render(w, r, ErrRender(err))
		return
	}
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, objNew)
}
func (ar *ProductAdminResource) DeleteView(w http.ResponseWriter, r *http.Request) {
	tx := db.GetDBFromRequestContext(r)
	obj := ar.GetObjectFromCtx(r)
	if obj == nil {
		render.Render(w, r, ErrNotFound)
		return
	}
	// delete product variations
	prodVariations := []models.ProductVariation{}
	if err := tx.Model(prodVariations).Where("product_id = ?", obj.ID).Delete(&prodVariations).Error; err != nil {
		log.Printf("ProdAdminR.Delete failed: %v", err)
		render.Render(w, r, ErrRender(err))
		return
	}
	// delete actual product
	if err := tx.Table(ar.TableName).Delete(obj).Error; err != nil {
		log.Printf("ProdAdminR.Delete failed: %v", err)
		render.Render(w, r, ErrRender(err))
		return
	}
	obj.Variations = prodVariations

	if err := render.Render(w, r, &resourceResponse{obj, nil}); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

func (ar *ProductAdminResource) ImageUpload(w http.ResponseWriter, r *http.Request) {

	log.Printf("PAR.ImageUpload - ")
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

	apifilename := ""

	for _, fheaders := range r.MultipartForm.File {
		for _, handler := range fheaders {

			infile, err := handler.Open()
			if err != nil {
				log.Fatal(err)
				render.Render(w, r,
					ErrRender(
						fmt.Errorf("open directory to save file reason: %v", err)))
				return
			}
			fileName := fmt.Sprintf("product_%v", betterguid.New())
			apifilename = fileName
			completePath := path.Join(config.ENVFileUploadDir(), fileName)
			outfile, err := os.OpenFile(completePath, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				log.Fatal(err)
				render.Render(w, r, ErrRender(err))
				return
			}
			// 32K buffer copy
			if _, err = io.Copy(outfile, infile); nil != err {
				log.Fatal(err)
				render.Render(w, r,
					ErrRender(
						fmt.Errorf("storing file failed reason: %v", err)))
				return
			}
		}
	}

	fileURI := fmt.Sprintf("%v%v", config.ENVFileUploadURI, apifilename)

	render.JSON(w, r, map[string]interface{}{"success": true, "image_name": fileURI})
	return
}
