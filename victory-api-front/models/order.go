package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"bitbucket.org/softwarehouseio/victory/victory-frontend/config"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/models/transition"
	"github.com/jinzhu/gorm"
	"github.com/kjk/betterguid"
)

type Address struct {
	gorm.Model
	UserID       *uint
	ContactName  string
	Telephone    string
	City         string
	Country      string
	AddressLine1 string
	AddressLine2 string
	PostalCode   string
	Notes        string
}

func (address Address) Stringify() string {
	return fmt.Sprintf("%v, %v, %v", address.AddressLine2, address.AddressLine1, address.City)
}
func (a Address) Equals(na Address) bool {
	eq := true
	eq = eq && a.ContactName == na.ContactName
	eq = eq && a.Telephone == na.Telephone
	eq = eq && a.City == na.City
	eq = eq && a.Country == na.Country
	eq = eq && a.AddressLine1 == na.AddressLine1
	eq = eq && a.AddressLine2 == na.AddressLine2
	eq = eq && a.PostalCode == na.PostalCode
	eq = eq && a.Notes == na.Notes
	return eq
}
func (a Address) Update(na Address) {
	a.ContactName = na.ContactName
	a.Telephone = na.Telephone
	a.City = na.City
	a.Country = na.Country
	a.AddressLine1 = na.AddressLine1
	a.AddressLine2 = na.AddressLine2
	a.PostalCode = na.PostalCode
	a.Notes = na.Notes
}

// Order contains all user/shipping specific information
// required to fulfill the order or/and inform the user about
// the order status
type Order struct {
	gorm.Model
	Reference         string // used by people and the payment gateway
	UserID            *uint
	User              User
	Email             string // either the users email or an email that was provided during checkout
	PaymentAmount     float32
	AbandonedReason   string
	DiscountValue     uint
	TrackingNumber    *string
	ShippedAt         *time.Time
	ReturnedAt        *time.Time
	CancelledAt       *time.Time
	ShippingAddressID uint
	ShippingAddress   Address
	//BillingAddressID  uint
	//BillingAddress    Address

	PaymentMethod  string
	ShippingMethod string

	// persist for later ...
	Subtotal     float64
	VAT          float64
	ShippingCost float64
	Total        float64

	Notes      string
	OrderItems []OrderItem
	transition.Transition
}

func (o *Order) URL() string {
	orderReferenceURL := fmt.Sprintf("https://%v/cart/checkout/order-received/%v/",
		config.Config.DomainName, o.Reference)
	return orderReferenceURL
}
func (o *Order) ThumbnailURL(relativeURL string) string {
	orderReferenceURL := fmt.Sprintf("https://%v%v",
		config.Config.DomainName, relativeURL)
	return orderReferenceURL
}
func (o *Order) GetOrderReceiptURL() string {
	orderReferenceURL := fmt.Sprintf("https://%v/r/%v/",
		config.Config.DomainName, o.Reference)
	return orderReferenceURL
}

// ContinueShoppingURL - redirects the user to the current domain and shop
func (o *Order) ContinueShoppingURL() string {
	shopURL := fmt.Sprintf("https://%v/shop/any/", config.Config.DomainName)
	return shopURL
}
func (o *Order) GetOrderStateURL() string {
	stateURL := fmt.Sprintf("/cart/checkout/order-state/%v/", o.Reference)
	return stateURL
}

// GeneratePaymentReference - add a unique human readable reference
func (o Order) GeneratePaymentReference() string {
	orderID := o.ID
	guid := betterguid.New()[1:5]
	// even if the order id is 0 ensure that this will be unique
	ref := fmt.Sprintf("%v-%v-%v", time.Now().Format("20060102-3PM"), orderID, guid)
	return ref
}

// OrderItem is created when the user presses "Add to Card"
// on the shop detail page.
type OrderItem struct {
	gorm.Model
	OrderID            uint // back reference for Order
	ProductVariationID uint // product id for this order item
	ProductVariation   ProductVariation

	PersistProductDetailsJSON string // store the complete order for auditing

	CustomizedName   string
	CustomizedNumber uint

	Quantity     uint `cartitem:"Quantity"`
	Price        float64
	DiscountRate uint
	transition.Transition
}

func (itm *OrderItem) JSONString() (string, error) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(true)
	if err := enc.Encode(itm); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (order Order) Amount() (amount float64) {
	for _, orderItem := range order.OrderItems {
		amount += orderItem.Amount()
	}
	return
}

func (item OrderItem) Amount() float64 {
	return item.Price * float64(item.Quantity) * float64(100-item.DiscountRate) / 100
}

var (
	OrderState = transition.New(&Order{})
	ItemState  = transition.New(&OrderItem{})
)

func init() {
	// Define Order's States

	OrderState.Initial("draft")
	OrderState.State("checkout").Enter(func(value interface{}, tx *gorm.DB) error {
		err := tx.Save(value).Error
		// freeze stock, change items's state
		return err // will be nil if all went well
	})
	OrderState.State("cancelled").Enter(func(value interface{}, tx *gorm.DB) error {
		tx.Model(value).UpdateColumn("cancelled_at", time.Now())
		return nil
	})
	OrderState.State("paid").Enter(func(value interface{}, tx *gorm.DB) (err error) {
		var orderItems []OrderItem

		tx.Model(value).Association("OrderItems").Find(&orderItems)
		for _, item := range orderItems {
			if err = ItemState.Trigger("pay", &item, tx); err == nil {
				if err = tx.Select("state").Save(&item).Error; err != nil {
					return err
				}
			}
		}
		tx.Save(value)
		// freeze stock, change items's state
		return nil
	})
	OrderState.State("paid_cancelled").Enter(func(value interface{}, tx *gorm.DB) error {
		// do refund, release stock, change items's state
		return nil
	})
	OrderState.State("processing").Enter(func(value interface{}, tx *gorm.DB) (err error) {
		var orderItems []OrderItem
		tx.Model(value).Association("OrderItems").Find(&orderItems)
		for _, item := range orderItems {
			if err = ItemState.Trigger("process", &item, tx); err == nil {
				if err = tx.Select("state").Save(&item).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	OrderState.State("shipped").Enter(func(value interface{}, tx *gorm.DB) (err error) {
		tx.Model(value).UpdateColumn("shipped_at", time.Now())

		var orderItems []OrderItem
		tx.Model(value).Association("OrderItems").Find(&orderItems)
		for _, item := range orderItems {
			if err = ItemState.Trigger("ship", &item, tx); err == nil {
				if err = tx.Select("state").Save(&item).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	OrderState.State("returned")

	OrderState.Event("checkout").To("checkout").From("draft")
	OrderState.Event("pay").To("paid").From("checkout")
	cancelEvent := OrderState.Event("cancel")
	cancelEvent.To("cancelled").From("draft", "checkout")
	cancelEvent.To("paid_cancelled").From("paid", "processing", "shipped")
	OrderState.Event("process").To("processing").From("paid")
	OrderState.Event("ship").To("shipped").From("processing")
	OrderState.Event("return").To("returned").From("shipped")

	// Define ItemItem's States
	ItemState.Initial("checkout")
	ItemState.State("cancelled").Enter(func(value interface{}, tx *gorm.DB) error {
		// release stock, upate order state
		return nil
	})
	ItemState.State("paid").Enter(func(value interface{}, tx *gorm.DB) error {
		// freeze stock, update order state
		return nil
	})
	ItemState.State("paid_cancelled").Enter(func(value interface{}, tx *gorm.DB) error {
		// do refund, release stock, update order state
		return nil
	})
	ItemState.State("processing")
	ItemState.State("shipped")
	ItemState.State("returned")

	ItemState.Event("checkout").To("checkout").From("draft")
	ItemState.Event("pay").To("paid").From("checkout")
	cancelItemEvent := ItemState.Event("cancel")
	cancelItemEvent.To("cancelled").From("checkout")
	cancelItemEvent.To("paid_cancelled").From("paid")
	ItemState.Event("process").To("processing").From("paid")
	ItemState.Event("ship").To("shipped").From("processing")
	ItemState.Event("return").To("returned").From("shipped")
}
