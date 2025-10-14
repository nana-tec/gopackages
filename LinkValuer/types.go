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
