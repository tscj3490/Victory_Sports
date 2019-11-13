package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"bitbucket.org/softwarehouseio/victory/victory-frontend/config"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/config/cart"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/config/i18n"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/db"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/models"
	"github.com/alexedwards/scs"
	"github.com/flosch/pongo2"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/gorilla/schema"
)

const KCartSessionKey = "cartSession"

var (
	formDecoder    *schema.Decoder
	sessionManager = scs.NewCookieManager(config.Config.SessionManagerKey)
)

func init() {
	log.Printf("init function cartController run ...")
	formDecoder = schema.NewDecoder()
	formDecoder.IgnoreUnknownKeys(true)
}

type CartData map[string]string

type CartResource struct {
	BaseURL string
}

func (c CartResource) ContextKey() string {
	return "cartCtx"
}

func (c CartResource) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(c.CartSessionContext)
	r.Get("/", c.View)
	r.Post("/add", c.Add)
	r.Post("/controls", c.Controls)

	// checkout needs access to the cart
	checkout := CheckoutResource{BaseURL: "/checkout/"}
	r.Mount(checkout.BaseURL, checkout.Routes())

	return r
}

// cart middleware
func (c CartResource) CartSessionContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Cart.SessionContext %v\n", r.URL.Path)

		session := sessionManager.Load(r)

		// try getting the session
		sessionKeyExists, _ := session.Exists(KCartSessionKey)
		sessionData := CartData{"state": "created"}

		if !sessionKeyExists {
			// make a new session, but there is nothing to store yet ...
			session.PutObject(w, KCartSessionKey, sessionData)
		} else {
			if err := session.GetObject(KCartSessionKey, &sessionData); err != nil {
				// we failed to get the session, lets remove it and make a new one
				session.PutObject(w, KCartSessionKey, sessionData)
			}
		}

		ctx := context.WithValue(r.Context(), KCartSessionKey, sessionData)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
func (c CartResource) GetCartSession(r *http.Request) *CartData {
	dst := CartData{}
	session := sessionManager.Load(r)
	_ = session.GetObject(KCartSessionKey, &dst)
	return &dst
}

// cart method handlers
func (c CartResource) View(w http.ResponseWriter, r *http.Request) {
	// view the cart
	var (
		tpl                 = pongo2.Must(pongo2.FromFile("templates/html/cart.html"))
		tplContext          = WebResource{}.GetTplContext(r)
		continueShoppingURL = "/shop/any/"
		tx                  = db.GetDBFromRequestContext(r)
	)

	shoppingCart, err := cart.GetCart(w, sessionManager.Load(r))
	if err != nil {
		fmt.Printf("failed to retrieve cart reason: %+v", shoppingCart)
	}
	cartItemIDs := shoppingCart.GetItemsIDS()

	selectedProductVariations := []*models.ProductVariation{}
	if err := tx.Model(models.ProductVariation{}).
		Where("id in (?)", cartItemIDs).
		Preload("Product").
		Preload("Size").
		Preload("Badge").
		Find(&selectedProductVariations).Error; err != nil {
		fmt.Printf("failed to fetch product variations based on cart reason: %v", err)
	}

	// calculate subtotal
	subtotal := CalculateSubtotal(selectedProductVariations, shoppingCart)

	getProductVariation := func(variationId uint) *models.ProductVariation {
		for _, pv := range selectedProductVariations {
			if pv.ID == variationId {
				return pv
			}
		}
		return nil
	}
	cartItems := shoppingCart.GetContent()

	tplContext.Update(pongo2.Context{
		"cartIsEmpty":               shoppingCart.IsEmpty(),
		"continueShoppingURL":       continueShoppingURL,
		"checkoutURL":               fmt.Sprintf("%v%v", c.BaseURL, "checkout/"),
		"selectedProductVariations": selectedProductVariations,
		"cartItems":                 cartItems,
		"subtotal":                  subtotal,
		"getProductVariation":       getProductVariation,
	})

	if err := tpl.ExecuteWriter(tplContext, w); err != nil {
		fmt.Printf("Tpl.ExecuteWriter Failed %v", err)
		e := ErrInternalServerError(err)
		render.Render(w, r, e)
		return
	}
}

type CartAddRequest struct {
	ItemID          uint
	ProductSizeIDs  []uint  `schema:"ProductSizes"`
	BadgeIDs        []uint  `schema:"Badges,omitempty"`
	CustomizeName   *string `schema:",omitempty"`
	CustomizeNumber *uint   `schema:",omitempty"`
}

func (cartData *CartAddRequest) BindForm(r *http.Request) error {
	// unfortunately - chi doesn't support ContentTypeForm yet
	return formDecoder.Decode(cartData, r.PostForm)
}
func (c *CartAddRequest) Validate() (bool, []error) {
	errs := []error{}

	if c.ItemID <= 0 {
		errs = append(errs, fmt.Errorf("ItemID has to be above 0"))
	}
	if len(c.ProductSizeIDs) == 0 {
		errs = append(errs, fmt.Errorf("ProductSizes has to have at least one valid integer item."))
	}
	if c.CustomizeNumber != nil && *c.CustomizeNumber > uint(100) {
		errs = append(errs, fmt.Errorf("CustomizeNumber has to be below 100."))
	}
	if c.CustomizeName != nil && len(*c.CustomizeName) > 50 {
		errs = append(errs, fmt.Errorf("CustomizeName can't be longer than 50 characters."))
	}

	isValid := len(errs) == 0
	return isValid, errs
}

func (c CartResource) Add(w http.ResponseWriter, r *http.Request) {
	// add items to the cart
	var (
		tx     = db.GetDBFromRequestContext(r)
		locale = i18n.GetLocaleContext(r)
	)
	localePrepath := ""
	if locale == "ar-AE" {
		localePrepath = "/ar"
	}

	// 1) receive the order
	r.ParseForm()

	cartAdd := &CartAddRequest{}
	if err := cartAdd.BindForm(r); err != nil {
		log.Printf("CartResource.Add failed to convert itemID: %v\n", err)
		errMsg := fmt.Errorf("invalid data in post request - bind failed")
		http.Redirect(w, r, fmt.Sprintf("%v%v#msg=%v", localePrepath, "/cart/", errMsg), http.StatusSeeOther)
		return
	}
	if isValid, errs := cartAdd.Validate(); !isValid {
		log.Printf("CartResource.Add failed to convert validate incoming data: %v\n", errs)
		errMsg := fmt.Errorf("invalid data in post request")
		http.Redirect(w, r, fmt.Sprintf("%v%v#msg=%v", localePrepath, "/cart/", errMsg), http.StatusSeeOther)
		return
	}

	product := models.Product{}
	if err := tx.Model(product).Where("id = ?", cartAdd.ItemID).First(&product).Error; err != nil {
		log.Printf("CartResource.Add failed to get product: %v", err)
		e := ErrInternalServerError(err)
		render.Render(w, r, e)
		return
	}

	productSizeIDs := cartAdd.ProductSizeIDs
	badgeIDs := cartAdd.BadgeIDs
	gotBadges := len(badgeIDs) > 0
	customizeName := cartAdd.CustomizeName
	customizeNumber := cartAdd.CustomizeNumber

	requiresCustomPrint := customizeName != nil || customizeNumber != nil

	availableVariations := []models.ProductVariation{}

	fmt.Printf("received postform: %v %v %#v\n", r.PostForm, len(productSizeIDs), productSizeIDs)

	q := tx.Model(models.ProductVariation{}).Where("product_id = ? AND size_id in (?) AND available_quantity > 0", product.ID, productSizeIDs)
	fmt.Printf("got Badges: %v\n", gotBadges)
	if gotBadges {
		q = q.Where("badge_id in (?)", badgeIDs)
	} else {
		q = q.Where("badge_id is null")
	}
	if requiresCustomPrint {
		q = q.Where("custom_print = 1")
	}

	// based on the badges and sizes try to find the ProductVariation that covers both
	// we will try to fulfil all sizes with the requested badge

	if err := q.Find(&availableVariations).Error; err != nil {
		fmt.Errorf("CartResource.Add failed to get productVariations: %v\n", err)
		errMsg := "The size and/or badge combination you selected was not available."
		http.Redirect(w, r, fmt.Sprintf("%v%v#%v", localePrepath, product.GetRelativeURL(), errMsg), http.StatusSeeOther)
		return
	}
	//fmt.Printf("should have triggered the query ... variations: %v\n", len(availableVariations))

	// 2) add the order into the cart

	// 2.1) get the cart
	shoppingCart, err := cart.GetCart(w, sessionManager.Load(r))
	if err != nil {
		fmt.Errorf("CartResource.Add failed to get cart: %v\n", err)
		errMsg := "The loading the cart failed."
		http.Redirect(w, r, fmt.Sprintf("%v%v#%v", localePrepath, product.GetRelativeURL(), errMsg), http.StatusSeeOther)
		return
	}
	// 2.2) add item to the session
	for _, pv := range availableVariations {
		cartItem := cart.CartItem{}
		cartItem.VariationID = pv.ID
		cartItem.Quantity = 1
		if customizeName != nil {
			cartItem.CustomizedName = *customizeName
		}
		if customizeNumber != nil {
			cartItem.CustomizedNumber = *customizeNumber
		}
		_, ok := shoppingCart.Add(&cartItem)
		if !ok {
			fmt.Errorf("CartResource.Add failed to add cartItem: %v to cart. Unknown Reason", cartItem)
		}
	}
	// 3) decide where to go next

	http.Redirect(w, r, fmt.Sprintf("%v%v#%v", localePrepath, "/cart/", "success ..."), http.StatusSeeOther)
	return
}

type CartControlsRequest struct {
	Plus              uint `schema:"plus,omitempty"`
	Minus             uint `schema:"minus,omitempty"`
	Remove            uint `schema:"remove,omitempty"`
	ProceedToCheckout bool `schema:"proceedToCheckout,omitempty"`
	ContinueShopping  bool `schema:"continueShopping,omitempty"`
}

func (cartData *CartControlsRequest) BindForm(r *http.Request) error {
	// unfortunately - chi doesn't support ContentTypeForm yet
	return formDecoder.Decode(cartData, r.PostForm)
}
func (c CartResource) Controls(w http.ResponseWriter, r *http.Request) {
	var (
		//tx = db.GetDBFromRequestContext(r)
		locale = i18n.GetLocaleContext(r)
	)
	localePrepath := ""
	if locale == "ar-AE" {
		localePrepath = "/ar"
	}
	r.ParseForm()

	cartControl := &CartControlsRequest{}
	if err := cartControl.BindForm(r); err != nil {
		// save the error
		log.Printf("CartResource.Controls CartControlsRequest.bindForm failed: %v", err)
		// redirect back to the cart
		errMsg := fmt.Errorf("invalid data in post request - bind failed")
		http.Redirect(w, r, fmt.Sprintf("%v%v#msg=%v", localePrepath, "/cart/", errMsg), http.StatusSeeOther)
		return
	}

	// case: proceed to checkout
	if cartControl.ProceedToCheckout {
		http.Redirect(w, r, fmt.Sprintf("%v%v", localePrepath, "/cart/checkout/"), http.StatusSeeOther)
		return
	}
	if cartControl.ContinueShopping {
		http.Redirect(w, r, fmt.Sprintf("%v%v", localePrepath, "/shop/any/"), http.StatusSeeOther)
		return
	}

	shoppingCart, err := cart.GetCart(w, sessionManager.Load(r))
	if err != nil {
		fmt.Errorf("CartResource.Controls failed to get cart: %v\n", err)
		http.Redirect(w, r, fmt.Sprintf("%v%v", localePrepath, "/cart/"), http.StatusSeeOther)
		return
	}

	if cartControl.Remove > 0 {
		if ok := shoppingCart.Remove(cartControl.Remove); !ok {
			fmt.Errorf("CartResource.Controls failed to remove item from cart ID: %v\n", cartControl.Remove)
			http.Redirect(w, r, fmt.Sprintf("%v%v", localePrepath, "/cart/#failedToRemoveItem"), http.StatusSeeOther)
			return
		}
	}
	isMinus := cartControl.Minus > 0
	isPlus := cartControl.Plus > 0
	// get cartItem
	cartItemID := cartControl.Plus
	if isMinus {
		cartItemID = cartControl.Minus
	}
	if cartItemID <= 0 {
		fmt.Errorf("CartResource.Controls failed to fetch item from cart. ID: %v\n", cartItemID)
		http.Redirect(w, r, fmt.Sprintf("%v%v", localePrepath, "/cart/#failedToGetItem"), http.StatusSeeOther)
		return
	}
	cartItem, ok := shoppingCart.GetContent()[cartItemID]
	if !ok {
		fmt.Errorf("CartResource.Controls failed to fetch item from cart. ID: %v\n", cartItemID)
		http.Redirect(w, r, fmt.Sprintf("%v%v", localePrepath, "/cart/#failedToGetItem"), http.StatusSeeOther)
		return
	}
	currentQuantity := cartItem.Quantity
	shoppingCart.Remove(cartItemID)
	if isMinus {
		// we dont have substract
		// need to remove the item and re-add
		newQuantity := currentQuantity - 1
		if newQuantity <= 0 {
			// no need to do anything anymore, we have remove the last quantity
			http.Redirect(w, r, fmt.Sprintf("%v%v", localePrepath, "/cart/#removeByQuantity"), http.StatusSeeOther)
			return
		}
		cartItem.Quantity = newQuantity
	}
	if isPlus {
		cartItem.Quantity = currentQuantity + 1
	}
	fmt.Printf("isPlus %v isMinus %v - quantity %v\n", isPlus, isMinus, cartItem.Quantity)
	if _, ok := shoppingCart.Add(cartItem); !ok {
		fmt.Errorf("CartResource.Controls failed to add item from cart. ID: %v\n", cartItemID)
		http.Redirect(w, r, fmt.Sprintf("%v%v", localePrepath, "/cart/#failedToAddItem"), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("%v%v", localePrepath, "/cart/"), http.StatusSeeOther)
	return
}
