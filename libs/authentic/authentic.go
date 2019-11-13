package authentic

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/alexedwards/scs"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

type AuthResource struct {
	BasePath             string // the base path to be allowed, defaults to "/"
	SessionObjectKeyName string // the name of the session object
	SuccessPath          string // the path to redirect to when the credentials are correct
	SessionManager       *scs.Manager
}

func (a AuthResource) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(render.SetContentType(render.ContentTypeHTML))
	//r.Get("/", a.Show)
	//r.Post("/email/", a.Process)
	r.Get("/email/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, a.BasePath, http.StatusSeeOther)
	})
	return r
}

func (a AuthResource) NewAuthResource(basePath string, sessionObjectKeyName string,
	successPath string,
	sessionManager *scs.Manager) *AuthResource {
	na := &AuthResource{
		BasePath:             basePath,
		SessionObjectKeyName: sessionObjectKeyName,
		SuccessPath:          successPath,
		SessionManager:       sessionManager,
	}

	return na
}

/*
	This SessionContext is to be used in chi middleware like this:

	m.Use(SessionContext(func(userID, userObject) {
		sql.Select - where userID == userID ...
		userObject = result ...
		return true
	} bool))

*/

func (a AuthResource) SessionContext(requireSessionObject bool, findInDb func(*http.Request, string) (interface{}, bool)) func(http.Handler) http.Handler {
	fmt.Printf("AR.SessionContext being generated %v %v", requireSessionObject, findInDb)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			//fmt.Printf("AR.UserSessionContext %v - %#v\n", r.URL.Path)
			for _, co := range r.Cookies() {
				fmt.Printf("Cookie: %v: %v\n", co.Name, co.Value)
			}
			if a.SessionObjectKeyName == "" {
				panic("AuthResource.SessionObjectKeyName not set")
			}
			if a.BasePath == "" {
				panic("AuthResource.BasePath not set")
			}
			// add the user session to the context
			session := a.SessionManager.Load(r)
			var (
				pathIsAllowed       = strings.Index(r.URL.Path, a.BasePath) == 0
				sessionKeyExists, _ = session.Exists(a.SessionObjectKeyName)
			)

			if requireSessionObject {
				// either session key doesnt exist
				if !sessionKeyExists {
					// we have to allow access an allowed endpoint
					if pathIsAllowed {
						next.ServeHTTP(w, r)
						return
					}
					// or we redirect to a.BasePath
					http.Redirect(w, r, a.BasePath, http.StatusSeeOther)
					return
				}
			}

			// session key exists, lets grab the object

			/*
				Instead of the actual object we're retrieving interface and passing interface.
				I believe this will allow me to factor out this bit of code into its own lib.
			*/
			ctx := r.Context()

			if userID, err := session.GetString(a.SessionObjectKeyName); err != nil {
				session.Destroy(w)
				if requireSessionObject {
					http.Redirect(w, r, a.BasePath, http.StatusSeeOther)
				} else {
					next.ServeHTTP(w, r)
				}
				return
			} else if gotUser, ok := findInDb(r, userID); !ok {
				fmt.Errorf("failed to fetch user %v by id\n", userID)
				session.Destroy(w)
				if requireSessionObject {
					http.Redirect(w, r, a.BasePath, http.StatusSeeOther)
				} else {
					next.ServeHTTP(w, r)
				}
				return
			} else {
				fmt.Printf("Authentic.ChiMiddleware: userID [%v] findInDB [%#v]\n", userID, gotUser)
				ctx = context.WithValue(r.Context(), a.SessionObjectKeyName, gotUser)
			}

			// make and attach enhanced session
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (a AuthResource) GetUserSession(r *http.Request) interface{} {
	return r.Context().Value(a.SessionObjectKeyName)
	// the casting happens somewhere else
	//user, ok := userObj.(*models.User)
	//if !ok {
	//	return nil
	//}
	//return user
}

func (a AuthResource) SetUserSession(w http.ResponseWriter, r *http.Request, sessionObjectId string) error {
	session := a.SessionManager.Load(r)
	if err := session.PutString(w, a.SessionObjectKeyName, sessionObjectId); err != nil {
		return fmt.Errorf("trouble setting the session cookie")
	}
	return nil
}

func (a AuthResource) DestroyUserSession(w http.ResponseWriter, r *http.Request) {
	session := a.SessionManager.Load(r)
	session.Destroy(w)
}
