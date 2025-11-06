package linkvaluer

import "encoding/json"

// TokenPair represents access and refresh tokens

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// CreateRequest is the payload for creating a valuation request

type CreateRequest struct {
	CustomerName       string `json:"customer_name"`
	CustomerPhone      string `json:"customer_phone"`
	RegistrationNumber string `json:"registration_number"`
	PolicyNumber       string `json:"policy_number"`
	CustomerEmail      string `json:"customer_email,omitempty"`
	InsuranceCompany   string `json:"insurance_company,omitempty"`
	CallBackURL        string `json:"callback_url,omitempty"`
	PartnerReference   string `json:"partner_reference,omitempty"`
}

// Generic API response wrappers (kept for internal use only)

type APIResponse struct {
	Success bool            `json:"success,omitempty"`
	Message string          `json:"message,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
	Raw     json.RawMessage `json:"-"`
}

// CallbackResponse is the payload for the callback
//
//	{
//	 "booking_no": "LV_0277593",
//	 "status": "completed",
//	 "assessment_id": 267095,
//	 "reg_no": "KDO 950L",
//	 "completion_date": "2025-10-14T12:05:10.643616Z",
//	 "pdf_url": "https://portal.linksvaluers.com/pdf-report/LV_027dhishih7593",
//	 "partner_reference": "LIVE_TEST_1Q2W",
//	 "customer_name": "test",
//	 "insurance_company": "Ibime",
//	 "policy_number": "9056QSQ22LR",
//	 "market_value": "250000",
//	 "duty_free_value": "150000",
//	 "windscreen_value": "50000",
//	 "radio_value": "25000"
//	}
type CallbackResponse struct {
	BookingNo        string  `json:"booking_no"`
	Status           string  `json:"status"`
	AssessmentID     int     `json:"assessment_id"`
	RegNo            string  `json:"reg_no"`
	CompletionDate   string  `json:"completion_date"`
	PdfUrl           string  `json:"pdf_url"`
	PartnerReference string  `json:"partner_reference"`
	CustomerName     string  `json:"customer_name"`
	InsuranceCompany string  `json:"insurance_company"`
	PolicyNumber     string  `json:"policy_number"`
	MarketValue      float64 `json:"market_value"`
	DutyFreeValue    float64 `json:"duty_free_value"`
	WindscreenValue  float64 `json:"windscreen_value"`
	RadioValue       float64 `json:"radio_value"`
}

// CreateValuationPayload is a typed response for CreateValuation
// Minimal shape based on observed patterns; extend as needed.

type CreateValuationPayload struct {
	Success bool   `json:"success,omitempty"`
	Message string `json:"message,omitempty"`
	Data    struct {
		BookingNo string `json:"booking_no,omitempty"`
	} `json:"data,omitempty"`
}

type CreateResponse = APIResponse

type ViewAPIRequestsResponse struct {
	Message string                   `json:"message,omitempty"`
	Data    []map[string]interface{} `json:"data,omitempty"`
	Client  string                   `json:"client,omitempty"`
}

type AssessmentsResponse = APIResponse

// Assessments models

type AssessmentItem struct {
	BookingNo        string  `json:"booking_no"`
	RegNo            string  `json:"reg_no"`
	Customer         string  `json:"customer"`
	ChassisNumber    string  `json:"chassis_number"`
	EngineNumber     string  `json:"engine_number"`
	EngineCapacity   string  `json:"engine_capacity"`
	Odometer         string  `json:"odometer"`
	AssessedValue    string  `json:"assessed_value"`
	PolicyNo         string  `json:"policy_no"`
	ManufactureYear  string  `json:"manufacture_year"`
	RegDate          string  `json:"reg_date"`
	Colour           string  `json:"colour"`
	TyreCondition    string  `json:"tyre_condition"`
	MechanicalCond   string  `json:"mechanical_condition"`
	ElectricalSystem string  `json:"electrical_system"`
	GeneralCondition string  `json:"general_condition"`
	Extras           string  `json:"extras"`
	Country          string  `json:"country"`
	Make             string  `json:"make"`
	Model            string  `json:"model"`
	Status           string  `json:"status"`
	DownloadURL      *string `json:"download_url"`
	CompletedOn      *string `json:"completed_on"`
	AssessedOn       *string `json:"assessed_on"`
}

type Pagination struct {
	Total       int `json:"total"`
	PerPage     int `json:"per_page"`
	CurrentPage int `json:"current_page"`
	LastPage    int `json:"last_page"`
}

type AssessmentsPayload struct {
	Data       []AssessmentItem `json:"data"`
	Pagination Pagination       `json:"pagination"`
}

// DecodeAssessments decodes the full assessments response body into AssessmentsPayload
func DecodeAssessments(raw json.RawMessage) (*AssessmentsPayload, error) {
	var p AssessmentsPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// DecodeAsessments (compat alias) delegates to DecodeAssessments
func DecodeAsessments(raw json.RawMessage) (*AssessmentsPayload, error) {
	return DecodeAssessments(raw)
}
