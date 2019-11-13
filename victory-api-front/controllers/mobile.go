package controllers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"bitbucket.org/softwarehouseio/victory/victory-frontend/models"
	"github.com/flosch/pongo2"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type MobileResource struct {
	BaseURL string
}

func (ar MobileResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(ar.MobileContext)

	//r.Get("/", http.RedirectHandler("dashboard/", http.StatusSeeOther).ServeHTTP)
	r.Get("/", ar.NGMobile)
	//mobileTeam := StatsTeamMobileResource{BaseURL: "/Teams"}
	//mobileLeague := StatsLeagueMobileResource{BaseURL: "/Leagues"}
	//r.Mount(mobileTeam.BaseURL, mobileTeam.Routes())
	//r.Mount(mobileLeague.BaseURL, mobileLeague.Routes())
	// r.Get("/dashboard/", ar.Dashboard)
	// r.Get("/orders/", ar.Orders)
	// r.Get("/orders/{orderID}/", ar.OrderView)
	// r.Get("/products/", ar.Products)
	// r.Get("/products/{productID}/", ar.ProductView)

	// userMobileRes := &UserMobileResource{BaseURL: ar.BaseURL + "/users/"}
	// statsLeague := StatsLeagueMobileResource{BaseURL: "/leagues"}

	// pARes := &ProductMobileResource{
	// 	BaseURL:   "/products",
	// 	TableName: "products",
	// 	CtxKey:    "prodARCtx",
	// 	ItemQueryInject: func(tx *gorm.DB) *gorm.DB {
	// 		tx.Model(models.Product{}).Preload("Collections")
	// 		return tx
	// 	},
	// }
	// variations := &ProductVariationMobileResource{
	// 	BaseURL:   "/variations",
	// 	TableName: "product_variations",
	// 	CtxKey:    "variationsCtx",
	// }
	// badges := &BadgeMobileResource{
	// 	BaseURL:   "/badges",
	// 	TableName: "badges",
	// 	CtxKey:    "badgesCtx",
	// }

	// r.Mount("/users", userMobileRes.Routes())
	// r.Mount(statsLeague.BaseURL, statsLeague.Routes())

	// r.Mount(pARes.BaseURL, pARes.Routes())
	// r.Mount(variations.BaseURL, variations.Routes())
	// r.Mount(badges.BaseURL, badges.Routes())

	// emailMobile := MobileEmail{BaseURL: "/email"}
	// r.Mount(emailMobile.BaseURL, emailMobile.Routes())

	return r
}

// decided to use metronic Mobile4
// MARK: Middleware
func (ar MobileResource) MobileContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tplContext := WebResource{}.GetTplContext(r)

		// var (
		// 	user, _ = tplContext["user"].(*models.User)
		// )
		// noUserFound := user == nil

		// if noUserFound {
		// 	http.Redirect(w, r, "/user/auth/signin?from=Mobile", http.StatusSeeOther)
		// 	return
		// }
		// if user != nil && user.UserAccessLevel < models.DefaultAdminLevel {
		// 	render.Render(w, r, ErrInvalidRequest(fmt.Errorf("Not An Admin")))
		// 	return
		// }
		tplContext.Update(pongo2.Context{
			//"continueShoppingURL": continueShoppingURL,
			"BaseURL": ar.BaseURL,
		})
		ctx := context.WithValue(r.Context(), TplContextWebResourceContextKey, tplContext)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// MARK: Views
func (ar *MobileResource) NGMobile(w http.ResponseWriter, r *http.Request) {
	var (
		tpl        = pongo2.Must(pongo2.FromFile("templates/html/ng-Mobile.html"))
		tplContext = WebResource{}.GetTplContext(r)
	)

	tplContext.Update(pongo2.Context{
		//"continueShoppingURL": continueShoppingURL,
		"lalala": fmt.Sprintf("%v%v", ar.BaseURL, "checkout/"),
	})

	if err := tpl.ExecuteWriter(tplContext, w); err != nil {
		fmt.Printf("Tpl.ExecuteWriter Failed %v", err)
		e := ErrInternalServerError(err)
		render.Render(w, r, e)
		return
	}
}

// MARK: E-Commerce
func (ar *MobileResource) Dashboard(w http.ResponseWriter, r *http.Request) {
	// //metronic_v3.7/theme/templates/Mobile4/ecommerce_index.html
	// var (
	// 	tpl        = pongo2.Must(pongo2.FromFile("templates/html/Mobile-dashboard.html"))
	// 	tplContext = WebResource{}.GetTplContext(r)
	// )

	// tplContext.Update(pongo2.Context{
	// 	//"continueShoppingURL": continueShoppingURL,
	// 	"lalala": fmt.Sprintf("%v%v", ar.BaseURL, "checkout/"),
	// })

	// if err := tpl.ExecuteWriter(tplContext, w); err != nil {
	// 	fmt.Println("Tpl.ExecuteWriter Failed %v", err)
	// 	render.Render(w, r, ErrInternalServerError)
	// 	return
	// }
}

func (ar *MobileResource) Orders(w http.ResponseWriter, r *http.Request) {
	// //	file:///Users/albsen/projects/albsen/themes/metronic_v3.7/theme/templates/Mobile4/ecommerce_orders.html
	// var (
	// 	tpl        = pongo2.Must(pongo2.FromFile("templates/html/Mobile-orders.html"))
	// 	tplContext = WebResource{}.GetTplContext(r)
	// )

	// tplContext.Update(pongo2.Context{
	// 	//"continueShoppingURL": continueShoppingURL,
	// 	"lalala": fmt.Sprintf("%v%v", ar.BaseURL, "checkout/"),
	// })

	// if err := tpl.ExecuteWriter(tplContext, w); err != nil {
	// 	fmt.Println("Tpl.ExecuteWriter Failed %v", err)
	// 	render.Render(w, r, ErrInternalServerError)
	// 	return
	// }
}

func (ar *MobileResource) OrderView(w http.ResponseWriter, r *http.Request) {
	// //	file:///Users/albsen/projects/albsen/themes/metronic_v3.7/theme/templates/Mobile4/ecommerce_orders_view.html
	// var (
	// 	tpl        = pongo2.Must(pongo2.FromFile("templates/html/Mobile-order-view.html"))
	// 	tplContext = WebResource{}.GetTplContext(r)
	// )

	// tplContext.Update(pongo2.Context{
	// 	//"continueShoppingURL": continueShoppingURL,
	// 	"lalala": fmt.Sprintf("%v%v", ar.BaseURL, "checkout/"),
	// })

	// if err := tpl.ExecuteWriter(tplContext, w); err != nil {
	// 	fmt.Println("Tpl.ExecuteWriter Failed %v", err)
	// 	render.Render(w, r, ErrInternalServerError)
	// 	return
	// }
}

func (ar *MobileResource) Products(w http.ResponseWriter, r *http.Request) {
	// //file:///Users/albsen/projects/albsen/themes/metronic_v3.7/theme/templates/Mobile4/ecommerce_products.html
	// var (
	// 	tpl        = pongo2.Must(pongo2.FromFile("templates/html/Mobile-products.html"))
	// 	tplContext = WebResource{}.GetTplContext(r)
	// )

	// tplContext.Update(pongo2.Context{
	// 	//"continueShoppingURL": continueShoppingURL,
	// 	"lalala": fmt.Sprintf("%v%v", ar.BaseURL, "checkout/"),
	// })

	// if err := tpl.ExecuteWriter(tplContext, w); err != nil {
	// 	fmt.Println("Tpl.ExecuteWriter Failed %v", err)
	// 	render.Render(w, r, ErrInternalServerError)
	// 	return
	// }
}

func (ar *MobileResource) ProductView(w http.ResponseWriter, r *http.Request) {
	// //file:///Users/albsen/projects/albsen/themes/metronic_v3.7/theme/templates/Mobile4/ecommerce_products_edit.html
	// var (
	// 	tpl        = pongo2.Must(pongo2.FromFile("templates/html/Mobile-product-view.html"))
	// 	tplContext = WebResource{}.GetTplContext(r)
	// )

	// tplContext.Update(pongo2.Context{
	// 	//"continueShoppingURL": continueShoppingURL,
	// 	"lalala": fmt.Sprintf("%v%v", ar.BaseURL, "checkout/"),
	// })

	// if err := tpl.ExecuteWriter(tplContext, w); err != nil {
	// 	fmt.Println("Tpl.ExecuteWriter Failed %v", err)
	// 	render.Render(w, r, ErrInternalServerError)
	// 	return
	// }
}

// MARK: User Management
func (ar *MobileResource) Users(w http.ResponseWriter, r *http.Request) {
	// //	file:///Users/albsen/projects/albsen/themes/metronic_v3.7/theme/templates/Mobile4/table_managed.html
	// /*
	//    Use the Managed Table, the big one on top
	// */
	// var (
	// 	tpl        = pongo2.Must(pongo2.FromFile("templates/html/Mobile-users.html"))
	// 	tplContext = WebResource{}.GetTplContext(r)
	// )

	// tplContext.Update(pongo2.Context{
	// 	//"continueShoppingURL": continueShoppingURL,
	// 	"lalala": fmt.Sprintf("%v%v", ar.BaseURL, "checkout/"),
	// })

	// if err := tpl.ExecuteWriter(tplContext, w); err != nil {
	// 	fmt.Println("Tpl.ExecuteWriter Failed %v", err)
	// 	render.Render(w, r, ErrInternalServerError)
	// 	return
	// }
}

/*
listView      => GET    /users
creationView  => POST   /users
showView      => GET    /users/:id
editionView   => PUT    /users/:id
deletionView  => DELETE /users/:id
*/

type UserMobileResource struct {
	BaseURL string
}

func (ur *UserMobileResource) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(render.SetContentType(render.ContentTypeJSON))
	// URLFormat is a middleware that parses the url extension from a request path and stores it
	// on the context as a string under the key `middleware.URLFormatCtxKey`. The middleware will
	// trim the suffix from the routing path and continue routing.
	r.Use(middleware.URLFormat)

	r.Get("/", ur.ListView)
	r.Post("/", ur.CreateView)

	r.Route("/{id}", func(in chi.Router) {
		in.Use(UserMobileCtx)
		in.Get("/", ur.ShowView)
		in.Put("/", ur.EditView)
		in.Delete("/", ur.DeleteView)
	})
	return r
}

// MARK: middleware
const UserMobileResourceContextKey = "userMobileCtx"

func UserMobileCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// var (
		// 	user    *models.User
		// 	err     error
		// 	userUID = chi.URLParam(r, "id")
		// 	tx      = db.GetDBFromRequestContext(r)
		// )

		// if userUID == "" {
		// 	render.Render(w, r, ErrNotFound)
		// 	return
		// }

		// user, err = models.User{}.GetUser(tx, userUID)

		// if err != nil {
		// 	render.Render(w, r, ErrInternalServerError)
		// 	return
		// }

		// ctx := context.WithValue(r.Context(), UserMobileResourceContextKey, user)
		// next.ServeHTTP(w, r.WithContext(ctx))
	})
}
func GetUserMobileCtx(r *http.Request) *models.User {
	user, _ := r.Context().Value(UserMobileResourceContextKey).(*models.User)
	return user
}

// MARK: Views
// listView      => GET    /users
func (ur *UserMobileResource) ListView(w http.ResponseWriter, r *http.Request) {
	// var (
	// 	users = []models.User{}
	// 	tx    = db.GetDBFromRequestContext(r)
	// )

	// if err := tx.Model(users).Order("created_at").Find(&users).Error; err != nil {
	// 	render.Render(w, r, ErrInternalServerError)
	// 	return
	// }

	// render.JSON(w, r, users)
	// return
}

// creationView  => POST   /users
func (ur *UserMobileResource) CreateView(w http.ResponseWriter, r *http.Request) {
	// var (
	// 	tx = db.GetDBFromRequestContext(r)
	// )
	// data := &UserMobileRequest{}
	// // serialize the payload
	// if err := render.Bind(r, data); err != nil {
	// 	fmt.Errorf("bind failed %v", err)
	// 	render.Render(w, r, ErrInvalidRequest(err))
	// 	return
	// }
	// user := data.User

	// if err := tx.Create(user).Error; err != nil {
	// 	fmt.Errorf("create failed %v", err)
	// 	render.Render(w, r, ErrInternalServerError)
	// 	return
	// }
	// render.Status(r, http.StatusCreated)
	// render.Render(w, r, NewUserMobileResponse(user))
}

// showView      => GET    /users/:id
func (ur *UserMobileResource) ShowView(w http.ResponseWriter, r *http.Request) {
	// var (
	// 	user = GetUserMobileCtx(r)
	// )
	// if user == nil {
	// 	render.Render(w, r, ErrNotFound)
	// 	return
	// }

	// if err := render.Render(w, r, NewUserMobileResponse(user)); err != nil {
	// 	render.Render(w, r, ErrRender(err))
	// 	return
	// }
}

// editionView   => PUT    /users/:id
func (ur *UserMobileResource) EditView(w http.ResponseWriter, r *http.Request) {
	// var (
	// 	tx   = db.GetDBFromRequestContext(r)
	// 	user = GetUserMobileCtx(r)
	// )
	// if user == nil {
	// 	render.Render(w, r, ErrNotFound)
	// 	return
	// }

	// data := &UserMobileRequest{User: user}
	// if err := render.Bind(r, data); err != nil {
	// 	render.Render(w, r, ErrInvalidRequest(err))
	// 	return
	// }
	// user = data.User
	// fmt.Println("triggering update:")
	// if err := tx.Model(user).Select("user_access_level", "email").Updates(user).Error; err != nil {
	// 	fmt.Errorf("Failed to update user: %v", err)
	// 	render.Render(w, r, ErrInternalServerError)
	// 	return
	// }
	// fmt.Println("update triggered")

	// if err := render.Render(w, r, NewUserMobileResponse(user)); err != nil {
	// 	render.Render(w, r, ErrRender(err))
	// 	return
	// }
}

// deletionView  => DELETE /users/:id
func (ur *UserMobileResource) DeleteView(w http.ResponseWriter, r *http.Request) {
	// render.Render(w, r, ErrNotImplemented)
	// return
}

// UserMobileRequest - the request payload for User data model
// Note: it's good practice to have well defined req / resp payloads
type UserMobileRequest struct {
	*models.User

	ProtectedID    string `json:"id"` // override 'id' json to have more control
	ProtectedEmail string `json:"email"`
}

func (uar *UserMobileRequest) Bind(r *http.Request) error {
	// post-processing after decode
	uar.ProtectedID = ""
	uar.ProtectedEmail = strings.ToLower(uar.ProtectedEmail)
	return nil
}

type UserMobileResponse struct {
	*models.User
}

func NewUserMobileResponse(user *models.User) *UserMobileResponse {
	resp := &UserMobileResponse{User: user}
	return resp
}
func (rd *UserMobileResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// pre-processing before a response is marshalled and sent across the wire
	return nil
}
