package controllers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"bitbucket.org/softwarehouseio/victory/victory-frontend/db"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/models"
	"github.com/flosch/pongo2"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"
)

type AdminResource struct {
	BaseURL string
}

func (ar AdminResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(ar.AdminContext)

	//r.Get("/", http.RedirectHandler("dashboard/", http.StatusSeeOther).ServeHTTP)
	r.Get("/", ar.NGAdmin)

	r.Get("/dashboard/", ar.Dashboard)
	r.Get("/orders/", ar.Orders)
	r.Get("/orders/{orderID}/", ar.OrderView)
	r.Get("/products/", ar.Products)
	r.Get("/products/{productID}/", ar.ProductView)

	userAdminRes := &UserAdminResource{BaseURL: ar.BaseURL + "/users/"}
	statsLeague := StatsLeagueAdminResource{BaseURL: "/leagues"}
	statsTeam := StatsTeamAdminResource{BaseURL: "/teams"}
	pARes := &ProductAdminResource{
		BaseURL:   "/products",
		TableName: "products",
		CtxKey:    "prodARCtx",
		ItemQueryInject: func(tx *gorm.DB) *gorm.DB {
			tx.Model(models.Product{}).Preload("Collections")
			return tx
		},
	}
	variations := &ProductVariationAdminResource{
		BaseURL:   "/variations",
		TableName: "product_variations",
		CtxKey:    "variationsCtx",
	}
	badges := &BadgeAdminResource{
		BaseURL:   "/badges",
		TableName: "badges",
		CtxKey:    "badgesCtx",
	}
	sportmonks := &SportmonksAdminResource{
		BaseURL: "/sportmonks",
	}
	orders := &OrdersAdminResource{
		BaseURL:   "/orders",
		TableName: "orders",
		CtxKey:    "ordersCtx",
	}

	r.Mount("/users", userAdminRes.Routes())
	r.Mount(statsLeague.BaseURL, statsLeague.Routes())
	r.Mount(statsTeam.BaseURL, statsTeam.Routes())
	r.Mount(pARes.BaseURL, pARes.Routes())
	r.Mount(variations.BaseURL, variations.Routes())
	r.Mount(badges.BaseURL, badges.Routes())
	r.Mount(sportmonks.BaseURL, sportmonks.Routes())
	r.Mount(orders.BaseURL, orders.Routes())

	emailAdmin := AdminEmail{BaseURL: "/email"}
	r.Mount(emailAdmin.BaseURL, emailAdmin.Routes())

	return r
}

// decided to use metronic admin4
// MARK: Middleware
func (ar AdminResource) AdminContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tplContext := WebResource{}.GetTplContext(r)

		var (
			user, _ = tplContext["user"].(*models.User)
		)
		noUserFound := user == nil

		if noUserFound {
			http.Redirect(w, r, "/user/auth/signin?from=admin", http.StatusSeeOther)
			return
		}
		if user != nil && user.UserAccessLevel < models.DefaultAdminLevel {
			render.Render(w, r, ErrInvalidRequest(fmt.Errorf("Not An Admin")))
			return
		}
		tplContext.Update(pongo2.Context{
			//"continueShoppingURL": continueShoppingURL,
			"BaseURL": ar.BaseURL,
		})
		ctx := context.WithValue(r.Context(), TplContextWebResourceContextKey, tplContext)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// MARK: Views
func (ar *AdminResource) NGAdmin(w http.ResponseWriter, r *http.Request) {
	var (
		tpl        = pongo2.Must(pongo2.FromFile("templates/html/ng-admin.html"))
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
func (ar *AdminResource) Dashboard(w http.ResponseWriter, r *http.Request) {
	//metronic_v3.7/theme/templates/admin4/ecommerce_index.html
	var (
		tpl        = pongo2.Must(pongo2.FromFile("templates/html/admin-dashboard.html"))
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

func (ar *AdminResource) Orders(w http.ResponseWriter, r *http.Request) {
	//	file:///Users/albsen/projects/albsen/themes/metronic_v3.7/theme/templates/admin4/ecommerce_orders.html
	var (
		tpl        = pongo2.Must(pongo2.FromFile("templates/html/admin-orders.html"))
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

func (ar *AdminResource) OrderView(w http.ResponseWriter, r *http.Request) {
	//	file:///Users/albsen/projects/albsen/themes/metronic_v3.7/theme/templates/admin4/ecommerce_orders_view.html
	var (
		tpl        = pongo2.Must(pongo2.FromFile("templates/html/admin-order-view.html"))
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

func (ar *AdminResource) Products(w http.ResponseWriter, r *http.Request) {
	//file:///Users/albsen/projects/albsen/themes/metronic_v3.7/theme/templates/admin4/ecommerce_products.html
	var (
		tpl        = pongo2.Must(pongo2.FromFile("templates/html/admin-products.html"))
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

func (ar *AdminResource) ProductView(w http.ResponseWriter, r *http.Request) {
	//file:///Users/albsen/projects/albsen/themes/metronic_v3.7/theme/templates/admin4/ecommerce_products_edit.html
	var (
		tpl        = pongo2.Must(pongo2.FromFile("templates/html/admin-product-view.html"))
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

// MARK: User Management
func (ar *AdminResource) Users(w http.ResponseWriter, r *http.Request) {
	//	file:///Users/albsen/projects/albsen/themes/metronic_v3.7/theme/templates/admin4/table_managed.html
	/*
	   Use the Managed Table, the big one on top
	*/
	var (
		tpl        = pongo2.Must(pongo2.FromFile("templates/html/admin-users.html"))
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

/*
listView      => GET    /users
creationView  => POST   /users
showView      => GET    /users/:id
editionView   => PUT    /users/:id
deletionView  => DELETE /users/:id
*/

type UserAdminResource struct {
	BaseURL string
}

func (ur *UserAdminResource) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(render.SetContentType(render.ContentTypeJSON))
	// URLFormat is a middleware that parses the url extension from a request path and stores it
	// on the context as a string under the key `middleware.URLFormatCtxKey`. The middleware will
	// trim the suffix from the routing path and continue routing.
	r.Use(middleware.URLFormat)

	r.Get("/", ur.ListView)
	r.Post("/", ur.CreateView)

	r.Route("/{id}", func(in chi.Router) {
		in.Use(UserAdminCtx)
		in.Get("/", ur.ShowView)
		in.Put("/", ur.EditView)
		in.Delete("/", ur.DeleteView)
	})
	return r
}

// MARK: middleware
const UserAdminResourceContextKey = "userAdminCtx"

func UserAdminCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			user    *models.User
			err     error
			userUID = chi.URLParam(r, "id")
			tx      = db.GetDBFromRequestContext(r)
		)

		if userUID == "" {
			render.Render(w, r, ErrNotFound)
			return
		}

		user, err = models.User{}.GetUser(tx, userUID)

		if err != nil {
			e := ErrInternalServerError(err)
			render.Render(w, r, e)
			return
		}

		ctx := context.WithValue(r.Context(), UserAdminResourceContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
func GetUserAdminCtx(r *http.Request) *models.User {
	user, _ := r.Context().Value(UserAdminResourceContextKey).(*models.User)
	return user
}

// MARK: Views
// listView      => GET    /users
func (ur *UserAdminResource) ListView(w http.ResponseWriter, r *http.Request) {
	var (
		users = []models.User{}
		tx    = db.GetDBFromRequestContext(r)
	)

	if err := tx.Model(users).Order("created_at").Find(&users).Error; err != nil {
		e := ErrInternalServerError(err)
		render.Render(w, r, e)
		return
	}

	render.JSON(w, r, users)
	return
}

// creationView  => POST   /users
func (ur *UserAdminResource) CreateView(w http.ResponseWriter, r *http.Request) {
	var (
		tx = db.GetDBFromRequestContext(r)
	)
	data := &UserAdminRequest{}
	// serialize the payload
	if err := render.Bind(r, data); err != nil {
		fmt.Errorf("bind failed %v", err)
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	user := data.User

	if err := tx.Create(user).Error; err != nil {
		fmt.Errorf("create failed %v", err)
		e := ErrInternalServerError(err)
		render.Render(w, r, e)
		return
	}
	render.Status(r, http.StatusCreated)
	render.Render(w, r, NewUserAdminResponse(user))
}

// showView      => GET    /users/:id
func (ur *UserAdminResource) ShowView(w http.ResponseWriter, r *http.Request) {
	var (
		user = GetUserAdminCtx(r)
	)
	if user == nil {
		render.Render(w, r, ErrNotFound)
		return
	}

	if err := render.Render(w, r, NewUserAdminResponse(user)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// editionView   => PUT    /users/:id
func (ur *UserAdminResource) EditView(w http.ResponseWriter, r *http.Request) {
	var (
		tx   = db.GetDBFromRequestContext(r)
		user = GetUserAdminCtx(r)
	)
	if user == nil {
		render.Render(w, r, ErrNotFound)
		return
	}

	data := &UserAdminRequest{User: user}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	user = data.User
	fmt.Println("triggering update:")
	if err := tx.Model(user).Select("user_access_level", "email").Updates(user).Error; err != nil {
		fmt.Errorf("Failed to update user: %v", err)
		e := ErrInternalServerError(err)
		render.Render(w, r, e)
		return
	}
	fmt.Println("update triggered")

	if err := render.Render(w, r, NewUserAdminResponse(user)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// deletionView  => DELETE /users/:id
func (ur *UserAdminResource) DeleteView(w http.ResponseWriter, r *http.Request) {
	render.Render(w, r, ErrNotImplemented)
	return
}

// UserAdminRequest - the request payload for User data model
// Note: it's good practice to have well defined req / resp payloads
type UserAdminRequest struct {
	*models.User

	ProtectedID    string `json:"id"` // override 'id' json to have more control
	ProtectedEmail string `json:"email"`
}

func (uar *UserAdminRequest) Bind(r *http.Request) error {
	// post-processing after decode
	uar.ProtectedID = ""
	uar.ProtectedEmail = strings.ToLower(uar.ProtectedEmail)
	return nil
}

type UserAdminResponse struct {
	*models.User
}

func NewUserAdminResponse(user *models.User) *UserAdminResponse {
	resp := &UserAdminResponse{User: user}
	return resp
}
func (rd *UserAdminResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// pre-processing before a response is marshalled and sent across the wire
	return nil
}
