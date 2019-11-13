// +build ignore

package main

import (
	"bitbucket.org/softwarehouseio/victory/victory-frontend/config/i18n/l10n"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/db"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/db/migrate"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/models"
	"fmt"
	"os"
	"strings"
)

/*
How to run this:

go run db/seeds/main.go db/seeds/seeds.go
*/

var (
	Root, _ = os.Getwd()
	DraftDB = db.DB
)

var (
	AdminUser        *models.User
	ManyToManyTables = []interface{}{
		"player_collections",
		"product_collections",
		"team_collections",
		"product_productsizes",
	}
)

func main() {
	fmt.Print("Starting Seeds \nTruncating Tables \n")
	DropTables(ManyToManyTables...)
	TruncateTables(migrate.Tables...)
	createRecords()
}

func TruncateTables(tables ...interface{}) {
	for _, table := range tables {
		if err := DraftDB.DropTableIfExists(table).Error; err != nil {
			panic(err)
		}

		DraftDB.AutoMigrate(table)
	}
}
func DropTables(tables ...interface{}) {
	for _, table := range tables {
		if err := DraftDB.DropTableIfExists(table).Error; err != nil {
			panic(err)
		}
	}
}

func createRecords() {
	fmt.Println("Start create sample data...")
	//createAdminUsers()
	createUsers()
	fmt.Println("--> Created users.")
	createCollections()
	fmt.Println("--> Created collections.")
	createBrands()
	fmt.Println("--> Created brands.")
	createLeagues()
	fmt.Println("--> Created leagues.")
	createTeams()
	fmt.Println("--> Created teams.")
	createCategories()
	fmt.Println("--> Created categories.")
	createSizes()
	fmt.Println("--> Created product sizes")
	createBadges()
	fmt.Println("--> Created badges")
	//DraftDB.LogMode(true)
	createProducts()
	fmt.Println("--> Created products")
	createPageContentObjects()
	fmt.Println("--> Created page content")
}

func createAdminUsers() {
	AdminUser = &models.User{
		Email:           "albsen@gmail.com",
		UserAccessLevel: models.DefaultAdminLevel,
	}
	DraftDB.Create(AdminUser)
}
func createUsers() {
	for _, u := range Seeds.Users {
		user := models.User{}
		user.Email = u.Email
		user.FirebaseID = u.FirebaseID
		user.UserAccessLevel = u.UserAccessLevel
		fmt.Printf("Creating user: %v", u)
		if err := DraftDB.Create(&user).Error; err != nil {
			panic(err)
		}
	}
}
func createCollections() {
	for _, c := range Seeds.Collections {
		collection := models.Collection{}
		collection.Name = c.Name
		collection.Code = strings.ToLower(c.Name)
		collection.Code = strings.Replace(collection.Code, " ", "", 999)
		collection.Code = strings.Replace(collection.Code, ",", "", 999)
		fmt.Printf("creating collection: %v Name: %v\n", collection.Code, c.Name)
		if err := DraftDB.Create(&collection).Error; err != nil {
			panic(err)
		}
	}
}
func createBrands() {
	for _, b := range Seeds.Brands {
		brand := models.Brand{
			Name: b.Name,
		}
		if err := DraftDB.Create(&brand).Error; err != nil {
			panic(err)
		}
	}
}
func createLeagues() {
	for _, b := range Seeds.Leagues {
		league := models.League{
			Name:          l10n.SetAll(b.Name),
			StatsLeagueID: b.StatsLeagueID,
		}
		if err := DraftDB.Create(&league).Error; err != nil {
			panic(err)
		}
	}
}
func createTeams() {
	for _, t := range Seeds.Teams {

		team := models.Team{}
		team.Name = l10n.SetAll(t.Name)
		team.Logo = t.Logo
		if t.BrandName != "" {
			if b := findBrandByName(t.BrandName); b != nil {
				team.BrandID = b.ID
			} else {
				panic(fmt.Errorf("please add brand with name: %v", t.BrandName))
			}
		}

		for _, c := range t.Collections {
			collection := findCollectionByName(c.Name)
			if collection == nil {
				panic(fmt.Errorf("please add Collection with Name: %v to the Collection yaml data seeds.", c.Name))
			}
			team.Collections = append(team.Collections, *collection)
		}
		if err := DraftDB.Create(&team).Error; err != nil {
			panic(err)
		}
	}
}
func createCategories() {
	for _, cat := range Seeds.Categories {
		category := models.Category{
			Name: cat.Name,
		}
		category.Code = strings.ToLower(cat.Name)
		category.Code = strings.Replace(category.Code, " ", "", 0)
		if err := DraftDB.Create(&category).Error; err != nil {
			panic(err)
		}
	}
}
func createSizes() {
	for _, s := range Seeds.ProductSizes {
		size := models.ProductSize{Name: s.Name}
		if err := DraftDB.Create(&size).Error; err != nil {
			panic(err)
		}
	}
}
func createProducts() {
	for _, p := range Seeds.Products {
		product := &models.Product{
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Thumbnail:   p.Thumbnail,
			Image:       p.Image,
			Gender:      p.Gender,
			KitCode:     p.KitCode,
			Variations:  []models.ProductVariation{},
			Image2:       p.Image,
			Image3:       p.Image,
			Image4:       p.Image,
		}

		category := findCategoryByName(p.CategoryName)
		if category == nil {
			panic(fmt.Errorf("createProducts failed, category with name: %v missing. Please ensure that it is being created.", p.CategoryName))
		}
		product.CategoryID = category.ID
		for _, c := range p.Collections {
			collection := findCollectionByName(c.Name)
			if collection == nil {
				panic(fmt.Errorf("createProducts failed, please add Collection with Name: %v to the Collection yaml data seeds.", c.Name))
			}
			product.Collections = append(product.Collections, *collection)
		}
		for _, s := range p.Sizes {
			size := findProductSizeByName(s.Name)
			if size == nil {
				panic(fmt.Errorf("createProducts failed, please add product size with Name: %v to the products yaml data seeds.", s.Name))
			}
			product.Sizes = append(product.Sizes, *size)
		}
		if p.BrandName != "" {
			brand := findBrandByName(p.BrandName)
			if brand == nil {
				panic(fmt.Errorf("createProducts failed, brand with name: %v missing. Please ensure that it is being created.", p.BrandName))
			}
			product.BrandID = brand.ID
		}
		if p.TeamName != "" {
			team := findTeamByName(p.TeamName)
			if team == nil {
				panic(fmt.Errorf("createProducts failed, team with name: %v missing. Please ensure that it is being created.", p.TeamName))
			}
			product.TeamID = team.ID
		}
		// creating all product variations with some fake stock data
		// for each size and badge one variation
		/*

			ProductID *uint
			Product Product

			BadgeID *uint
			Badge Badge
			SizeID uint
			Size ProductSize

			SKU string
			Featured bool
			Price uint
			AvailableQuantity uint
		*/

		if err := DraftDB.Create(&product).Error; err != nil {
			panic(err)
		}

		for _, s := range product.Sizes {
			// always a variation without badges
			pvs := models.ProductVariation{}
			pvs.Product = *product
			pvs.AvailableQuantity = 7
			//pvs.Featured = false
			pvs.Size = s
			//pvs.Price = product.Price
			pvs.SKU = fmt.Sprintf("%v-%v", product.Name, s.Name)

			if err := DraftDB.Debug().Create(&pvs).Error; err != nil {
				panic(err)
			}

			// add CustomPrint
			pvc := models.ProductVariation{}
			pvc.Product = *product
			pvc.AvailableQuantity = 7
			pvc.Size = s
			pvc.CustomPrint = true
			pvc.SKU = fmt.Sprintf("%v-%v-%v", product.Name, s.Name, "CustomPrint")
			if err := DraftDB.Debug().Create(&pvc).Error; err != nil {
				panic(err)
			}

			for _, b := range p.Badges {
				if b.Name == "" {
					// exit early if no badge is available
					continue
				}
				badge := findBadgeByName(b.Name)
				if badge == nil {
					continue
				}
				pv := models.ProductVariation{}
				pv.Product = *product
				pv.AvailableQuantity = 7
				pv.Size = s
				pv.Badge = *badge
				pv.SKU = fmt.Sprintf("%v-%v-%v", product.Name, s.Name, badge.Name)
				if err := DraftDB.Debug().Create(&pv).Error; err != nil {
					panic(err)
				}

				pvcc := models.ProductVariation{}
				pvcc.Product = *product
				pvcc.AvailableQuantity = 7
				pvcc.Size = s
				pvcc.Badge = *badge
				pvcc.CustomPrint = true
				pvcc.SKU = fmt.Sprintf("%v-%v-%v-%v", product.Name, s.Name, badge.Name, "CustomPrint")
				if err := DraftDB.Debug().Create(&pvcc).Error; err != nil {
					panic(err)
				}
			}
		}
	}
}

func createBadges() {
	for _, b := range Seeds.Badges {
		badge := models.Badge{
			Name:      b.Name,
			Price:     b.Price,
			Thumbnail: b.Thumbnail,
			Image:     b.Image,
		}
		if err := DraftDB.Create(&badge).Error; err != nil {
			panic(err)
		}
	}
}

func createPageContentObjects() {
	for _, c := range Seeds.PageContent {
		pc := models.PageContent{
			Page: c.Page,
			Identifier: c.Identifier,
			Link: c.Link,
			Text: l10n.SetAll(c.Text),
		}
		if err := DraftDB.Create(&pc).Error; err != nil {
			panic(err)
		}
	}
}

// MARK: helper functions

func findCategoryByName(name string) *models.Category {
	category := &models.Category{}
	if err := DraftDB.Where(&models.Category{Name: name}).First(category).Error; err != nil {
		fmt.Errorf("can't find category with name = %q, got err %v", name, err)
	}
	if category.Name == "" {
		return nil
	}
	return category
}

func findCollectionByName(name string) *models.Collection {
	collection := &models.Collection{}
	if err := DraftDB.Where(&models.Collection{Name: name}).First(collection).Error; err != nil {
		fmt.Errorf("can't find collection with name = %q, got err %v", name, err)
	}
	if collection.Name == "" {
		return nil
	}
	return collection
}

func findBrandByName(name string) *models.Brand {

	brand := &models.Brand{}
	if err := DraftDB.Where(&models.Brand{Name: name}).First(brand).Error; err != nil {
		fmt.Errorf("can't find brand with name = %q, got err %v", name, err)
	}
	if brand.Name == "" {
		return nil
	}
	return brand
}

func findProductSizeByName(name string) *models.ProductSize {
	size := &models.ProductSize{}
	if err := DraftDB.Where(&models.ProductSize{Name: name}).First(size).Error; err != nil {
		fmt.Errorf("cant't find product size with name = %s, got err %s", name, err)
	}
	if size.Name == "" {
		return nil
	}
	return size
}
func findTeamByName(name string) *models.Team {
	team := &models.Team{}
	if err := DraftDB.Where("name LIKE ?", fmt.Sprintf("%%%v%%", name)).First(team).Error; err != nil {
		fmt.Errorf("cant't find team with name = %s, got err %s", name, err)
	}
	if team.Name == "" {
		return nil
	}
	return team
}
func findBadgeByName(name string) *models.Badge {
	badge := &models.Badge{}
	if err := DraftDB.Where(&models.Badge{Name: name}).First(badge).Error; err != nil {
		fmt.Errorf("cant't find badge with name = %s, got err %s", name, err)
	}
	if badge.Name == "" {
		return nil
	}
	return badge
}
