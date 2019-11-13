package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"

	"regexp"
	"strings"
	"time"

	"bitbucket.org/softwarehouseio/victory/victory-frontend/config"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/config/cart"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/config/i18n"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/config/insights"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/db"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/models"
	"github.com/alexedwards/scs"
	"github.com/flosch/pongo2"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/gorilla/schema"
	"github.com/jinzhu/gorm"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/vincent-petithory/countries"
	"gopkg.in/go-playground/validator.v9"
)

var (
	validate *validator.Validate
)

const (
	KOrderID                                  = "orderID"
	KCheckoutFormPaymentMethodCreditCardValue = "creditCard"
	KCheckoutFormPaymentMethodCOD             = "cashOnDelivery"
	KDeliveryChargesExpress                   = 3 // 3 aed
	KDeliveryChargesRegular                   = 0 // 0 aed
)

func init() {
	validate = validator.New()
}

type CheckoutResource struct {
	BaseURL string
}

func (co CheckoutResource) ContextKey() string {
	return "checkoutCtx"
}

func (co CheckoutResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", co.View)
	r.Post("/process", co.Process)
	r.Get(fmt.Sprintf("/order-gateway/{%v}/", KOrderID), co.OrderGateway)
	r.Get(fmt.Sprintf("/order-received/{%v}/", KOrderID), co.OrderReceived)
	r.Get(fmt.Sprintf("/order-state/{%v}/", KOrderID), co.OrderState)
	r.Get("/webhooks/eu-gateway", co.WebhookMastercard)
	r.Post("/webhooks/eu-gateway", co.WebhookMastercard)

	return r
}

// middleware

// view controllers
func CalculateSubtotal(selectedProductVariations []*models.ProductVariation, shoppingCart *cart.Cart) float64 {
	subtotal := float64(0)

	for _, pv := range selectedProductVariations {

		thisCartItem, ok := shoppingCart.CartItems[pv.ID]
		var quantity uint = 0
		if ok {
			// only if the item is in the cart we want to count it in the subtotal
			quantity = thisCartItem.Quantity
		}

		subtotal += pv.GetPrice() * float64(quantity)
	}
	return subtotal
}
func CalculateVAT(subtotal float64) float64 {
	vat := subtotal * config.Config.VAT
	return vat
}
func CalculateShippingCost(shippingMethod string) float64 {
	shipping := float64(KDeliveryChargesRegular)
	if shippingMethod == "expressDelivery" {
		shipping = float64(KDeliveryChargesExpress)
	}
	return shipping
}

func (co CheckoutResource) View(w http.ResponseWriter, r *http.Request) {
	var (
		tpl        = pongo2.Must(pongo2.FromFile("templates/html/checkout.html"))
		tplContext = WebResource{}.GetTplContext(r)
		tx         = db.GetDBFromRequestContext(r)
		user, _    = authResource.GetUserSession(r).(*models.User)
		address    = models.Address{}
	)

	session := sessionManager.Load(r)
	shoppingCart, err := cart.GetCart(w, session)
	if err != nil {
		log.Printf("CheckoutResource.View failed to retrieve cart reason: %+v", shoppingCart)
	}
	if shoppingCart.IsEmpty() {
		log.Printf("CheckoutResource.View. Tried to acceess checkout with an empty shopping cart ... ")
		http.Redirect(w, r, fmt.Sprintf("%v", "/cart/"), http.StatusSeeOther)
	}
	cartItemIDs := shoppingCart.GetItemsIDS()

	selectedProductVariations := []*models.ProductVariation{}
	if err := tx.Model(models.ProductVariation{}).
		Where("id in (?)", cartItemIDs).
		Preload("Product").
		Preload("Size").
		Preload("Badge").
		Find(&selectedProductVariations).Error; err != nil {
		fmt.Errorf("failed to fetch product variations based on cart reason: %v", err)
	}

	if user != nil {
		if err := tx.Model(&address).Where("user_id = ?", user.ID).First(&address).Error; err != nil && err != gorm.ErrRecordNotFound {
			log.Printf("CR.View - Address query failed: %v", err)
		}
	}

	// previous data
	checkoutData := CheckoutRequest{}
	_ = checkoutData.BindSession(session)

	// only if the checkout request is empty, lets try to render ...
	if ok, err := session.Exists(checkoutData.SessionKey()); user != nil && (!ok || err != nil) {
		// either something is not ok, or we received an error
		checkoutData.BindAddress(&address)
		checkoutData.Email = user.Email
	}

	// calculate subtotal
	subtotal := CalculateSubtotal(selectedProductVariations, shoppingCart)
	// add VAT
	vat := CalculateVAT(subtotal)
	// add shipment
	shipping := CalculateShippingCost(checkoutData.Delivery)

	// total
	total := subtotal + shipping + vat

	tplContext.Update(pongo2.Context{
		"selectedProductVariations": selectedProductVariations,
		"cartItems":                 shoppingCart.CartItems,
		"subtotal":                  subtotal,
		"vat":                       vat,
		"shipping":                  shipping,
		"total":                     total,
		"cd":                        checkoutData,
		"countries":                 countries.Countries,
		"DeliveryChargesExpress":    KDeliveryChargesExpress,
	})

	if err := tpl.ExecuteWriter(tplContext, w); err != nil {
		fmt.Printf("Tpl.ExecuteWriter Failed %v", err)
		e := ErrInternalServerError(err)
		render.Render(w, r, e)
		return
	}
}

type AddressRequest struct {
	ContactName  string `schema:",omitempty" validate:"required,lt=300"`
	Telephone    string `schema:",omitempty" validate:"required,lt=22"`
	Email        string `schema:",omitempty" validate:"required,email"`
	AddressLine1 string `schema:",omitempty" validate:"required,lt=300"`
	AddressLine2 string `schema:",omitempty" validate:"required,lt=300"`
	PostalCode   string `schema:",omitempty" validate:"lt=7"`
	City         string `schema:",omitempty" validate:"required,lt=300"`
	Country      string `schema:",omitempty" validate:"required,lt=3,gt=1"`
	Notes        string `schema:",omitempty" validate:"lt=300"`
}

func (a AddressRequest) Actions() map[string]bool {
	return map[string]bool{
		"saveShipping": true,
	}
}
func (reqData *AddressRequest) BindForm(r *http.Request) error {
	// unfortunately - chi doesn't support ContentTypeForm yet
	formDec := schema.NewDecoder()
	formDec.IgnoreUnknownKeys(true)
	formDec.ZeroEmpty(true)
	return formDec.Decode(reqData, r.PostForm)
}

// AddressReq.BindAddressModel assigns the values of Rq to Mdl
func (a *AddressRequest) BindAddressModel(na *models.Address) {
	na.ContactName = a.ContactName
	na.Telephone = a.Telephone
	na.City = a.City
	na.Country = a.Country
	na.AddressLine1 = a.AddressLine1
	na.AddressLine2 = a.AddressLine2
	na.PostalCode = a.PostalCode
	na.Notes = a.Notes
}

// BindAddress assigns the values of Mdl to Rq
func (a *AddressRequest) BindAddress(na *models.Address) {
	a.ContactName = na.ContactName
	a.Telephone = na.Telephone
	a.City = na.City
	a.Country = na.Country
	a.AddressLine1 = na.AddressLine1
	a.AddressLine2 = na.AddressLine2
	a.PostalCode = na.PostalCode
	a.Notes = na.Notes
}

func (c *AddressRequest) Validate() (bool, []error) {
	errs := []error{}

	// catch xss
	c.ContactName = template.HTMLEscapeString(c.ContactName)
	c.Telephone = template.HTMLEscapeString(c.Telephone)
	c.Email = template.HTMLEscapeString(c.Email)
	c.AddressLine1 = template.HTMLEscapeString(c.AddressLine1)
	c.AddressLine2 = template.HTMLEscapeString(c.AddressLine2)
	c.PostalCode = template.HTMLEscapeString(c.PostalCode)
	c.City = template.HTMLEscapeString(c.City)
	c.Country = template.HTMLEscapeString(c.Country)
	c.Notes = template.HTMLEscapeString(c.Notes)

	if err := validate.Struct(c); err != nil {
		msgs := err.(validator.ValidationErrors)
		for _, m := range msgs {
			log.Printf("Form Validation Error: %v\n", m)
		}
		errs = append(errs, fmt.Errorf("validation failed: %v", msgs.Error()))
	}

	if c.Email != "" {
		if m, _ := regexp.MatchString(`^([\w\.\_]{2,10})@(\w{1,}).([a-z]{2,4})$`, c.Email); !m {
			log.Printf("Email Malformed\n")
			errs = append(errs, fmt.Errorf("Email malformed. %v", c.Email))
		}
	}

	isValid := len(errs) == 0
	return isValid, errs
}

type CheckoutRequest struct {
	AddressRequest

	Delivery string `schema:",omitempty" validate:"required,lt=300"`
	Payment  string `schema:",omitempty" validate:"required,lt=20"`

	Action string `schema:",omitempty" validate:"required,lt=300"`
}

// Action() map[string]bool - returns all supported actions
func (c CheckoutRequest) Actions() map[string]bool {
	actions := c.AddressRequest.Actions()
	actions["saveDelivery"] = true
	actions["order"] = true
	return actions
}
func (c CheckoutRequest) SessionKey() string {
	return "checkoutCtx"
}

func (reqData *CheckoutRequest) BindForm(r *http.Request) error {
	// unfortunately - chi doesn't support ContentTypeForm yet
	formDec := schema.NewDecoder()
	formDec.IgnoreUnknownKeys(true)
	formDec.ZeroEmpty(true)
	return formDec.Decode(reqData, r.PostForm)
}
func (reqData *CheckoutRequest) BindSession(session *scs.Session) error {
	return session.GetObject(reqData.SessionKey(), reqData)
}
func (c *CheckoutRequest) Validate() (bool, []error) {
	errs := []error{}

	// catch xss
	c.ContactName = template.HTMLEscapeString(c.ContactName)
	c.Telephone = template.HTMLEscapeString(c.Telephone)
	c.Email = template.HTMLEscapeString(c.Email)
	c.AddressLine1 = template.HTMLEscapeString(c.AddressLine1)
	c.AddressLine2 = template.HTMLEscapeString(c.AddressLine2)
	c.PostalCode = template.HTMLEscapeString(c.PostalCode)
	c.City = template.HTMLEscapeString(c.City)
	c.Country = template.HTMLEscapeString(c.Country)
	c.Notes = template.HTMLEscapeString(c.Notes)
	c.Delivery = template.HTMLEscapeString(c.Delivery)
	c.Action = template.HTMLEscapeString(c.Action)

	if err := validate.Struct(c); err != nil {
		msgs := err.(validator.ValidationErrors)
		for _, m := range msgs {
			log.Printf("Form Validation Error: %v\n", m)
		}
		errs = append(errs, fmt.Errorf("validation failed: %v; value: ", msgs.Error()))
	}

	if c.Email != "" {
		if m, _ := regexp.MatchString(`^([\w\.\_]{2,10})@(\w{1,}).([a-z]{2,4})$`, c.Email); !m {
			log.Printf("Email Malformed\n")
			errs = append(errs, fmt.Errorf("Email malformed. %v", c.Email))
		}
	}

	if _, ok := c.Actions()[c.Action]; !ok {
		log.Printf("No Action submitted.\n")
		errs = append(errs, fmt.Errorf("No Action Defined."))
	}

	if c.Action == "order" && c.Payment == "" {
		log.Printf("Can't process order without a selected payment option.\n")
		errs = append(errs, fmt.Errorf("Can't process order without a selected payment option."))
	}

	isValid := len(errs) == 0

	if !isValid {
		msg := ""
		for _, e := range errs {
			msg += e.Error()
			msg += " - "
		}
		insights.Sentry.CaptureMessage(
			fmt.Sprintf("CheckoutRequest.Validate failed: %v", msg),
			map[string]string{
				"package":  "checkout",
				"instance": "CheckoutRequest",
				"method":   "Validate",
			})
	}

	return isValid, errs
}

func (co CheckoutResource) Process(w http.ResponseWriter, r *http.Request) {
	var (
		session     = sessionManager.Load(r)
		checkoutReq = CheckoutRequest{}
		tx          = db.GetDBFromRequestContext(r)
		user, _     = authResource.GetUserSession(r).(*models.User)
		tplContext  = WebResource{}.GetTplContext(r)
		locale      = i18n.GetLocaleContext(r)
	)
	localePrepath := ""
	if locale == "ar-AE" {
		localePrepath = "/ar"
	}

	r.ParseForm()
	if err := checkoutReq.BindForm(r); err != nil {
		// save the error
		fmt.Printf("\nCheckoutResource.Shipment CheckoutRequest.bindForm failed: %v", err)
		// redirect back to the cart
		errMsg := fmt.Errorf("invalid data in post request - bind failed")
		http.Redirect(w, r, fmt.Sprintf("%v#msg=%v", "/cart/checkout/", errMsg), http.StatusSeeOther)
		return
	}
	if ok, errs := checkoutReq.Validate(); !ok {
		for _, err := range errs {
			log.Printf("CheckoutRequest.Validate failed: %v", err)
		}
		errMsg := fmt.Errorf("invalid data in post request - validation failed")
		http.Redirect(w, r, fmt.Sprintf("%v#msg=%v", "/cart/checkout/", errMsg), http.StatusSeeOther)
		return
	}

	if err := session.PutObject(w, checkoutReq.SessionKey(), checkoutReq); err != nil {
		log.Printf("CheckoutRequest.session store failed: %v", err)
		errMsg := fmt.Errorf("internalServerError")
		http.Redirect(w, r, fmt.Sprintf("%v#msg=%v", "/cart/checkout/", errMsg), http.StatusSeeOther)
		return
	}

	// only continue after this if we're ready to actually do the checkout and payment
	if checkoutReq.Action != "order" {
		http.Redirect(w, r, fmt.Sprintf("%v", "/cart/checkout/"), http.StatusSeeOther)
		return
	}

	// TODO: 	only on CashOnDelivery: process and store the actual order right away -> persist it to DB, deduct stock, etc.
	// 			and redirect to the order confirmation.

	// create a new order
	// 1) grab cart content

	//orderItem := &models.OrderItem{}
	shoppingCart, err := cart.GetCart(w, sessionManager.Load(r))
	if err != nil {
		fmt.Printf("failed to retrieve cart reason: %+v", shoppingCart)
	}

	cartItemIDs := shoppingCart.GetItemsIDS()
	cartItems := shoppingCart.GetContent()

	selectedProductVariations := []*models.ProductVariation{}
	if err := tx.Model(models.ProductVariation{}).
		Where("id in (?)", cartItemIDs).
		Preload("Product").
		Preload("Size").
		Preload("Badge").
		Find(&selectedProductVariations).Error; err != nil {
		log.Printf("failed to fetch product variations based on cart reason: %v", err)
	}
	// only create an order if we actually have items to process :)
	if len(selectedProductVariations) == 0 {
		log.Printf("CheckoutResource.Process. Tried to submit an order without any items ... odd! exiting")
		http.Redirect(w, r, fmt.Sprintf("%v", "/cart/checkout/"), http.StatusSeeOther)
		return
	}

	address := models.Address{
		ContactName:  checkoutReq.ContactName,
		Telephone:    checkoutReq.Telephone,
		City:         checkoutReq.City,
		Country:      checkoutReq.Country,
		AddressLine1: checkoutReq.AddressLine1,
		AddressLine2: checkoutReq.AddressLine2,
	}

	if user != nil {
		address.UserID = &user.ID
	}

	// calculate subtotal
	subtotal := CalculateSubtotal(selectedProductVariations, shoppingCart)
	// add VAT
	vat := CalculateVAT(subtotal)
	// add shipment
	shipping := CalculateShippingCost(checkoutReq.Delivery)
	// total
	total := subtotal + shipping + vat

	// create the order and store all content
	order := &models.Order{
		ShippingMethod:  checkoutReq.Delivery,
		PaymentMethod:   checkoutReq.Payment,
		ShippingAddress: address,
		Subtotal:        subtotal,
		VAT:             vat,
		ShippingCost:    shipping,
		Total:           total,
		Notes:           checkoutReq.Notes,
		Email:           checkoutReq.Email,
	}

	if user != nil {
		order.User = *user
	}
	order.SetState("draft")

	if err := tx.Create(&order).Error; err != nil {
		log.Printf("CheckoutRes.Process Create 1 Err: %v", err)
		http.Redirect(w, r, fmt.Sprintf("%v", "/cart/checkout/#msg=process failed to create new order object"), http.StatusSeeOther)
		return
	}
	//if err := models.OrderState.Trigger("checkout", order, tx, "user pressed submit on checkout page - before adding all order items"); err != nil {
	//	log.Printf("CheckoutRes.Process Create 1 Err: %v", err)
	//	http.Redirect(w, r, fmt.Sprintf("%v","/cart/checkout/#msg=process failed to create new order object"), http.StatusSeeOther)
	//	return
	//}

	// create the individual order items
	for _, item := range selectedProductVariations {
		cartItm, ok := cartItems[item.ID]
		if !ok {
			log.Printf("CheckoutResource.Process. cartItem missing for item.ID %v", item.ID)
			continue
		}
		log.Printf("have in cart PVA.ID: %v. quantity ordered: %v - name: %v number :%v",
			item.ID, cartItm.Quantity, cartItm.CustomizedName, cartItm.CustomizedNumber)
		itm := models.OrderItem{
			OrderID:          order.ID,
			ProductVariation: *item,
			CustomizedName:   cartItm.CustomizedName,
			CustomizedNumber: cartItm.CustomizedNumber,
			Quantity:         cartItm.Quantity,
			Price:            item.Product.Price,
		}
		log.Printf(" ")
		log.Printf("Order Item: %v", itm)
		log.Printf(" ")
		itm.PersistProductDetailsJSON, err = itm.JSONString()
		if err != nil {
			log.Printf("CheckoutResource.Process failed serializing OrderItem to JSON")
		}
		itm.SetState("draft")
		if err := tx.Create(&itm).Error; err != nil {
			log.Printf("CheckoutRes.Process Create 2 Err: %v", err)
			http.Redirect(w, r, fmt.Sprintf("%v", "/cart/checkout/#msg=process failed to create new order object"), http.StatusSeeOther)
			return
		}

		order.OrderItems = append(order.OrderItems, itm)
	}
	order.Reference = order.GeneratePaymentReference()

	if err := models.OrderState.Trigger("checkout", order, tx, "added the orderItems - user pressed submit on the checkout page"); err != nil {
		log.Printf("CheckoutRes.Process Create Err: %v", err)
		http.Redirect(w, r, fmt.Sprintf("%v", "/cart/checkout/#msg=process failed to update new order object with order items"), http.StatusSeeOther)
		return
	}

	// based on the selection (cash on delivery/pay by creditcard) either redirect the user
	// to the payment gateway integration page or send him to the confirmation

	orderReference := order.Reference
	// default to payment gateway, bcs ... well, it makes sense
	orderRedirectURL := fmt.Sprintf("%v/cart/checkout/order-gateway/%v/", localePrepath, orderReference)

	if order.PaymentMethod == KCheckoutFormPaymentMethodCreditCardValue {
		// just for completeness sake ... to highlight both options
		orderRedirectURL = fmt.Sprintf("%v/cart/checkout/order-gateway/%v/", localePrepath, orderReference)
	}
	if order.PaymentMethod == KCheckoutFormPaymentMethodCOD {
		// if the payment method is cash on delivery, we can do everything needed right away
		shoppingCart.EmptyCart()
		co.EmailConfirmation(order, tplContext)
		orderRedirectURL = fmt.Sprintf("%v/cart/checkout/order-received/%v/", localePrepath, orderReference)
	}

	http.Redirect(w, r, orderRedirectURL, http.StatusSeeOther)
}

func (co CheckoutResource) EmailConfirmation(order *models.Order, tplContext pongo2.Context) {
	// setup email with sendgrid
	//
	m := mail.NewV3Mail()
	address := config.Config.EmailReplyAddress
	name := config.Config.EmailReplyName
	m.SetFrom(mail.NewEmail(name, address))

	p := mail.NewPersonalization()
	tos := []*mail.Email{
		mail.NewEmail(order.ShippingAddress.ContactName, order.Email),
	}
	p.AddTos(tos...)

	m.Subject = fmt.Sprintf("Order # %v confirmation.", order.ID)
	m.AddPersonalizations(p)

	//c := mail.NewContent("text/html", )
	tmplt, ok := config.EmailTemplates["customer-processing-order"]
	if !ok {
		log.Println("Tpl.ExecuteWriter Failed to write tempalte not found")
		return
	}

	tpl := pongo2.Must(pongo2.FromFile(tmplt))
	tplContext.Update(pongo2.Context{
		"order": order,
	})
	emailHtmlContent, err := tpl.Execute(tplContext)
	if err != nil {
		fmt.Printf("Tpl.ExecuteWriter Failed %v", err)
		return
	}

	c := mail.NewContent("text/html", emailHtmlContent)
	m.AddContent(c)
	//
	// sendgrid API
	key := config.ENVSendgrid()
	if key == "" {
		log.Printf("EmailConfirmation failed, no %v sendgrid key defined in ENV", config.OO_ENV_SENDGRID)
		return
	}
	request := sendgrid.GetRequest(key, config.SendGridAPIEndpoint, config.SendGridAPIDomain)
	request.Method = "POST"
	request.Body = mail.GetRequestBody(m)
	response, err := sendgrid.API(request)
	if err != nil {
		log.Printf("EmailConfirmation failed, send failed: %v", err)
		return
	}
	log.Printf("EmailConfirmation success: StatusCode %v, \n Headers: %v, \n Body %v", response.StatusCode, response.Headers, response.Body)
}

type PaymentGatewayReceipt struct {
	Amount     float64                `json:"amount"`
	Billing    map[string]interface{} `json:"billing"`
	Chargeback struct {
		Amount   float64 `json:"amount"`
		Currency string  `json:"currency"`
	} `json:"chargeback"`
	CreationTime          string                      `json:"creationTime"`
	Currency              string                      `json:"currency"`
	Customer              map[string]interface{}      `json:"customer"`
	Description           string                      `json:"description"`
	Device                map[string]interface{}      `json:"device"`
	FundingStatus         string                      `json:"fundingStatus"`
	ID                    string                      `json:"id"` // order.Reference
	Merchant              string                      `json:"merchant"`
	MerchantCategoryCode  int                         `json:"merchantCategoryCode,string"`
	Result                string                      `json:"result"`
	Risk                  map[string]interface{}      `json:"risk"`
	SourceOfFunds         map[string]interface{}      `json:"sourceOfFunds"`
	Status                string                      `json:"status"`
	TotalAuthorizedAmount float64                     `json:"totalAuthorizedAmount"`
	TotalRefundedAmount   float64                     `json:"totalRefundedAmount"`
	TotalCapturedAmount   float64                     `json:"totalCapturedAmount"`
	Transaction           []PaymentGatewayTransaction `json:"transaction"` // array

}
type PaymentGatewayOrder struct {
	Amount     float64 `json:"amount"`
	Chargeback struct {
		Amount   float64 `json:"amount"`
		Currency string  `json:"currency"`
	} `json:"chargeback"`
	CreationTime          string  `json:"creationTime"`
	Currency              string  `json:"currency"`
	Description           string  `json:"description"`
	FundingStatus         string  `json:"fundingStatus"`
	ID                    string  `json:"id"` // order.Reference
	MerchantCategoryCode  int     `json:"merchantCategoryCode,string"`
	Status                string  `json:"status"`
	TotalAuthorizedAmount float64 `json:"totalAuthorizedAmount"`
	TotalRefundedAmount   float64 `json:"totalRefundedAmount"`
	TotalCapturedAmount   float64 `json:"totalCapturedAmount"`
}

type PaymentGatewayTransaction struct {
	Secure3DID            string                           `json:"3DSecureId"`
	Secure3D              map[string]interface{}           `json:"3DSecure"`
	AuthorizationResponse map[string]interface{}           `json:"authorizationResponse"`
	Billing               map[string]interface{}           `json:"billing"`
	Customer              map[string]interface{}           `json:"customer"`
	Device                map[string]interface{}           `json:"device"`
	GatewayEntryPoint     string                           `json:"gatewayEntryPoint"`
	Merchant              string                           `json:"merchant"`
	Order                 PaymentGatewayOrder              `json:"order"`
	Response              map[string]interface{}           `json:"response"`
	Result                string                           `json:"result"`
	Risk                  map[string]interface{}           `json:"risk"`
	SourceOfFunds         map[string]interface{}           `json:"sourceOfFunds"`
	TimeOfRecord          string                           `json:"timeOfRecord"`
	TransactionDetails    PaymentGatewayTransactionDetails `json:"transaction"`
	Version               string                           `version`
}
type PaymentGatewayTransactionDetails struct {
	Acquirer          map[string]interface{} `json:"acquirer"`
	Amount            float64                `json:"amount"`
	AuthorizationCode string                 `json:"authorizationCode"`
	Currency          string                 `json:"currency"`
	Frequency         string                 `json:"frequency"`
	Funding           map[string]string      `json:"funding"`
	ID                string                 `json:"id"` // order.Reference
	Receipt           string                 `json:"receipt"`
	Source            string                 `json:"source"`
	Terminal          string                 `json:"terminal"`
	Type              string                 `json:"type"`
}

func (co CheckoutResource) WebhookMastercard(w http.ResponseWriter, r *http.Request) {
	var (
		tx = db.GetDBFromRequestContext(r)
	)
	resp := struct {
		Done string
	}{
		Done: "done",
	}
	// check the X-Notification-Secret for:
	// B6BB40B0C62C9B91B9A61C1B83DA6E96
	/*
		- The X-Notification-Id header uniquely identifies the notification. This header will be identical for duplicate transactions.
		- The X-Notification-Attempt header indicates the number of attempts made to send the notification.
	*/
	/*
		map[
			sourceOfFunds: map[
				provided: map[
					card: map[
						expiry: map[month:5 year:21]
						fundingMethod:DEBIT
						issuer:U.S. BANK NATIONAL ASSOCIATION
						nameOnCard:Test Tester
						number:511111xxxxxx1118
						scheme:MASTERCARD
						brand:MASTERCARD
					]
				]
				type:CARD
			]
			3DSecureId:e15577be-ae24-48d7-9894-2156addd9119
			authorizationResponse:map[
				commercialCard:888
				processingCode:003000
				responseCode:00
				stan:113994
				avsCode:A
				cardSecurityCodeError:M
				transactionIdentifier:123456789
				commercialCardIndicator:3
				financialNetworkCode:777
			]
			customer:map[
				firstName:Test
				lastName:Tester
			]
			device:map[
				browser:Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36
				ipAddress:94.207.120.136
			]
			gatewayEntryPoint:CHECKOUT
			response:map[
				acquirerCode:00
				acquirerMessage:Approved
				cardSecurityCode:map[
					acquirerCode:M
					gatewayCode:MATCH
				]
				cardholderVerification:map[
					avs:map[
						acquirerCode:A
						gatewayCode:ADDRESS_MATCH
					]
				]
				gatewayCode:APPROVED
				gatewayRecommendation:PROCEED
			]
			result:SUCCESS
			timeOfRecord:2018-07-25T16:22:15.155Z
			billing:map[
				address:map[
					street:10001 Alpha St
				]
			]
			order:map[
				currency:AED
				id:5
				totalCapturedAmount:346.5
				amount:346.5
				creationTime:2018-07-25T16:22:15.155Z
				description:Your order reference number is: 5
				fundingStatus:NOT_SUPPORTED
				merchantCategoryCode:5651
				status:CAPTURED
				totalAuthorizedAmount:346.5
				totalRefundedAmount:0
				chargeback:map[
					amount:0
					currency:AED
				]
			]
			risk:map[
				response:map[
					gatewayCode:REVIEW_REQUIRED
					review:map[
						decision:PENDING
					]
					rule:[
						map[id:GATEKEEPER
							name:Gatekeeper
							recommendation:REVIEW
							score:0
							type:EXTERNAL_RULE
						]
						map[name:MSO_3D_SECURE recommendation:NO_ACTION type:MSO_RULE data:NO_LIABILITY_SHIFT]
						map[data:511111 name:MSO_BIN_RANGE recommendation:NO_ACTION type:MSO_RULE]
						map[recommendation:NO_ACTION type:MSO_RULE data:M name:MSO_CSC]
						map[data:94.207.120.136 name:MSO_IP_ADDRESS_RANGE recommendation:NO_ACTION type:MSO_RULE]
						map[data:ARE name:MSO_IP_COUNTRY recommendation:NO_ACTION type:MSO_RULE]
					]
					totalScore:0
				]
			]
			transaction:map[
				amount:346.5
				source:INTERNET
				terminal:ADIB0002
				type:PAYMENT
				acquirer:map[
					timeZone:+0400
					transactionId:123456789
					batch:2.0180725e+07
					date:0725
					id:ADIB_S2I
					merchantId:7008899
					settlementDate:2018-07-25
				]
				authorizationCode:113994
				currency:AED
				frequency:SINGLE
				funding:map[status:NOT_SUPPORTED]
				id:1
				receipt:820616113994
			]
			version:48
			3DSecure:map[veResEnrolled:N xid:2lO0M34u5D5T6fa7r+GdRISo5fU=]
			merchant:TEST7008899
		]
	*/

	log.Printf("checkout.WebhookMastercard Method: %v", r.Method)
	log.Printf("checkout.WebhookMastercard Headers: %v", r.Header)
	log.Printf("checkout.WebhookMastercard Parameters: %v", r.URL.Query())

	payload := PaymentGatewayTransaction{}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("checkout.WebhookMastercard failed: %v or len(data) is 0", err)
		render.Render(w, r, ErrInternalServerError(err))
		return
	}

	var prettyJSON bytes.Buffer
	err2 := json.Indent(&prettyJSON, data, "", "\t")
	log.Printf("received: %v, %v", string(prettyJSON.Bytes()), err2)

	err = json.Unmarshal(data, &payload)
	if err != nil {
		render.Render(w, r, ErrInternalServerError(err))
		return
	}
	orderReference := payload.Order.ID

	if payload.Result != "SUCCESS" || orderReference == "" {
		msg := fmt.Sprintf("CheckoutResource.WebhookMastercard failed; order reference: %v transaction failed with result %v", orderReference, payload.Result)
		insights.Sentry.CaptureMessage(
			msg,
			map[string]string{
				"package":  "checkout",
				"instance": "CheckoutResource",
				"method":   "WebhookMastercard",
			})
		log.Printf("CheckoutRes.WebhookMastercard pay Err: %v", msg)
		// return OK to the gateway so it doesn't keep trying ...
		render.Status(r, http.StatusOK)
		render.JSON(w, r, resp)
		return
	}

	order := models.Order{}
	if err := tx.Model(order).Where("reference = ?", orderReference).
		Find(&order).Error; err != nil {
		msg := fmt.Sprintf("CheckoutResource.WebhookMastercard failed; order reference: %v transaction failed.", orderReference)
		insights.Sentry.CaptureMessage(
			msg,
			map[string]string{
				"package":  "checkout",
				"instance": "CheckoutResource",
				"method":   "WebhookMastercard",
			})
		log.Print(msg)
		// return OK to the gateway so it doesn't keep trying ...
		render.Render(w, r, ErrInternalServerError(err))
		return
	}
	if order.ID == 0 {
		msg := fmt.Sprintf("CheckoutResource.WebhookMastercard failed; order reference: %v order doesnt exist", orderReference)
		insights.Sentry.CaptureMessage(
			msg,
			map[string]string{
				"package":  "checkout",
				"instance": "CheckoutResource",
				"method":   "WebhookMastercard",
			})
		log.Printf("CheckoutRes.WebhookMastercard pay Err: %v", msg)
		// return OK to the gateway so it doesn't keep trying ...
		render.Status(r, http.StatusOK)
		render.JSON(w, r, resp)
		return
	}
	// this should set the state to paid
	if err := models.OrderState.Trigger("pay", &order, tx, "payment gateway responded with a a success response"); err != nil {
		insights.Sentry.CaptureMessage(
			fmt.Sprintf("CheckoutResource.WebhookMastercard failed; order reference: %v not able to save that the 'pay' state trigger", orderReference),
			map[string]string{
				"package":  "checkout",
				"instance": "CheckoutResource",
				"method":   "WebhookMastercard",
			})

		log.Printf("CheckoutRes.WebhookMastercard pay Err: %v", err)
		// return OK to the gateway so it doesn't keep trying ...
		render.Status(r, http.StatusOK)
		render.JSON(w, r, resp)
		return
	}

	log.Printf("checkout.WebhookMastercard succeeded: %v", payload)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
	return
}

// OrderGatewayRetrieveSessionID
func (co CheckoutResource) OrderGatewayRetrieveSessionID(r *http.Request, url, apiOperation, apiPassword, returnURL, merchant, orderID, orderAmount, orderCurrency string) (string, string) {

	err := r.ParseForm()
	if err != nil {
		log.Printf("http.Request parse was failed %v", err)
	}

	apiUsername := "merchant." + merchant

	jsonGetSessionRequest := fmt.Sprintf("apiOperation=%s&apiPassword=%s&interaction.returnUrl=%s&apiUsername=%s&merchant=%s&order.id=%s&order.amount=%s&order.currency=%s",
		apiOperation, apiPassword, returnURL, apiUsername, merchant, orderID, orderAmount, orderCurrency)

	body := strings.NewReader(jsonGetSessionRequest)

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		fmt.Println(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("unexpected error in sending req %v", err)
	}

	rBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	rBodyStr := string(rBody)

	fmt.Printf("rBodyStr=%s\n", rBodyStr)
	rBodyIdStrArr := strings.SplitAfter(rBodyStr, "session.id=")
	if len(rBodyIdStrArr) < 2 {
		return "", ""
	}

	sessionIdArr := strings.Split(rBodyIdStrArr[1], "&")
	if len(sessionIdArr) == 0 {
		log.Println("Session Id info was failed")
	}

	sessionID := sessionIdArr[0]

	rBodyVersionStrArr := strings.SplitAfter(rBodyStr, "session.version=")
	sessionVersionArr := strings.Split(rBodyVersionStrArr[1], "&")
	if len(sessionVersionArr) == 0 {
		log.Println("Session Version info was failed")
	}

	sessionVersion := sessionVersionArr[0]

	return sessionID, sessionVersion
}

// OrderGateway
func (co CheckoutResource) OrderGateway(w http.ResponseWriter, r *http.Request) {
	var (
		tpl           = pongo2.Must(pongo2.FromFile("templates/html/checkout-gateway.html"))
		tplContext    = WebResource{}.GetTplContext(r)
		tx            = db.GetDBFromRequestContext(r)
		locale        = i18n.GetLocaleContext(r)
		url           = "https://eu-gateway.mastercard.com/api/nvp/version/49"
		apiOperation  = "CREATE_CHECKOUT_SESSION"
		apiPassword   = "6bd29ed347449c82289929a70b2aaf46"
		returnURL     = "http://localhost/feedback"
		orderCurrency = "AED"
	)

	localePrepath := ""
	if locale == "ar-AE" {
		localePrepath = "/ar"
	}

	log.Printf("checkout.OrderGateway rendering")

	canceled := len(r.URL.Query()["canceled"]) > 0

	order := models.Order{}
	orderReference := chi.URLParam(r, KOrderID)
	if err := tx.Model(order).Where("reference = ?", orderReference).
		Preload("OrderItems").
		Preload("OrderItems.ProductVariation").
		Preload("OrderItems.ProductVariation.Size").
		Preload("OrderItems.ProductVariation.Product").
		Preload("OrderItems.ProductVariation.Badge").
		Find(&order).Error; err != nil {
		log.Printf("not able to find order %v", err)
		render.Render(w, r, ErrNotFound)
		return
	}
	// confirm that this order is pending payment
	// if not redirect to the confirmation page
	if order.State == "paid" {
		http.Redirect(w, r, fmt.Sprintf("%v/cart/checkout/order-received/%v/", localePrepath, order.Reference), http.StatusSeeOther)
		return
	}

	gateway := config.ENVGetPaymentGateway()

	orderID := fmt.Sprintf("%d", order.ID)
	orderAmount := fmt.Sprintf("%f", order.Total)
	sessionID, sessionVersion := co.OrderGatewayRetrieveSessionID(r, url, apiOperation, apiPassword, returnURL, gateway.MerchantID, orderID, orderAmount, orderCurrency)

	tplContext.Update(pongo2.Context{
		config.GATEWAY_MERCHANT_ID:            gateway.MerchantID,
		config.GATEWAY_MERCHANT_NAME:          gateway.MerchantName,
		config.GATEWAY_MERCHANT_ADDRESS_LINE1: gateway.MerchantAddressLine1,
		config.GATEWAY_MERCHANT_ADDRESS_LINE2: gateway.MerchantAddressLine2,
		"gatewayCompleteReturnPage":           returnURL,
		"order":                               &order,
		"canceled":                            canceled,
		"SESSION_ID":                          sessionID,
		"SESSION_VERSION":                     sessionVersion,
	})

	if err := tpl.ExecuteWriter(tplContext, w); err != nil {
		log.Printf("Tpl.ExecuteWriter Failed %v", err)
		e := ErrInternalServerError(err)
		render.Render(w, r, e)
		return
	}
}

func (co CheckoutResource) OrderState(w http.ResponseWriter, r *http.Request) {
	var (
		tx = db.GetDBFromRequestContext(r)
	)
	resp := struct {
		State     string
		UpdatedAt time.Time
	}{
		State:     "draft",                          // default state
		UpdatedAt: time.Now().Add(-100 * time.Hour), // default now minus 100 hours ... expired
	}

	order := models.Order{}
	orderReference := chi.URLParam(r, KOrderID)
	if err := tx.Model(order).Where("reference = ?", orderReference).
		Preload("OrderItems").
		Preload("OrderItems.ProductVariation").
		Preload("OrderItems.ProductVariation.Size").
		Preload("OrderItems.ProductVariation.Product").
		Preload("OrderItems.ProductVariation.Badge").
		Find(&order).Error; err != nil {
		log.Printf("not able to find order %v", err)
		render.Render(w, r, ErrNotFound)
		return
	}

	resp.State = order.State
	resp.UpdatedAt = order.UpdatedAt

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}

func (co CheckoutResource) OrderReceived(w http.ResponseWriter, r *http.Request) {
	var (
		tpl        = pongo2.Must(pongo2.FromFile("templates/html/checkout-order-received.html"))
		tplContext = WebResource{}.GetTplContext(r)
		tx         = db.GetDBFromRequestContext(r)
	)

	order := models.Order{}
	orderReference := chi.URLParam(r, KOrderID)
	if err := tx.Model(order).Where("reference = ?", orderReference).
		Preload("OrderItems").
		Preload("OrderItems.ProductVariation").
		Preload("OrderItems.ProductVariation.Size").
		Preload("OrderItems.ProductVariation.Product").
		Preload("OrderItems.ProductVariation.Badge").
		Find(&order).Error; err != nil {
		log.Printf("not able to find order %v", err)
		render.Render(w, r, ErrNotFound)
		return
	}

	// calculate subtotal
	subtotal := order.Subtotal
	// add VAT
	vat := order.VAT
	// add shipment
	shipping := order.ShippingCost

	// total
	total := subtotal + shipping + vat

	tplContext.Update(pongo2.Context{
		"subtotal":               subtotal,
		"vat":                    vat,
		"shipping":               shipping,
		"total":                  total,
		"countries":              countries.Countries,
		"DeliveryChargesExpress": KDeliveryChargesExpress,
		"order":                  order,
	})

	if err := tpl.ExecuteWriter(tplContext, w); err != nil {
		fmt.Printf("Tpl.ExecuteWriter Failed %v", err)
		e := ErrInternalServerError(err)
		render.Render(w, r, e)
		return
	}
}
