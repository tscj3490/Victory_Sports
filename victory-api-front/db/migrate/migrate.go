package migrate

import (
	"bitbucket.org/softwarehouseio/victory/victory-frontend/db"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/models"
	"bitbucket.org/softwarehouseio/victory/victory-frontend/models/stateChangeLog"
)

func autoMigrate(values ...interface{}) {
	for _, value := range values {
		db.DB.AutoMigrate(value)
	}
}

var Tables = []interface{}{
	&models.User{},
	&models.Brand{},
	&models.Category{},
	&models.Collection{},
	&models.League{},
	&models.Player{},
	&models.Product{},
	&models.ProductSize{},
	&models.ProductVariation{},
	&models.Team{},
	&models.Badge{},
	&models.Order{},
	&models.OrderItem{},
	&models.Address{},
	&stateChangeLog.StateChangeLog{},
	&models.StatsLeague{},
	&models.StatsTeam{},
	&models.StatsSeason{},
	&models.StatsStage{},
	&models.StatsRound{},
	&models.PageContent{},
}

func DoAutoMigrate() {
	for _, table := range Tables {
		autoMigrate(table)
	}
}
