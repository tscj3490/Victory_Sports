package controllers

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"strings"

	"bitbucket.org/softwarehouseio/victory/victory-frontend/models"
	"github.com/flosch/pongo2"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"errors"

	"gopkg.in/dgrijalva/jwt-go.v3"

	"context"
	"log"

	"bitbucket.org/softwarehouseio/victory/victory-frontend/config"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/db"
	"github.com/jinzhu/gorm"
	"github.com/vincent-petithory/countries"
)

type UserResource struct {
	BaseURL string // usually "/user/"
}

func (ur UserResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Route("/auth/", func(a chi.Router) {
		a.Get("/signinSuccess", ur.SignInSuccess)
		a.Post("/persistSession", ur.PersistSession)
		a.Post("/signout", ur.SignOut)
		a.Get("/signin", ur.SignIn)
	})

	r.Route("/profile/", func(p chi.Router) {
		p.Use(ur.UserContext)
		p.Get("/", ur.ProfileView)
		p.Post("/edit", ur.ProfileEdit)
		p.Get("/history/", ur.ProfileHistory)
	})

	return r
}

func (ur UserResource) UserContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tplContext := WebResource{}.GetTplContext(r)

		var (
			user, _ = tplContext["user"].(*models.User)
		)
		noUserFound := user == nil

		if noUserFound {
			http.Redirect(w, r, "/user/auth/signin?from=profile", http.StatusSeeOther)
			return
		}
		tplContext.Update(pongo2.Context{
			"BaseURL": ur.BaseURL,
		})
		ctx := context.WithValue(r.Context(), TplContextWebResourceContextKey, tplContext)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (ur UserResource) GetUser(r *http.Request, userID string) (interface{}, bool) {
	var (
		tx = db.GetDBFromRequestContext(r)
	)
	log.Printf("UserRes.GetUser: [%v] \n", userID)

	var err error
	userObject, err := models.User{}.GetUser(tx, userID)
	log.Printf("UserRes.GetUser Found: %v\n", userObject)

	if err != nil {
		log.Printf("UserResource.GetUser failed: %v", err)
		return nil, false
	}
	return userObject, true
}
func (ur UserResource) SignOut(w http.ResponseWriter, r *http.Request) {
	authResource.DestroyUserSession(w, r)
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return
}
func (ur UserResource) SignIn(w http.ResponseWriter, r *http.Request) {
	var (
		tpl        = pongo2.Must(pongo2.FromFile("templates/html/user-signin.html"))
		tplContext = WebResource{}.GetTplContext(r)
	)
	fromQueryParam := r.URL.Query().Get("from")

	if fromQueryParam != "" {
		tplContext = tplContext.Update(pongo2.Context{
			"generateSessionURL": "/user/auth/persistSession" + fmt.Sprintf("?from=%v", fromQueryParam),
		})
	} else {
		tplContext = tplContext.Update(pongo2.Context{
			"generateSessionURL": "/user/auth/persistSession",
		})
	}

	if err := tpl.ExecuteWriter(tplContext, w); err != nil {
		log.Printf("Tpl.ExecuteWriter Failed %v", err)
		e := ErrInternalServerError(err)
		render.Render(w, r, e)
		return
	}
}

/**
 * @description: Render the SignInSuccess Page
 */
func (ur UserResource) SignInSuccess(w http.ResponseWriter, r *http.Request) {
	var (
		tpl        = pongo2.Must(pongo2.FromFile("templates/html/user-signin-success.html"))
		tplContext = WebResource{}.GetTplContext(r)
	)

	fromQueryParam := r.URL.Query().Get("from")

	if fromQueryParam != "" {
		tplContext = tplContext.Update(pongo2.Context{
			"generateSessionURL": "/user/auth/persistSession" + fmt.Sprintf("?from=%v", fromQueryParam),
		})
	} else {
		tplContext = tplContext.Update(pongo2.Context{
			"generateSessionURL": "/user/auth/persistSession",
		})
	}

	if err := tpl.ExecuteWriter(tplContext, w); err != nil {
		log.Printf("Tpl.ExecuteWriter Failed %v", err)
		e := ErrInternalServerError(err)
		render.Render(w, r, e)
		return
	}
}

/**
 * @description: Takes the firebase auth token and sets the session cookie to sign in the user.
 */
func (ur UserResource) PersistSession(w http.ResponseWriter, r *http.Request) {
	var (
		tx = db.GetDBFromRequestContext(r)
	)
	// 1) read the values and turn them into a struct
	r.ParseForm()

	// 2) check if we've seen this firebaseUserData before in the database
	//fmt.Println("hello %#v", r.PostForm)
	firebaseUserData := models.FirebaseUserData{
		UID:     r.PostForm.Get("uid"),
		Email:   r.PostForm.Get("email"),
		IdToken: r.PostForm.Get("idToken"),
	}
	log.Printf("UR.PersistSession: hm: %#v", firebaseUserData)
	log.Printf("UR.PersistSession: %v", r.PostForm)

	//var FirebaseApp *firebase.App
	//if app, err := firebase.NewApp(context.Background(), nil); err != nil {
	//	panic(fmt.Errorf("Error initializing Firebase Admin SDK: %v", err))
	//} else {
	//	FirebaseApp = app
	//}
	//client, err := config.FirebaseApp.Auth(context.Background())
	//if err != nil {
	//	log.Printf("error instantiating firebase auth\n", err)
	//	render.Render(w, r, ErrInternalServerError(err))
	//	return
	//}
	//token, err := client.VerifyIDToken(firebaseUserData.IdToken)
	//if err != nil {
	//	log.Printf("error verifying ID token: %v\n%v\n", err, token)
	//	render.Render(w, r, ErrInternalServerError(err))
	//	return
	//}
	//if firebaseUserData.FirebaseID != token.FirebaseID {
	//	log.Printf("error verifying ID token mismatch\n", err)
	//	render.Render(w, r, ErrInternalServerError(nil))
	//	return
	//}
	uID, err := VerifyIDToken(firebaseUserData.IdToken, config.ENVFirebaseProjectID())
	if err != nil {
		log.Printf("error verifying ID token mismatch: %v", err)
		e := ErrInternalServerError(err)
		render.Render(w, r, e)
		return
	}
	if firebaseUserData.UID != uID {
		log.Printf("firebaseUserData.FirebaseID doesnt match server FirebaseID\n")
		e := ErrInternalServerError(nil)
		render.Render(w, r, e)
		return
	}

	// 3) update / create firebaseUserData data in DB
	newUser := models.User{
		FirebaseID: firebaseUserData.UID,
		Email:      firebaseUserData.Email,
	}
	user := models.User{}
	// 3.1 search for an already existing user with the same firebaseID
	err = tx.Model(user).Where("firebase_id = ?", firebaseUserData.UID).Find(&user).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Printf("UserResource.PersistSession Err: %v", err)
		e := ErrInternalServerError(err)
		render.Render(w, r, e)
		return
	}
	if err == gorm.ErrRecordNotFound {
		// doesnt exist, lets create it
		if err := tx.Model(newUser).Create(&newUser).Error; err != nil {
			log.Printf("UserResource.PersistSession Create Err: %v", err)
			e := ErrInternalServerError(err)
			render.Render(w, r, e)
			return
		} else {
			user = newUser
		}
	}
	log.Printf("UserRes.PersistSes.User Details: [%v]", user)

	// 4) persist session by writing db id into session cookie store

	if err := authResource.SetUserSession(w, r, fmt.Sprintf("%v", user.ID)); err != nil {
		//http.Redirect(w, r, "/", http.StatusSeeOther)
		e := ErrInternalServerError(err)
		render.Render(w, r, e)
		return
	}

	// 5) redirect to either / or some from target

	fromQueryParam := r.URL.Query().Get("from")
	redirectTarget := "/"
	switch fromQueryParam {
	case "admin":
		redirectTarget = "/admin/"
	}

	http.Redirect(w, r, redirectTarget, http.StatusSeeOther)
	//render.Render(w, r, ErrInvalidRequest(fmt.Errorf("something ok... %v", uID)))

	return
}

func (ur UserResource) ProfileHistory(w http.ResponseWriter, r *http.Request) {
	var (
		tpl        = pongo2.Must(pongo2.FromFile("templates/html/user-profile-history.html"))
		tplContext = WebResource{}.GetTplContext(r)
		tx         = db.GetDBFromRequestContext(r)
		user, _    = authResource.GetUserSession(r).(*models.User)
		userID     = user.ID
	)

	orders := []models.Order{}

	if err := tx.Model(orders).Where("user_id = ?", userID).
		Preload("OrderItems.ProductVariation.Product").
		Find(&orders).Error; err != nil {
		log.Printf("UR.ProfileHistory")
	}

	tplContext = tplContext.Update(pongo2.Context{
		"BaseURL": ur.BaseURL,
		"orders":  orders,
	})

	if err := tpl.ExecuteWriter(tplContext, w); err != nil {
		log.Printf("Tpl.ExecuteWriter Failed %v", err)
		e := ErrInternalServerError(err)
		render.Render(w, r, e)
		return
	}
}

func (ur UserResource) ProfileView(w http.ResponseWriter, r *http.Request) {
	var (
		tpl        = pongo2.Must(pongo2.FromFile("templates/html/user-profile.html"))
		tplContext = WebResource{}.GetTplContext(r)
		tx         = db.GetDBFromRequestContext(r)
		user, _    = authResource.GetUserSession(r).(*models.User)
		address    = models.Address{}
	)

	userID := user.ID

	err := tx.Model(address).Where("user_id = ?", userID).First(&address).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Printf("UserRes.Profile View Address fetch failed: %v", err)
		// redirect back to the cart
		errMsg := fmt.Errorf("fetching the address resulted in an unexpected error please try again later")
		http.Redirect(w, r, fmt.Sprintf("%v#msg=%v", "/", errMsg), http.StatusSeeOther)
		return
	}

	tplContext = tplContext.Update(pongo2.Context{
		"BaseURL":   ur.BaseURL,
		"countries": countries.Countries,
		"address":   address,
	})

	if err := tpl.ExecuteWriter(tplContext, w); err != nil {
		log.Printf("Tpl.ExecuteWriter Failed %v", err)
		e := ErrInternalServerError(err)
		render.Render(w, r, e)
		return
	}
}

func (ur UserResource) ProfileEdit(w http.ResponseWriter, r *http.Request) {
	var (
		addressReq = AddressRequest{}
		profileURL = fmt.Sprintf("%vprofile/", ur.BaseURL)
		user, _    = authResource.GetUserSession(r).(*models.User)
		tx         = db.GetDBFromRequestContext(r)
	)

	if user == nil {
		// no user, exit!!!
		errMsg := fmt.Errorf("active user session required, please log in to edit your user data.")
		http.Redirect(w, r, fmt.Sprintf("%v#msg=%v", "/", errMsg), http.StatusSeeOther)
		return
	}

	r.ParseForm()

	if err := addressReq.BindForm(r); err != nil {
		// save the error
		log.Printf("UserRes.Profile Edit AddressRequest.bindForm failed: %v", err)
		// redirect back to the cart
		errMsg := fmt.Errorf("invalid data in post request - bind failed")
		http.Redirect(w, r, fmt.Sprintf("%v#msg=%v", profileURL, errMsg), http.StatusSeeOther)
		return
	}
	// all good, lets store it as address in the db

	address := models.Address{}
	userID := user.ID
	storageMethod := tx.Save

	err := tx.Model(address).Where("user_id = ?", userID).First(&address).Error
	if err == gorm.ErrRecordNotFound {
		storageMethod = tx.Create
	}

	addressReq.BindAddressModel(&address) // store data from address request
	address.UserID = &user.ID

	if err := storageMethod(&address).Error; err != nil {
		errMsg := fmt.Errorf("failed to save or create address in your profile update")
		http.Redirect(w, r, fmt.Sprintf("%v#msg=%v", profileURL, errMsg), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, profileURL, http.StatusSeeOther)
	return
}

const (
	clientCertURL = "https://www.googleapis.com/robot/v1/metadata/x509/securetoken@system.gserviceaccount.com"
)

func VerifyIDToken(idToken string, googleProjectID string) (string, error) {
	keys, err := fetchPublicKeys()

	if err != nil {
		return "", err
	}

	if idToken == "" {
		return "", fmt.Errorf("Verification Token Empty")
	}

	parsedToken, err := jwt.Parse(idToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		kid := token.Header["kid"]

		certPEM := string(*keys[kid.(string)])
		certPEM = strings.Replace(certPEM, "\\n", "\n", -1)
		certPEM = strings.Replace(certPEM, "\"", "", -1)
		block, _ := pem.Decode([]byte(certPEM))
		var cert *x509.Certificate
		cert, _ = x509.ParseCertificate(block.Bytes)
		rsaPublicKey := cert.PublicKey.(*rsa.PublicKey)

		return rsaPublicKey, nil
	})

	if err != nil {
		return "", err
	}

	errMessage := ""

	claims := parsedToken.Claims.(jwt.MapClaims)

	if claims["aud"].(string) != googleProjectID {
		errMessage = "Firebase Auth ID token has incorrect 'aud' claim: " + claims["aud"].(string)
	} else if claims["iss"].(string) != "https://securetoken.google.com/"+googleProjectID {
		errMessage = "Firebase Auth ID token has incorrect 'iss' claim"
	} else if claims["sub"].(string) == "" || len(claims["sub"].(string)) > 128 {
		errMessage = "Firebase Auth ID token has invalid 'sub' claim"
	}

	if errMessage != "" {
		return "", errors.New(errMessage)
	}

	return string(claims["sub"].(string)), nil
}

func fetchPublicKeys() (map[string]*json.RawMessage, error) {
	resp, err := http.DefaultClient.Get(clientCertURL)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var objmap map[string]*json.RawMessage
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&objmap)

	return objmap, err
}
