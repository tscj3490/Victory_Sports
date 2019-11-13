package models

///////////////////////   Rates   /////////////////////////////////////
// CalculateRatesRequest
type CalculateRatesRequest struct {
	Shipment        Shipment    `json:"shipment,omitempty"`
	Async           bool        `json:"async,omitempty"`
	IsDocument      bool        `json:"is_document,omitempty"`
	ShipperAccounts []Reference `json:"shipper_accounts,omitempty"`
}

// Shipment
type Shipment struct {
	ShipFrom ShipFromAddress `json:"ship_from,omitempty"`
	ShipTo   ShipToAddress   `json:"ship_to,omitempty"`
	Parcels  []Parcel        `json:"parcels,omitempty"`
}

// ShipFromAddress
type ShipFromAddress struct {
	Country     string `json:"country,omitempty"`
	ContactName string `json:"contact_name,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Fax         string `json:"fax,omitempty"`
	Email       string `json:"email,omitempty"`
	CompanyName string `json:"company_name,omitempty"`
	Street1     string `json:"street1,omitempty"`
	Street2     string `json:"street2,omitempty"`
	Street3     string `json:"street3,omitempty"`
	City        string `json:"city,omitempty"`
	State       string `json:"state,omitempty"`
	PostalCode  string `json:"postal_code,omitempty"`
	Type        string `json:"type,omitempty"`
	TaxID       string `json:"tax_id,omitempty"`
}

// ShipToAddress
type ShipToAddress struct {
	Country     string `json:"country,omitempty"`
	ContactName string `json:"contact_name,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Fax         string `json:"fax,omitempty"`
	Email       string `json:"email,omitempty"`
	CompanyName string `json:"company_name,omitempty"`
	Street1     string `json:"street1,omitempty"`
	Street2     string `json:"street2,omitempty"`
	Street3     string `json:"street3,omitempty"`
	City        string `json:"city,omitempty"`
	State       string `json:"state,omitempty"`
	PostalCode  string `json:"postal_code,omitempty"`
	Type        string `json:"type,omitempty"`
	TaxID       string `json:"tax_id,omitempty"`
}

// Parcel
type Parcel struct {
	BoxType     string    `json:"box_type,omitempty"`
	Dimension   Dimension `json:"dimension,omitempty"`
	Items       []Item    `json:"items,omitempty"`
	Description string    `json:"description,omitempty"`
	Weight      Weight    `json:"weight,omitempty"`
}

// Dimension
type Dimension struct {
	Width  float32 `json:"width,omitempty"`
	Height float32 `json:"height,omitempty"`
	Depth  float32 `json:"depth,omitempty"`
	Unit   string  `json:"unit,omitempty"`
}

// Item
type Item struct {
	Description   string `json:"description,omitempty"`
	Quantity      uint   `json:"quantity,omitempty"`
	Price         Money  `json:"price,omitempty"`
	Weight        Weight `json:"weight,omitempty"`
	ItemID        string `json:"item_id,omitempty"`
	OriginCountry string `json:"origin_country,omitempty"`
	Sku           string `json:"sku,omitempty"`
	HsCode        string `json:"hs_code,omitempty"`
}

// Money
type Money struct {
	Amount   float32 `json:"amount,omitempty"`
	Currency string  `json:"currency,omitempty"`
}

// Weight
type Weight struct {
	Unit  string  `json:"unit,omitempty"`
	Value float32 `json:"value,omitempty"`
}

// Reference
type Reference struct {
	ID string `json:"id,omitempty"`
}

// CalculateRatesResponse
type CalculateRatesResponse struct {
	Meta ResponseMeta               `json:"meta,omitempty"`
	Data CalculateRatesResponseData `json:"data,omitempty"`
}

// CalculateRatesResponse
type CalculateRatesResponseData struct {
	ID        string `json:"id,omitempty"`
	Status    string `json:"status,omitempty"`
	Rates     []Rate `json:"rates,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// ResponseMeta
type ResponseMeta struct {
	Code    uint     `json:"code,omitempty"`
	Message string   `json:"message,omitempty"`
	Details []string `json:"details,omitempty"`
}

// Rate
type Rate struct {
	ShipperAccount  ShipperAccountInfo `json:"shipper_account,omitempty"`
	ServiceType     string             `json:"service_type,omitempty"`
	ServiceName     string             `json:"service_name,omitempty"`
	ChargeWeight    Weight             `json:"charge_weight,omitempty"`
	TotalCharge     Money              `json:"total_charge,omitempty"`
	PickupDeadline  string             `json:"pickup_deadline,omitempty"`
	BookingCutOff   string             `json:"booking_cut_off,omitempty"`
	DeliveryDate    string             `json:"delivery_date,omitempty"`
	TransitTime     uint               `json:"transit_time,omitempty"`
	DetailedCharges []DetailedCharges  `json:"detailed_charges,omitempty"`
	InfoMessage     string             `json:"info_message,omitempty"`
	ErrorMessage    string             `json:"error_message,omitempty"`
}

// ShipperAccountInfo
type ShipperAccountInfo struct {
	ID          string `json:"id,omitempty"`
	Slug        string `json:"slug,omitempty"`
	Description string `json:"description,omitempty"`
}

// DetailedCharges
type DetailedCharges struct {
	Type   string `json:"type,omitempty"`
	Charge Money  `json:"money,omitempty"`
}

// ListAllRatesRequest
type ListAllRatesRequest struct {
	Status       string `json:"status,omitempty"`
	Limit        string `json:"limit,omitempty"`
	CreatedAtMin string `json:"created_at_min,omitempty"`
	CreatedAtMax string `json:"created_at_max,omitempty"`
	NextToken    string `json:"next_token,omitempty"`
}

// ListAllRatesResponse
type ListAllRatesResponse struct {
	Meta ResponseMeta             `json:"meta,omitempty"`
	Data ListAllRatesResponseData `json:"data,omitempty"`
}

// ListAllRatesResponse
type ListAllRatesResponseData struct {
	NextToken string       `json:"next_token,omitempty"`
	Limit     float32      `json:"limit,omitempty"`
	Rates     []RateRecord `json:"rates,omitempty"`
}

// RateRecord
type RateRecord struct {
	ID        string `json:"id,omitempty"`
	Status    string `json:"status,omitempty"`
	Rates     []Rate `json:"rates,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// GetRatesResponse
type GetRatesResponse struct {
	Meta ResponseMeta `json:"meta,omitempty"`
	Data RateRecord   `json:"data,omitempty"`
}

///////////////////////   Label   /////////////////////////////////////

// CreateLabelRequest
type CreateLabelRequest struct {
	ServiceType    string      `json:"service_type,omitempty"`
	ShipperAccount Reference   `json:"shipper_account,omitempty"`
	Shipment       Shipment    `json:"shipment,omitempty"`
	Async          bool        `json:"async,omitempty"`
	ReturnShipment bool        `json:"return_shipment,omitempty"`
	PaperSize      string      `json:"paper_size,omitempty"`
	ShipDate       string      `json:"ship_date,omitempty"`
	ServiceOptions interface{} `json:"service_options,omitempty"`
	IsDocument     bool        `json:"is_document,omitempty"`
	Invoice        *Invoice    `json:"invoice,omitempty"`
	References     []string    `json:"references,omitempty"`
	Billing        *Billing    `json:"billing,omitempty"`
	Customs        *Customs    `json:"customs,omitempty"`
}

// ServiceOptionInsurance
type ServiceOptionInsurance struct {
	Type         string `json:"type,omitempty"`
	InsuredValue Money  `json:"insured_value,omitempty"`
}

// ServiceOptionCOD
type ServiceOptionCOD struct {
	Type     string `json:"type,omitempty"`
	CodValue Money  `json:"cod_value,omitempty"`
}

// ServiceOptionGeneral
type ServiceOptionGeneral struct {
	Type    string `json:"type,omitempty"`
	Enabled bool   `json:"enabled,omitempty"`
}

// Invoice
type Invoice struct {
	Date           string `json:"date,omitempty"`
	Number         string `json:"number,omitempty"`
	Type           string `json:"type,omitempty"`
	NumberOfCopies uint   `json:"number_of_copies,omitempty"`
}

// Billing
type Billing struct {
	PaidBy         string                `json:"paid_by,omitempty"`
	Method         *PaymentMethodAccount `json:"method,omitempty"`
	Type           string                `json:"type,omitempty"`
	NumberOfCopies uint                  `json:"number_of_copies,omitempty"`
}

// PaymentMethodAccount
type PaymentMethodAccount struct {
	Type          string `json:"type,omitempty"`
	AccountNumber string `json:"account_number,omitempty"`
	PostalCode    string `json:"postal_code,omitempty"`
	Country       string `json:"country,omitempty"`
}

// Customs
type Customs struct {
	Purpose         string          `json:"purpose,omitempty"`
	TermsOfTrade    string          `json:"terms_of_trade,omitempty"`
	Eei             interface{}     `json:"eei,omitempty"`
	Billing         *CustomsBilling `json:"billing,omitempty"`
	ImporterAddress *Address        `json:"importer_address,omitempty"`
	Passport        *Passport       `json:"passport,omitempty"`
}

// Aes
type Aes struct {
	Type      string `json:"type,omitempty"`
	ItnNumber string `json:"itn_number,omitempty"`
}

// NoEei
type NoEei struct {
	Type         string `json:"type,omitempty"`
	FtrExemption string `json:"ftr_exemption,omitempty"`
}

// CustomsBilling
type CustomsBilling struct {
	PaidBy string                `json:"paid_by,omitempty"`
	Method *PaymentMethodAccount `json:"method,omitempty"`
}

// Passport
type Passport struct {
	Number    string `json:"number,omitempty"`
	IssueDate string `json:"issue_date,omitempty"`
}

// CreateLabelResponse
type CreateLabelResponse struct {
	Meta ResponseMeta            `json:"meta,omitempty"`
	Data CreateLabelResponseData `json:"data,omitempty"`
}

// CreateLabelResponseData
type CreateLabelResponseData struct {
	ID              string             `json:"id,omitempty"`
	Status          string             `json:"status,omitempty"`
	ShipDate        string             `json:"ship_date,omitempty"`
	TrackingNumbers []string           `json:"tracking_numbers,omitempty"`
	Files           Files              `json:"files,omitempty"`
	Rate            Rate               `json:"rate,omitempty"`
	CreatedAt       string             `json:"created_at,omitempty"`
	UpdatedAt       string             `json:"updated_at,omitempty"`
	References      []string           `json:"references,omitempty"`
	ShipperAccount  ShipperAccountInfo `json:"shipper_account,omitempty"`
	ServiceType     string             `json:"service_type,omitempty"`
}

// Files
type Files struct {
	Label              LabelFile              `json:"label,omitempty"`
	Invoice            InvoiceFile            `json:"invoice,omitempty"`
	CustomsDeclaration CustomsDeclarationFile `json:"customs_declaration,omitempty"`
	Manifest           interface{}            `json:"manifest,omitempty"`
}

// LabelFile
type LabelFile struct {
	PaperSize string `json:"paper_size,omitempty"`
	URL       string `json:"url,omitempty"`
	FileType  string `json:"file_type,omitempty"`
}

// InvoiceFile
type InvoiceFile struct {
	PaperSize string `json:"paper_size,omitempty"`
	URL       string `json:"url,omitempty"`
	FileType  string `json:"file_type,omitempty"`
}

// CustomsDeclarationFile
type CustomsDeclarationFile struct {
	PaperSize string `json:"paper_size,omitempty"`
	URL       string `json:"url,omitempty"`
	FileType  string `json:"file_type,omitempty"`
}

// ManifestPDFFile
type ManifestPDFFile struct {
	PaperSize string `json:"paper_size,omitempty"`
	URL       string `json:"url,omitempty"`
	FileType  string `json:"file_type,omitempty"`
}

// ManifestCommonFile
type ManifestCommonFile struct {
	PaperSize string `json:"paper_size,omitempty"`
	URL       string `json:"url,omitempty"`
	FileType  string `json:"file_type,omitempty"`
}

// ListAllLabelsRequest
type ListAllLabelsRequest struct {
	ShipperAccountID string `json:"shipper_account_id,omitempty"`
	Status           string `json:"status,omitempty"`
	Limit            string `json:"limit,omitempty"`
	CreatedAtMin     string `json:"created_at_min,omitempty"`
	CreatedAtMax     string `json:"created_at_max,omitempty"`
	TrackingNumbers  string `json:"tracking_numbers,omitempty"`
	S                string `json:"s,omitempty"`
	NextToken        string `json:"next_token,omitempty"`
}

// ListAllLabelsResponse
type ListAllLabelsResponse struct {
	Meta ResponseMeta              `json:"meta,omitempty"`
	Data ListAllLabelsResponseData `json:"data,omitempty"`
}

// ListAllLabelsResponseData
type ListAllLabelsResponseData struct {
	NextToken string   `json:"next_token,omitempty"`
	Limit     float32  `json:"limit,omitempty"`
	Labels    []*Label `json:"labels,omitempty"`
}

// Label
type Label struct {
	ID              string   `json:"id,omitempty"`
	Status          string   `json:"status,omitempty"`
	ShipDate        string   `json:"ship_date,omitempty"`
	TrackingNumbers []string `json:"tracking_numbers,omitempty"`
	Files           Files    `json:"files,omitempty"`
	Rate            Rate     `json:"rate,omitempty"`
	CreatedAt       string   `json:"created_at,omitempty"`
	UpdatedAt       string   `json:"updated_at,omitempty"`
}

// GetLabelResponse
type GetLabelResponse struct {
	Meta ResponseMeta         `json:"meta,omitempty"`
	Data GetLabelResponseData `json:"data,omitempty"`
}

// GetLabelResponseData
type GetLabelResponseData struct {
	ID              string   `json:"id,omitempty"`
	Status          string   `json:"status,omitempty"`
	ShipDate        string   `json:"ship_date,omitempty"`
	TrackingNumbers []string `json:"tracking_numbers,omitempty"`
	Files           Files    `json:"files,omitempty"`
	Rate            Rate     `json:"rate,omitempty"`
	CreatedAt       string   `json:"created_at,omitempty"`
	UpdatedAt       string   `json:"updated_at,omitempty"`
}

///////////////////////   ShipperAccount   /////////////////////////////////////

// CreateShipperAccountRequest
type CreateShipperAccountRequest struct {
	Slug        string                     `json:"slug,omitempty"`
	Description string                     `json:"description,omitempty"`
	Address     ShipFromAddress            `json:"address,omitempty"`
	Timezone    string                     `json:"timezone,omitempty"`
	Credentials *ShipperAccountCredentials `json:"credentials,omitempty"`
}

// ShipperAccountCredentials
type ShipperAccountCredentials struct {
	AccountNumber string `json:"account_number,omitempty"`
	Username      string `json:"username,omitempty"`
	Password      string `json:"password,omitempty"`
	AccountPin    string `json:"account_pin,omitempty"`
	AccountEntity string `json:"account_entity,omitempty"`
}

// CreateShipperAccountResponse
type CreateShipperAccountResponse struct {
	Meta ResponseMeta                     `json:"meta,omitempty"`
	Data CreateShipperAccountResponseData `json:"data,omitempty"`
}

// CreateShipperAccountResponseData
type CreateShipperAccountResponseData struct {
	ID          string          `json:"id,omitempty"`
	Address     ShipFromAddress `json:"address,omitempty"`
	Slug        string          `json:"slug,omitempty"`
	Status      string          `json:"status,omitempty"`
	Description string          `json:"description,omitempty"`
	Type        string          `json:"type,omitempty"`
	Timezone    string          `json:"timezone,omitempty"`
	CreatedAt   string          `json:"created_at,omitempty"`
	UpdatedAt   string          `json:"updated_at,omitempty"`
}

// DeleteShipperAccountResponse
type DeleteShipperAccountResponse struct {
	Meta ResponseMeta                     `json:"meta,omitempty"`
	Data DeleteShipperAccountResponseData `json:"data,omitempty"`
}

// DeleteShipperAccountResponseData
type DeleteShipperAccountResponseData struct {
	ID string `json:"id,omitempty"`
}

// ListAllShipperAccountsRequest
type ListAllShipperAccountsRequest struct {
	Slug      string `json:"slug,omitempty"`
	Limit     string `json:"limit,omitempty"`
	NextToken string `json:"next_token,omitempty"`
}

// ListAllShipperAccountsResponse
type ListAllShipperAccountsResponse struct {
	Meta ResponseMeta                       `json:"meta,omitempty"`
	Data ListAllShipperAccountsResponseData `json:"data,omitempty"`
}

// ListAllShipperAccountResponseData
type ListAllShipperAccountsResponseData struct {
	Limit           int               `json:"limit,omitempty"`
	NextToken       string            `json:"next_token,omitempty"`
	ShipperAccounts []*ShipperAccount `json:"shipper_accounts,omitempty"`
}

// ShipperAccount
type ShipperAccount struct {
	ID          string          `json:"id,omitempty"`
	Address     ShipFromAddress `json:"address,omitempty"`
	Slug        string          `json:"slug,omitempty"`
	Status      string          `json:"status,omitempty"`
	Description string          `json:"description,omitempty"`
	Type        string          `json:"type,omitempty"`
	Timezone    string          `json:"timezone,omitempty"`
	CreatedAt   string          `json:"created_at,omitempty"`
	UpdatedAt   string          `json:"updated_at,omitempty"`
}

// GetShipperAccountResponse
type GetShipperAccountResponse struct {
	Meta ResponseMeta                  `json:"meta,omitempty"`
	Data GetShipperAccountResponseData `json:"data,omitempty"`
}

// GetShipperAccountResponseData
type GetShipperAccountResponseData struct {
	ID          string          `json:"id,omitempty"`
	Address     ShipFromAddress `json:"address,omitempty"`
	Slug        string          `json:"slug,omitempty"`
	Status      string          `json:"status,omitempty"`
	Description string          `json:"description,omitempty"`
	Type        string          `json:"type,omitempty"`
	Timezone    string          `json:"timezone,omitempty"`
	CreatedAt   string          `json:"created_at,omitempty"`
	UpdatedAt   string          `json:"updated_at,omitempty"`
}

// UpdateShipperAccountCredRequest
type UpdateShipperAccountCredRequest struct {
	AccountNumber string `json:"account_number,omitempty"`
	UserName      string `json:"username,omitempty"`
	Password      string `json:"password,omitempty"`
	AccountPin    string `json:"account_pin,omitempty"`
	AccountEntity string `json:"account_entity,omitempty"`
}

// UpdateShipperAccountCredResponse
type UpdateShipperAccountCredResponse struct {
	Meta ResponseMeta                         `json:"meta,omitempty"`
	Data UpdateShipperAccountCredResponseData `json:"data,omitempty"`
}

// UpdateShipperAccountCredResponseData
type UpdateShipperAccountCredResponseData struct {
	ID string `json:"id,omitempty"`
}

// UpdateShipperAccountInfoRequest
type UpdateShipperAccountInfoRequest struct {
	Description string                 `json:"description,omitempty"`
	Timezone    string                 `json:"timezone,omitempty"`
	Address     *AddressShipperAccount `json:"address,omitempty"`
}

// AddressShipperAccount
type AddressShipperAccount struct {
	ContactName string `json:"contact_name,omitempty"`
	CompanyName string `json:"company_name,omitempty"`
	Street1     string `json:"street1,omitempty"`
	Country     string `json:"country,omitempty"`
}

// UpdateShipperAccountInfoResponse
type UpdateShipperAccountInfoResponse struct {
	Meta ResponseMeta                         `json:"meta,omitempty"`
	Data UpdateShipperAccountInfoResponseData `json:"data,omitempty"`
}

// UpdateShipperAccountInfoResponseData
type UpdateShipperAccountInfoResponseData struct {
	ID          string                `json:"id,omitempty"`
	Address     AddressShipperAccount `json:"address,omitempty"`
	Slug        string                `json:"slug,omitempty"`
	Status      string                `json:"status,omitempty"`
	Description string                `json:"description,omitempty"`
	Timezone    string                `json:"timezone,omitempty"`
	Type        string                `json:"type,omitempty"`
	CreatedAt   string                `json:"created_at,omitempty"`
	UpdatedAt   string                `json:"updated_at,omitempty"`
}

///////////////////////   BulkDownload   /////////////////////////////////////

// CreateBulkDownloadRequest
type CreateBulkDownloadRequest struct {
	Labels []Reference `json:"labels,omitempty"`
	Async  bool        `json:"async,omitempty"`
}

// CreateBulkDownloadResponse
type CreateBulkDownloadResponse struct {
	Meta ResponseMeta                   `json:"meta,omitempty"`
	Data CreateBulkDownloadResponseData `json:"data,omitempty"`
}

// CreateBulkDownloadResponseData
type CreateBulkDownloadResponseData struct {
	ID            string      `json:"id,omitempty"`
	Status        string      `json:"status,omitempty"`
	Files         Files       `json:"files,omitempty"`
	Labels        []Reference `json:"labels,omitempty"`
	InvalidLabels []Reference `json:"invalid_labels,omitempty"`
	CreatedAt     string      `json:"created_at,omitempty"`
	UpdatedAt     string      `json:"updated_at,omitempty"`
}

// ListAllBulkDownloadsRequest
type ListAllBulkDownloadsRequest struct {
	Status       string `json:"status,omitempty"`
	Limit        string `json:"limit,omitempty"`
	CreatedAtMin string `json:"created_at_min,omitempty"`
	CreatedAtMax string `json:"created_at_max,omitempty"`
	NextToken    string `json:"next_token,omitempty"`
}

// ListAllBulkDownloadsResponse
type ListAllBulkDownloadsResponse struct {
	Meta ResponseMeta                     `json:"meta,omitempty"`
	Data ListAllBulkDownloadsResponseData `json:"data,omitempty"`
}

// ListAllBulkDownloadsResponseData
type ListAllBulkDownloadsResponseData struct {
	Limit         int             `json:"limit,omitempty"`
	NextToken     string          `json:"next_token,omitempty"`
	CreatedAtMin  string          `json:"created_at_min,omitempty"`
	CreatedAtMax  string          `json:"created_at_max,omitempty"`
	BulkDownloads []*BulkDownload `json:"bulk-downloads,omitempty"`
}

// BulkDownload
type BulkDownload struct {
	ID            string       `json:"id,omitempty"`
	Status        string       `json:"status,omitempty"`
	Files         Files        `json:"files,omitempty"`
	Labels        []*Reference `json:"labels,omitempty"`
	InvalidLabels []*Reference `json:"invalid_labels,omitempty"`
	CreatedAt     string       `json:"created_at,omitempty"`
	UpdatedAt     string       `json:"updated_at,omitempty"`
}

// GetBulkDownloadResponse
type GetBulkDownloadResponse struct {
	Meta ResponseMeta                `json:"meta,omitempty"`
	Data GetBulkDownloadResponseData `json:"data,omitempty"`
}

// GetBulkDownloadResponseData
type GetBulkDownloadResponseData struct {
	ID            string       `json:"id,omitempty"`
	Status        string       `json:"status,omitempty"`
	Files         Files        `json:"files,omitempty"`
	Labels        []*Reference `json:"labels,omitempty"`
	InvalidLabels []*Reference `json:"invalid_labels,omitempty"`
	CreatedAt     string       `json:"created_at,omitempty"`
	UpdatedAt     string       `json:"updated_at,omitempty"`
}
