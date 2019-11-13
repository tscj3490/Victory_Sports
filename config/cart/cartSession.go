package cart

import (
	"github.com/alexedwards/scs"
	"net/http"
)

const CartSessionKey = "__meta_vic_cart"
type CartSessionStorage struct {
	Session *scs.Session
	ResponseWriter http.ResponseWriter
}

func (cs CartSessionStorage) Restore() (map[uint]*CartItem, error) {
	var list map[uint]*CartItem
	list = make(map[uint]*CartItem)

	if err := cs.Session.GetObject(CartSessionKey, &list); err != nil {
		list = make(map[uint]*CartItem)
		return list, err
	}
	return list, nil
}
func (cs CartSessionStorage) Save(data map[uint]*CartItem) error {
	return cs.Session.PutObject(cs.ResponseWriter, CartSessionKey, data)
}

func GetCart(w http.ResponseWriter, session *scs.Session) (*Cart, error) {
	storage := CartSessionStorage{session, w}
	restored, _ := storage.Restore()
	bucket := &Cart{
		CartItems: restored,
		storage: storage,
	}
	return bucket, nil
}