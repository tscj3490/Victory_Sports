package controllers

import (
	"encoding/json"
	"testing"
	"time"

	"bitbucket.org/softwarehouseio/victory/victory-frontend/models"
)

const (
	APIToken              = "0c51b581-0b7f-4896-a730-13470102d88a"
	DemoRateID            = "5f957cd4-b16f-4764-a0bd-9cc5e856efc5"
	DemoLabelID           = "cbe23b71-936e-4ac6-8686-8e155be4dbf9"
	DemoShipperAccountID1 = "a4897019-9d0c-4260-bd2e-3a316edcf1a8"
	DemoShipperAccountID2 = "a9e33d17-28a1-437b-9d60-8c474a355f13"
	DemoShipperAccountID3 = "07b55d06-48af-4b02-a6b0-1e311e22b1e6"
	DemoBulkDownloadID    = "a913396e-7df2-46ff-86b2-49372e21d227"
)

var calcRates = &models.CalculateRatesRequest{
	Async: false,
	ShipperAccounts: []models.Reference{
		{
			ID: "678f7d4a-03a1-4d3c-9cd5-d07ab1315f74",
		},
	},
	IsDocument: false,
	Shipment: models.Shipment{
		ShipFrom: models.ShipFromAddress{
			ContactName: "han tig",
			CompanyName: "bigcity",
			Street1:     "Shenyang, Liaoning, China",
			City:        "Shenyang",
			State:       "Liaoning",
			PostalCode:  "015000",
			Country:     "China",
			Type:        "business",
		},
		ShipTo: models.ShipToAddress{
			ContactName: "Dr. Moises Corwin",
			Phone:       "1-140-225-6410",
			Email:       "Giovanna42@yahoo.com",
			Street1:     "28292 Daugherty Orchard",
			City:        "Amman",
			PostalCode:  "560034",
			Country:     "JOR",
			Type:        "residential",
		},
		Parcels: []models.Parcel{
			{
				Description: "Food XS",
				BoxType:     "custom",
				Weight: models.Weight{
					Value: 2,
					Unit:  "kg",
				},
				Dimension: models.Dimension{
					Width:  20,
					Height: 40,
					Depth:  40,
					Unit:   "cm",
				},
				Items: []models.Item{
					{
						Description:   "Food Bar",
						OriginCountry: "USA",
						Quantity:      2,
						Price: models.Money{
							Amount:   3,
							Currency: "USD",
						},
						Weight: models.Weight{
							Value: 0.6,
							Unit:  "kg",
						},
						Sku: "imac2014",
					},
				},
			},
		},
	},
}

var searchDataForRate = models.ListAllRatesRequest{
	Status:       "calculated",
	Limit:        "20",
	NextToken:    "1",
	CreatedAtMax: "2018-09-19T04:12:11+00:00",
	CreatedAtMin: "2018-10-08T04:12:10+00:00",
}

var labels = &models.CreateLabelRequest{
	Async: false,
	Billing: &models.Billing{
		PaidBy: "shipper",
	},
	Customs: &models.Customs{
		Billing: &models.CustomsBilling{
			PaidBy: "recipient",
		},
		Purpose: "merchandise",
	},
	ReturnShipment: false,
	IsDocument:     false,
	ServiceType:    "aramex_priority_express",
	PaperSize:      "default",
	ShipperAccount: models.Reference{
		ID: DemoShipperAccountID3,
	},
	References: []string{"Handle with care"},
	Shipment: models.Shipment{
		ShipFrom: models.ShipFromAddress{
			ContactName: "[Aramex] Contact name",
			CompanyName: "[Aramex] Testing Company",
			Email:       "aramext@test.com",
			Street1:     "Testing Street",
			Phone:       "+919428293450",
			City:        "Mumbai",
			State:       "MH",
			PostalCode:  "400064",
			Country:     "IND",
			Type:        "business",
		},
		ShipTo: models.ShipToAddress{
			ContactName: "Dr. Moises Corwin",
			Phone:       "1-140-225-6410",
			Email:       "Giovanna42@yahoo.com",
			Street1:     "28292 Daugherty Orchard",
			City:        "Amman",
			PostalCode:  "560034",
			Country:     "JOR",
			Type:        "residential",
		},
		Parcels: []models.Parcel{
			{
				Description: "Food XS",
				BoxType:     "custom",
				Weight: models.Weight{
					Value: 2,
					Unit:  "kg",
				},
				Dimension: models.Dimension{
					Width:  20,
					Height: 40,
					Depth:  40,
					Unit:   "cm",
				},
				Items: []models.Item{
					{
						Description:   "Food Bar",
						OriginCountry: "USA",
						Quantity:      2,
						Price: models.Money{
							Amount:   3,
							Currency: "USD",
						},
						Weight: models.Weight{
							Value: 0.6,
							Unit:  "kg",
						},
						Sku: "imac2014",
					},
				},
			},
		},
	},
}

var searchDataForLabel = models.ListAllLabelsRequest{
	ShipperAccountID: "07b55d06-48af-4b02-a6b0-1e311e22b1e6",
	Status:           "created",
	Limit:            "20",
	NextToken:        "",
	S:                "376",
	CreatedAtMax:     "2018-10-06T10:38:05+00:00",
	CreatedAtMin:     "2018-10-07T10:38:05+00:00",
}

var shipperAccount = &models.CreateShipperAccountRequest{
	Slug:        "aramex",
	Description: "My Shipper Account",
	Timezone:    "Asia/Hong_Kong",
	Credentials: &models.ShipperAccountCredentials{
		AccountNumber: "11111",
		Username:      "johnwon",
		Password:      "12345",
		AccountPin:    "22222",
		AccountEntity: "33333",
	},
	Address: models.ShipFromAddress{
		Country:     "USA",
		ContactName: "Sir Foo",
		Phone:       "2125551234",
		Fax:         "+1 206-654-3100",
		Email:       "foo@foo.com",
		CompanyName: "Foo Store",
		Street1:     "255 New town",
		Street2:     "Wow Avenue",
		Street3:     "Boring part of town",
		City:        "Beverly Hills",
		Type:        "business",
		PostalCode:  "90210",
		State:       "CA",
		TaxID:       "911-70-1234",
	},
}

var searchDataForShipperAccounts = models.ListAllShipperAccountsRequest{
	Slug:      "aramex",
	Limit:     "30",
	NextToken: "eyJzbHVnIjoiYXJhbWV4IiwiaWQiOiI3MTg4NDE5ZS1iYTc0LTRhMTYtYmVjNS1iZDllYzUxMDE0NDUiLCJ1c2VyX2lkIjoiYTM3ODhlOGMtNzQwNi00ZWY4LWI3NDgtZWMzYmJiMGE3MzEyIn0=",
}

var shipperAccountCred = &models.UpdateShipperAccountCredRequest{
	AccountNumber: "11111",
	UserName:      "johnwon",
	Password:      "12345",
	AccountPin:    "22222",
	AccountEntity: "33333",
}

var shipperAccountInfo = &models.UpdateShipperAccountInfoRequest{
	Description: "My Shipper Account",
	Timezone:    "Asia/Hong_Kong",
	Address: &models.AddressShipperAccount{
		Country:     "USA",
		ContactName: "Sir Foo",
		CompanyName: "Foo Store",
		Street1:     "255 New town",
	},
}

var searchDataForBulkDownloads = models.ListAllBulkDownloadsRequest{
	Status:       "created",
	Limit:        "20",
	NextToken:    "",
	CreatedAtMax: "2018-10-07T11:45:55+00:00",
	CreatedAtMin: "2018-10-06T11:45:55+00:00",
}

var bulkDownloads = &models.CreateBulkDownloadRequest{
	Async: false,
	Labels: []models.Reference{
		{
			ID: "6b31853b-7fe2-4192-bb31-4c6954ba3cbb",
		},
		{
			ID: "2261acd6-22c8-4083-8c99-a4616f34738a",
		},
	},
}

func TestCalculateRates(t *testing.T) {

	ret, err := CalculateRates(calcRates, APIToken)
	if err != nil {
		t.Errorf("Error : %s\n", err)
		return
	}

	text, _ := json.Marshal(ret.Data)
	t.Logf("Result : %s\n", text)
	time.Sleep(time.Second)
}

func TestListAllRates(t *testing.T) {

	ret, err := ListAllRates(searchDataForRate, APIToken)
	if err != nil {
		t.Errorf("Error : %s\n", err)
		return
	}

	text, _ := json.Marshal(ret.Data)
	t.Logf("Result : %s\n", text)
	time.Sleep(time.Second)
}

func TestGetRates(t *testing.T) {

	ret, err := GetRates(DemoRateID, APIToken)
	if err != nil {
		t.Errorf("Error : %s\n", err)
		return
	}

	text, _ := json.Marshal(ret.Data)
	t.Logf("Result : %s\n", text)
	time.Sleep(time.Second)
}

func TestCreateLabel(t *testing.T) {

	ret, err := CreateLabel(labels, APIToken)
	if err != nil {
		t.Errorf("Error : %s\n", err)
		return
	}

	text, _ := json.Marshal(ret.Data)
	t.Logf("Result : %s\n", text)
	time.Sleep(time.Second)
}

func TestListAllLabels(t *testing.T) {

	ret, err := ListAllLabels(searchDataForLabel, APIToken)
	if err != nil {
		t.Errorf("Error : %s\n", err)
		return
	}

	text, _ := json.Marshal(ret.Data)
	t.Logf("Result : %s\n", text)
	time.Sleep(time.Second)
}

func TestGetLabel(t *testing.T) {

	ret, err := GetLabel(DemoLabelID, APIToken)
	if err != nil {
		t.Errorf("Error : %s\n", err)
		return
	}

	text, _ := json.Marshal(ret.Data)
	t.Logf("Result : %s\n", text)
	time.Sleep(time.Second)
}

func TestCreateShipperAccount(t *testing.T) {

	ret, err := CreateShipperAccount(shipperAccount, APIToken)
	if err != nil {
		t.Errorf("Error : %s\n", err)
		return
	}

	text, _ := json.Marshal(ret.Data)
	t.Logf("Result : %s\n", text)
	time.Sleep(time.Second)
}

func TestDeleteShipperAccount(t *testing.T) {

	ret, err := DeleteShipperAccount(DemoShipperAccountID1, APIToken)
	if err != nil {
		t.Errorf("Error : %s\n", err)
		return
	}

	text, _ := json.Marshal(ret.Data)
	t.Logf("Result : %s\n", text)
	time.Sleep(time.Second)
}

func TestListAllShipperAccounts(t *testing.T) {

	ret, err := ListAllShipperAccounts(searchDataForShipperAccounts, APIToken)
	if err != nil {
		t.Errorf("Error : %s\n", err)
		return
	}

	text, _ := json.Marshal(ret.Data)
	t.Logf("Result : %s\n", text)
	time.Sleep(time.Second)
}

func TestGetShipperAccount(t *testing.T) {

	ret, err := GetShipperAccount(DemoShipperAccountID2, APIToken)
	if err != nil {
		t.Errorf("Error : %s\n", err)
		return
	}

	text, _ := json.Marshal(ret.Data)
	t.Logf("Result : %s\n", text)
	time.Sleep(time.Second)
}

func TestUpdateShipperAccountCred(t *testing.T) {

	ret, err := UpdateShipperAccountCred(shipperAccountCred, DemoShipperAccountID2, APIToken)
	if err != nil {
		t.Errorf("Error : %s\n", err)
		return
	}

	text, _ := json.Marshal(ret.Data)
	t.Logf("Result : %s\n", text)
	time.Sleep(time.Second)
}

func TestUpdateShipperAccountInfo(t *testing.T) {

	ret, err := UpdateShipperAccountInfo(shipperAccountInfo, DemoShipperAccountID2, APIToken)
	if err != nil {
		t.Errorf("Error : %s\n", err)
		return
	}

	text, _ := json.Marshal(ret.Data)
	t.Logf("Result : %s\n", text)
	time.Sleep(time.Second)
}

func TestCreateBulkDownload(t *testing.T) {

	ret, err := CreateBulkDownload(bulkDownloads, APIToken)
	if err != nil {
		t.Errorf("Error : %s\n", err)
		return
	}

	text, _ := json.Marshal(ret.Data)
	t.Logf("Result : %s\n", text)
	time.Sleep(time.Second)
}

func TestListAllBulkDownloads(t *testing.T) {

	ret, err := ListAllBulkDownloads(searchDataForBulkDownloads, APIToken)
	if err != nil {
		t.Errorf("Error : %s\n", err)
		return
	}

	text, _ := json.Marshal(ret.Data)
	t.Logf("Result : %s\n", text)
	time.Sleep(time.Second)
}

func TestGetBulkDownload(t *testing.T) {

	ret, err := GetBulkDownload(DemoBulkDownloadID, APIToken)
	if err != nil {
		t.Errorf("Error : %s\n", err)
		return
	}

	text, _ := json.Marshal(ret.Data)
	t.Logf("Result : %s\n", text)
	time.Sleep(time.Second)
}
