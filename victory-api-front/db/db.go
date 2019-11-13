package db

import (
	"bytes"
	"context"
	"encoding/gob"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"bitbucket.org/softwarehouseio/victory/victory-frontend/config"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/libs/gosportmonks"
	"github.com/jinzhu/gorm"

	// the only way to import sqlite with gorm
	_ "github.com/jinzhu/gorm/dialects/sqlite"

	cache "github.com/patrickmn/go-cache"
	"github.com/pressly/chainstore"
	"github.com/pressly/chainstore/boltstore"
	"github.com/pressly/chainstore/lrumgr"
)

// ContextKey defined type used for context's key
type ContextKey string

// ContextDBName db name used for context
var ContextDBName ContextKey = "ContextDB"

var (
	DB            *gorm.DB
	TestDB        *gorm.DB
	CacheDB       chainstore.Store
	InMemoryCache *cache.Cache
)

const (
	CacheDefaultExpiration = 50 * time.Minute
	CacheCleanupInterval   = 10 * time.Minute
)

func init() {
	log.Printf("DB: %v", config.Config.DBArgs)
	var err error
	DB, err = gorm.Open(config.Config.DBDialect, config.Config.DBArgs)
	TestDB, err = gorm.Open(config.Config.DBDialect, config.Config.DBArgs)

	// this not only restores but also initializes
	// err = CacheDBRestore()

	// diskStore := lrumgr.New(500*1024*1024, // 500MB of working data
	// 	metricsmgr.New("chainstore.ex.bolt",
	// 	boltstore.New(config.Config.CacheDBArgs, "CacheDBStore"),
	// 	),
	// )
	diskStore := lrumgr.New(500*1024*1024, // 500MB of working data
		boltstore.New(config.Config.CacheDBArgs, "CacheDBStore"),
	)
	CacheDB = chainstore.New(diskStore)
	gob.Register(map[string]cache.Item{})
	gob.Register(gosportmonks.League{})
	gob.Register([]gosportmonks.Fixture{})
	gob.Register([]gosportmonks.Team{})
	gob.Register([]gosportmonks.Topscorer{})
	gob.Register([]gosportmonks.Standing{})
	gob.Register([]gosportmonks.PlayerSquadStats{})
	CacheDBRestore()

	if err != nil {
		panic(err)
	}

	if os.Getenv("DEBUG") != "" {
		DB.LogMode(true)
	}

	// this hooks the special lib callbacks that
	// tweak the func callbacks needed to do the on the fly
	// translation of data
}

// CacheDBRestore load the persisted stored file and refil the cache
var CacheDBStoreKey string = "cachedbstore"

func CacheDBRestore() error {
	var err error
	err = CacheDB.Open()
	if err != nil {
		log.Printf("Open: Failed: %q", err)
	}
	defer CacheDB.Close()
	ctx := context.Background()

	var val []byte
	val, err = CacheDB.Get(ctx, CacheDBStoreKey)
	if err != nil {
		log.Printf("Put: Failed %q", err)
	}

	decodedItems := map[string]cache.Item{}
	buf := bytes.NewBuffer(val)
	decoder := gob.NewDecoder(buf)

	err = decoder.Decode(&decodedItems)
	if err != nil {
		log.Printf("CacheDBRestore failed: %v", err)
		InMemoryCache = cache.New(CacheDefaultExpiration, CacheCleanupInterval)
		// panic(err)
	}
	InMemoryCache = cache.NewFrom(
		CacheDefaultExpiration, CacheCleanupInterval, decodedItems)

	// var err error

	// // CacheDB, err = skv.Open(config.Config.CacheDBArgs)
	// defer CacheDB.Close()
	// items := map[string]cache.Item{}

	// err = CacheDB.Get(CacheDBStoreKey, &items)
	// if err != nil {
	// 	log.Printf("CacheDBRestore failed: %v", err)
	// 	InMemoryCache = cache.New(CacheDefaultExpiration, CacheCleanupInterval)
	// 	return nil
	// }
	// InMemoryCache = cache.NewFrom(CacheDefaultExpiration, CacheCleanupInterval, items)
	return nil
}

// CacheDBSave persist the current cached items into our CacheDB
func CacheDBSave(ctx context.Context) error {
	var err error
	err = CacheDB.Open()
	if err != nil {
		log.Printf("Open: Failed: %q", err)
	}
	defer CacheDB.Close()

	m := InMemoryCache.Items()
	buf := new(bytes.Buffer)
	encoder := gob.NewEncoder(buf)

	err = encoder.Encode(m)
	if err != nil {
		panic(err)
	}

	vals := buf.Bytes()

	err = CacheDB.Put(ctx, CacheDBStoreKey, vals)
	if err != nil {
		log.Printf("CacheDBSave-Put: Failed: %v, %q", CacheDBStoreKey, err)
	}
	// var err error
	// CacheDB, err = skv.Open(config.Config.CacheDBArgs)
	// defer CacheDB.Close()

	// items := InMemoryCache.Items()

	// err = CacheDB.Put(CacheDBStoreKey, items)
	// if err != nil {
	// 	log.Printf("CacheDBSave failed: %v", err)
	// 	return err
	// }
	// log.Printf("CacheDB Stored %v\n", items)

	// testItems := map[string]cache.Item{}
	// err = CacheDB.Get(CacheDBStoreKey, &testItems)
	// if err != nil {
	// 	log.Printf("CacheDBRestoreTest failed: %v", err)
	// }
	// log.Printf("CacheDB Stored Test %v\n", testItems)

	return nil
}

// GetDBFromRequestContext get database from request context
func GetDBFromRequestContext(req *http.Request) *gorm.DB {
	db := req.Context().Value(ContextDBName)
	if tx, ok := db.(*gorm.DB); ok {
		return tx
	}

	return nil
}

// GetDBFromContext get database from request context
func GetDBFromContext(context context.Context) *gorm.DB {
	db := context.Value(ContextDBName)
	if tx, ok := db.(*gorm.DB); ok {
		return tx
	}

	return nil
}

func DBRequestContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// this is useful to chain extra request depended functions onto the database

		// add the database as transaction which triggers a commit when the transaction finishes
		tx := DB.Begin()
		defer tx.Commit()
		ctx := context.WithValue(req.Context(), ContextDBName, tx)
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}
func Escape(value string) string {
	replace := map[string]string{"\\": "\\\\", "'": `''`, "\\0": "\\\\0", "\n": "\\n", "\r": "\\r", `"`: `\"`, "\x1a": "\\Z"}

	for b, a := range replace {
		value = strings.Replace(value, b, a, -1)
	}

	return value
}

/*
func Escape(source string) string {
	var j int = 0
	if len(source) == 0 {
		return ""
	}
	tempStr := source[:]
	desc := make([]byte, len(tempStr)*2)
	for i := 0; i < len(tempStr); i++ {
		flag := false
		var escape byte
		switch tempStr[i] {
		case '\r':
			flag = true
			escape = '\r'
			break
		case '\n':
			flag = true
			escape = '\n'
			break
		case '\\':
			flag = true
			escape = '\\'
			break
		case '\'':
			flag = true
			escape = '\''
			break
		case '"':
			flag = true
			escape = '"'
			break
		case '\032':
			flag = true
			escape = 'Z'
			break
		default:
		}
		if flag {
			desc[j] = '\\'
			desc[j+1] = escape
			j = j + 2
		} else {
			desc[j] = tempStr[i]
			j = j + 1
		}
	}
	return string(desc[0:j])
}
*/
