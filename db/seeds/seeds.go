package main

import (
	"path/filepath"
	"github.com/jinzhu/configor"
	"log"
)

var Seeds = struct {
	Users []struct {
		Email string
		FirebaseID string
		UserAccessLevel int64
	}
	Collections []struct {
		Name string
	}
	Leagues []struct {
		Name map[string]string
		StatsLeagueID int
	}
	Teams []struct {
		Name map[string]string
		Collections []struct {
			Name string
		}
		Logo string
		BrandName string
	}
	Players []struct {
		Name string
		TeamID int
		Collections []struct {
			Name string
		}
	}
	Brands []struct{
		Name string
	}
	Categories []struct{
		Name string
	}
	Products []struct{
		Name string
		Description string
		Price float64
		Thumbnail string
		Image string
		Gender string // "youth", "unisex", "male", "female"
		KitCode string

		CategoryID int //
		CategoryName string // options are: t-shirt,pants,shoes,merchandise

		Collections []struct {
			Name string
		}
		Sizes []struct {
			Name string
		}

		LeagueName int // possible league reference
		TeamName string // possible team reference
		PlayerName int // possible player reference

		BrandName string // denormalized

		Badges []struct {
			Name string
		}
	}
	ProductSizes []struct {
		Name string
	}
	Badges []struct {
		Name string
		Image string
		Thumbnail string
		Price float64
	}
	PageContent []struct {
		Page 					   string
		Identifier 				   string
		Text map[string]string
		Link 					   string
	}
}{}

func init() {
	filepaths, _ := filepath.Glob("db/seeds/data/*.yml")

	if err := configor.New(&configor.Config{Debug:false, Verbose: false}).Load(&Seeds, filepaths...); err != nil {
		log.Printf("What? %v", err)
		panic(err)
	}
}
