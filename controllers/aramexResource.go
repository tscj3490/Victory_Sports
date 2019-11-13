package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"bitbucket.org/softwarehouseio/victory/victory-frontend/models"
)

var postmenRootURL = "https://sandbox-api.postmen.com/v3"

// CalculateRates
func CalculateRates(calcRateRequest *models.CalculateRatesRequest, token string) (models.CalculateRatesResponse, error) {
	// Send a POST request
	jsonStr, err := json.Marshal(calcRateRequest)

	if err != nil {
		fmt.Println(err)
	}

	req, err := http.NewRequest("POST", postmenRootURL+"/rates", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("postmen-api-key", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("unexpected error in sending req: %v", err)
	}
	if resp.StatusCode == http.StatusInternalServerError {
		fmt.Println("Skipping CalculateRatesRequest because already created.")
	}
	r := models.CalculateRatesResponse{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	json.Unmarshal(body, &r)

	return r, nil
}

// ListAllRates
func ListAllRates(searchRequest models.ListAllRatesRequest, token string) (models.ListAllRatesResponse, error) {
	req, err := http.NewRequest("GET", postmenRootURL+"/rates?status="+searchRequest.Status+"&limit="+searchRequest.Limit+
		"&next_token="+searchRequest.NextToken+"&created_at_max="+searchRequest.CreatedAtMax+"&created_at_min="+searchRequest.CreatedAtMin, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("postmen-api-key", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("unexpected error in sending req: %v", err)
	}
	if resp.StatusCode == http.StatusInternalServerError {
		fmt.Println("Skipping ListAllRatesRequest because already created.")
	}

	r := models.ListAllRatesResponse{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	json.Unmarshal(body, &r)

	return r, nil
}

// GetRates
func GetRates(id string, token string) (models.GetRatesResponse, error) {
	req, err := http.NewRequest("GET", postmenRootURL+"/rates/"+id, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("postmen-api-key", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("unexpected error in sending req: %v", err)
	}
	if resp.StatusCode == http.StatusInternalServerError {
		fmt.Println("Skipping GetRatesRequest because already created.")
	}

	r := models.GetRatesResponse{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	json.Unmarshal(body, &r)

	return r, nil
}

// CreateLabel
func CreateLabel(createLabelRequest *models.CreateLabelRequest, token string) (models.CreateLabelResponse, error) {
	// Send a POST request
	jsonStr, err := json.Marshal(createLabelRequest)

	if err != nil {
		fmt.Println(err)
	}

	req, err := http.NewRequest("POST", postmenRootURL+"/labels", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("postmen-api-key", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("unexpected error in sending req: %v", err)
	}
	if resp.StatusCode == http.StatusInternalServerError {
		fmt.Println("Skipping CreateLabelRequest because already created.")
	}
	r := models.CreateLabelResponse{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	json.Unmarshal(body, &r)

	return r, nil
}

// ListAllLabels
func ListAllLabels(searchRequest models.ListAllLabelsRequest, token string) (models.ListAllLabelsResponse, error) {
	req, err := http.NewRequest("GET", postmenRootURL+"/labels?shipper_account_id="+searchRequest.ShipperAccountID+"&status="+searchRequest.Status+
		"&limit="+searchRequest.Limit+"&next_token="+searchRequest.NextToken+"&created_at_max="+searchRequest.CreatedAtMax+
		"&created_at_min="+searchRequest.CreatedAtMin+"&s="+searchRequest.S, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("postmen-api-key", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("unexpected error in sending req: %v", err)
	}
	if resp.StatusCode == http.StatusInternalServerError {
		fmt.Println("Skipping ListAllLabelsRequest because already created.")
	}

	r := models.ListAllLabelsResponse{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	json.Unmarshal(body, &r)

	return r, nil
}

// GetLabel
func GetLabel(id string, token string) (models.GetLabelResponse, error) {
	req, err := http.NewRequest("GET", postmenRootURL+"/labels/"+id, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("postmen-api-key", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("unexpected error in sending req: %v", err)
	}
	if resp.StatusCode == http.StatusInternalServerError {
		fmt.Println("Skipping GetLabelRequest because already created.")
	}

	r := models.GetLabelResponse{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	json.Unmarshal(body, &r)

	return r, nil
}

// CreateShipperAccount
func CreateShipperAccount(createShipperAccountRequest *models.CreateShipperAccountRequest, token string) (models.CreateShipperAccountResponse, error) {
	// Send a POST request
	jsonStr, err := json.Marshal(createShipperAccountRequest)

	if err != nil {
		fmt.Println(err)
	}

	req, err := http.NewRequest("POST", postmenRootURL+"/shipper-accounts", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("postmen-api-key", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("unexpected error in sending req: %v", err)
	}
	if resp.StatusCode == http.StatusInternalServerError {
		fmt.Println("Skipping CreateShipperAccountRequest because already created.")
	}
	r := models.CreateShipperAccountResponse{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	json.Unmarshal(body, &r)

	return r, nil
}

// DeleteShipperAccount
func DeleteShipperAccount(id string, token string) (models.DeleteShipperAccountResponse, error) {
	// Send a DELETE request
	req, err := http.NewRequest("DELETE", postmenRootURL+"/shipper-accounts/"+id, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("postmen-api-key", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("unexpected error in sending req: %v", err)
	}
	if resp.StatusCode == http.StatusInternalServerError {
		fmt.Println("Skipping DeleteShipperAccountRequest because already created.")
	}

	r := models.DeleteShipperAccountResponse{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	json.Unmarshal(body, &r)

	return r, nil
}

// ListAllShipperAccounts
func ListAllShipperAccounts(searchRequest models.ListAllShipperAccountsRequest, token string) (models.ListAllShipperAccountsResponse, error) {
	// Send a GET request
	req, err := http.NewRequest("GET", postmenRootURL+"/shipper-accounts?slug="+searchRequest.Slug+"&limit="+searchRequest.Limit+
		"&next_token="+searchRequest.NextToken, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("postmen-api-key", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("unexpected error in sending req: %v", err)
	}
	if resp.StatusCode == http.StatusInternalServerError {
		fmt.Println("Skipping ListAllShipperAccountsRequest because already created.")
	}

	r := models.ListAllShipperAccountsResponse{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	json.Unmarshal(body, &r)

	return r, nil
}

// GetShipperAccount
func GetShipperAccount(id string, token string) (models.GetShipperAccountResponse, error) {
	// Send a GET request
	req, err := http.NewRequest("GET", postmenRootURL+"/shipper-accounts/"+id, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("postmen-api-key", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("unexpected error in sending req: %v", err)
	}
	if resp.StatusCode == http.StatusInternalServerError {
		fmt.Println("Skipping GetLabelRequest because already created.")
	}

	r := models.GetShipperAccountResponse{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	json.Unmarshal(body, &r)

	return r, nil
}

// UpdateShipperAccountCred
func UpdateShipperAccountCred(updateShipperAccountCredRequest *models.UpdateShipperAccountCredRequest, id string, token string) (models.UpdateShipperAccountCredResponse, error) {
	// Send a PUT request
	jsonStr, err := json.Marshal(updateShipperAccountCredRequest)

	if err != nil {
		fmt.Println(err)
	}

	req, err := http.NewRequest("PUT", postmenRootURL+"/shipper-accounts/"+id+"/credentials", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("postmen-api-key", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("unexpected error in sending req: %v", err)
	}
	if resp.StatusCode == http.StatusInternalServerError {
		fmt.Println("Skipping UpdateShipperAccountCredRequest because already created.")
	}
	r := models.UpdateShipperAccountCredResponse{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	json.Unmarshal(body, &r)

	return r, nil
}

// UpdateShipperAccountInfo
func UpdateShipperAccountInfo(updateShipperAccountInfoRequest *models.UpdateShipperAccountInfoRequest, id string, token string) (models.UpdateShipperAccountInfoResponse, error) {
	// Send a PUT request
	jsonStr, err := json.Marshal(updateShipperAccountInfoRequest)

	if err != nil {
		fmt.Println(err)
	}

	req, err := http.NewRequest("PUT", postmenRootURL+"/shipper-accounts/"+id+"/info", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("postmen-api-key", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("unexpected error in sending req: %v", err)
	}
	if resp.StatusCode == http.StatusInternalServerError {
		fmt.Println("Skipping UpdateShipperAccountInfoRequest because already created.")
	}
	r := models.UpdateShipperAccountInfoResponse{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	json.Unmarshal(body, &r)

	return r, nil
}

// CreateBulkDownload
func CreateBulkDownload(createBulkDownloadRequest *models.CreateBulkDownloadRequest, token string) (models.CreateBulkDownloadResponse, error) {
	// Send a POST request
	jsonStr, err := json.Marshal(createBulkDownloadRequest)

	if err != nil {
		fmt.Println(err)
	}

	req, err := http.NewRequest("POST", postmenRootURL+"/bulk-downloads", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("postmen-api-key", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("unexpected error in sending req: %v", err)
	}
	if resp.StatusCode == http.StatusInternalServerError {
		fmt.Println("Skipping CreateBulkDownloadRequest because already created.")
	}
	r := models.CreateBulkDownloadResponse{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	json.Unmarshal(body, &r)

	return r, nil
}

// ListAllBulkDownloads
func ListAllBulkDownloads(searchRequest models.ListAllBulkDownloadsRequest, token string) (models.ListAllBulkDownloadsResponse, error) {
	// Send a GET request
	req, err := http.NewRequest("GET", postmenRootURL+"/bulk-downloads?status="+searchRequest.Status+"&limit="+searchRequest.Limit+
		"&next_token="+searchRequest.NextToken+"&created_at_max="+searchRequest.CreatedAtMax+"&created_at_min="+searchRequest.CreatedAtMin, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("postmen-api-key", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("unexpected error in sending req: %v", err)
	}
	if resp.StatusCode == http.StatusInternalServerError {
		fmt.Println("Skipping ListAllBulkDownloadsRequest because already created.")
	}

	r := models.ListAllBulkDownloadsResponse{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	json.Unmarshal(body, &r)

	return r, nil
}

// GetBulkDownload
func GetBulkDownload(id string, token string) (models.GetBulkDownloadResponse, error) {
	// Send a GET request
	req, err := http.NewRequest("GET", postmenRootURL+"/bulk-downloads/"+id, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("postmen-api-key", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("unexpected error in sending req: %v", err)
	}
	if resp.StatusCode == http.StatusInternalServerError {
		fmt.Println("Skipping GetBulkDownloadRequest because already created.")
	}

	r := models.GetBulkDownloadResponse{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	json.Unmarshal(body, &r)

	return r, nil
}
