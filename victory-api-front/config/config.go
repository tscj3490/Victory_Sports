package config

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/jinzhu/configor"
)

const (
	OO_ENV_ROOT_DIR                = "OO_ENV_ROOT_DIR"
	OO_ENV_PUBLIC_DIR              = "OO_ENV_PUBLIC_DIR"
	OO_ENV_WORK_DIR                = "OO_ENV_WORK_DIR"
	OO_ENV_VERSION                 = "OO_ENV_VERSION"
	OO_ENV_SENDGRID                = "OO_ENV_SENDGRID"
	OO_ENV_SENTRYDSN               = "OO_ENV_SENTRYDSN"
	OO_ENV_SENTRYENV               = "OO_ENV_SENTRYENV"
	GATEWAY_MERCHANT_ID            = "GATEWAY_MERCHANT_ID"
	GATEWAY_MASTERCARD_SECRET      = "GATEWAY_MASTERCARD_SECRET"
	GATEWAY_MERCHANT_NAME          = "GATEWAY_MERCHANT_NAME"
	GATEWAY_MERCHANT_ADDRESS_LINE1 = "GATEWAY_MERCHANT_ADDRESS_LINE1"
	GATEWAY_MERCHANT_ADDRESS_LINE2 = "GATEWAY_MERCHANT_ADDRESS_LINE2"
	AppName                        = "victory-frontend"
	AppRepository                  = "bitbucket.org/softwarehouseio/victory/victory-frontend/"
)

// random defaults
const (
	DefaultBadgePrice = float64(10) // 10 AED
)

var (
	Root                 = ENVRootDir()
	AccessTokens         = []string{"0c0200529a7c488c3a3d9032d533ce73"}
	baseEmailTemplateDir = "templates/emails/"
	EmailTemplates       = map[string]string{
		"customer-completed-order":  fmt.Sprintf("%v%v.html", baseEmailTemplateDir, "customer-completed-order"),
		"customer-note":             fmt.Sprintf("%v%v.html", baseEmailTemplateDir, "customer-note"),
		"customer-on-hold-order":    fmt.Sprintf("%v%v.html", baseEmailTemplateDir, "customer-on-hold-order"),
		"customer-processing-order": fmt.Sprintf("%v%v.html", baseEmailTemplateDir, "customer-processing-order"),
		"customer-refund-order":     fmt.Sprintf("%v%v.html", baseEmailTemplateDir, "customer-refund-order"),
	}
	SendGridAPIEndpoint = "/v3/mail/send"
	SendGridAPIDomain   = "https://api.sendgrid.com"
	menuLinks           = map[string]func(params []interface{}) string{
		"home":         func(params []interface{}) string { return "/" },
		"shop":         func(params []interface{}) string { return fmt.Sprintf("/shop/%v/", params...) },
		"stats":        func(params []interface{}) string { return "/stats/" },
		"cart":         func(params []interface{}) string { return "/cart/" },
		"user_auth":    func(params []interface{}) string { return fmt.Sprintf("/user/auth/%v", params...) },
		"user_profile": func(params []interface{}) string { return "/user/profile/" },
		"language":     func(params []interface{}) string { return fmt.Sprintf("/%v/", params...) },
	}
)

func MenuLinks(key string, p string) string {
	params := []interface{}{}
	for _, prm := range strings.Split(p, ",") {
		params = append(params, prm)
	}
	resp := ""
	if cb, ok := menuLinks[key]; ok {
		resp = cb(params)
	}
	return resp
}

var Config = struct {
	Port                  uint    `default:"8080" env:"PORT"`
	DBArgs                string  `default:"victory-frontend.db" env:"DBARGS"`
	DBDialect             string  `default:"sqlite3" env:"DBDIALECT"`
	TokenHeaderName       string  `default:"key"`
	SessionManagerKey     string  `default:"2c8a3d885b25f0f7e1b8eaeb2c1cbdca"`
	VAT                   float64 `default:"0.05"`
	ShippingUAE           uint    `default:"0"`                        // in aed
	ShippingExpressUAE    uint    `default:"3"`                        // in aed
	ShippingInternational uint    `default:"100"`                      // in aed
	DomainName            string  `default:"victory.softwarehouse.io"` // the domain used for all links sent via email
	EmailReplyAddress     string  `default:"hello@victory.softwarehouse.io"`
	EmailReplyName        string  `default:"Victory Store"`
	TestDBArgs            string  `default:"victory-frontend_test.db" env:"DBARGS"`
	CacheDBArgs           string  `default:"go-cache-boltdb.db" env:"CACHEDBARGS"`
}{}

func init() {
	if err := configor.Load(&Config); err != nil {
		panic(err)
	}
	Config.DBArgs = filepath.Join(ENVWorkDir(), "victory-frontend.db")          //ENVWorkDir() + "/victory.db"
	Config.TestDBArgs = filepath.Join(ENVWorkDir(), "victory-frontend_test.db") //ENVWorkDir() + "/victory.db"
	Config.CacheDBArgs = filepath.Join(ENVWorkDir(), "victory-go-cache-boltdb.db")

	// remapping db path to be in the work dir
	fmt.Printf(
		"config loaded with: Root: %s, ENVWorkDir: %s, ENVPublicDir: %s, ENVVersion: %s, ENVFirebaseConfig: %s, ENVFirebaseProjectID: %s",
		Root, ENVWorkDir(), ENVPublicDir(), ENVVersion(), ENVFirebaseConfig(), ENVFirebaseProjectID())
}

func ENVFirebaseProjectID() string {
	gcloudProjectID := os.Getenv("GCLOUD_PROJECT")
	if gcloudProjectID == "" {
		fmt.Println("Error initializing Firebase Admin SDK: Missing FIREBASE_CONFIG or GCLOUD_PROJECT variable")
		return string("")
	} else {
		return gcloudProjectID
	}
}

func ENVFirebaseConfig() string {
	// verify that the FIREBASE_CONFIG is set in ENV
	// FIREBASE_CONFIG can either hold the config in full as json blob or be a pointer to a file
	hasFirebaseConfig := os.Getenv("FIREBASE_CONFIG")

	if hasFirebaseConfig == "" {
		fmt.Println("Error initializing Firebase Admin SDK: Missing FIREBASE_CONFIG or GCLOUD_PROJECT variable")
		return string("")
	} else {
		return hasFirebaseConfig
	}
}

func ENVRootDir() string {
	rootDir := os.Getenv(OO_ENV_ROOT_DIR)
	if rootDir == "" {
		rootDir = os.Getenv("GOPATH") + "/src/" + AppRepository
	}
	return rootDir
}

func ENVPublicDir() string {
	publicDir := os.Getenv(OO_ENV_PUBLIC_DIR)
	if publicDir == "" {
		publicDir = filepath.Join(Root, "public")
	}
	return publicDir
}
func ENVWorkDir() string {
	publicDir := os.Getenv(OO_ENV_WORK_DIR)
	if publicDir == "" {
		publicDir = fmt.Sprintf("%v/%v", os.TempDir(), AppName)
	}
	return publicDir
}
func ENVVersion() string {
	version := os.Getenv(OO_ENV_VERSION)
	if version == "" {
		version = "dev-1234" // usually version should be a git commit id
	}
	return version
}
func ENVSendgrid() string {
	version := os.Getenv(OO_ENV_SENDGRID)
	//if version == "" {
	//	version = "nothing" // usually version should be a git commit id
	//}
	return version
}

const ENVFileUploadURI = "/uploads/"

func ENVFileUploadDir() string {
	return path.Join(ENVWorkDir(), "uploads")
}

type ENVPaymentGateway struct {
	MerchantID           string
	MerchantName         string
	Secret               string
	MerchantAddressLine1 string
	MerchantAddressLine2 string
}

func ENVGetPaymentGateway() ENVPaymentGateway {
	gateway := ENVPaymentGateway{}

	gateway.Secret = os.Getenv(GATEWAY_MASTERCARD_SECRET)
	gateway.MerchantID = os.Getenv(GATEWAY_MERCHANT_ID)
	gateway.MerchantName = os.Getenv(GATEWAY_MERCHANT_NAME)

	gateway.MerchantAddressLine1 = os.Getenv(GATEWAY_MERCHANT_ADDRESS_LINE1)
	gateway.MerchantAddressLine2 = os.Getenv(GATEWAY_MERCHANT_ADDRESS_LINE2)

	return gateway
}

type ENVSentryConfig struct {
	DSN         string
	Environment string
	Release     string
}

func ENVSentry() ENVSentryConfig {
	dsn := "https://50c3c1e181ac40288c2c466ff4cc6132:eeeb40f77c2747cba99abac759bb644e@sentry.io/1258584"
	if d := os.Getenv(OO_ENV_SENTRYDSN); d != "" {
		dsn = d
	}
	env := "dev"
	if e := os.Getenv(OO_ENV_SENTRYENV); e != "" {
		env = e
	}
	sentryConfig := ENVSentryConfig{
		DSN:         dsn,
		Environment: env,
		Release:     ENVVersion(),
	}

	return sentryConfig
}
