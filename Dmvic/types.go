// Package dmvic provides types for DMVIC API request and response structures.
package dmvic

import (
	"encoding/json"
	"fmt"
	"strings"
)

// LoginResponse represents the response from DMVIC login authentication.
// It contains authentication token, user information, and session details.
type LoginResponse struct {
	Token               string  `json:"token"`               // Authentication token for API requests
	LoginUserID         string  `json:"loginUserId"`         // Unique identifier for the logged-in user
	IssueAt             string  `json:"issueAt"`             // Token issuance timestamp
	Expires             string  `json:"expires"`             // Token expiration timestamp
	Code                int     `json:"code"`                // Response code (negative values indicate errors)
	LoginHistoryID      int     `json:"LoginHistoryId"`      // Login session history identifier
	FirstName           string  `json:"firstName"`           // User's first name
	LastName            string  `json:"lastName"`            // User's last name
	LoggedInEntityID    int     `json:"loggedinEntityId"`    // Entity ID of the logged-in organization
	APIMSubscriptionKey *string `json:"ApimSubscriptionKey"` // API management subscription key
	IndustryTypeID      int     `json:"IndustryTypeId"`      // Industry type identifier
}

// CertificateRequest represents a request to retrieve certificate information.
type CertificateRequest struct {
	CertificateNumber string `json:"certificateNumber"` // Certificate number to query
}

// CertificateResponse represents the response from certificate retrieval operations.
type CertificateResponse struct {
	Success          bool               `json:"success"`          // Indicates if the operation was successful
	Error            FlexibleDmvicError `json:"error,omitempty"`  // Error details if operation failed
	APIRequestNumber string             `json:"apiRequestNumber"` // Unique API request identifier
	Inputs           CertificateRequest `json:"inputs"`           // Original request parameters
	CallbackObj      CallbackURL        `json:"callbackObj"`      // Callback URL information
}

type DoubleInsuranceDetails struct {
	CoverEndDate           string `json:"CoverEndDate"`
	InsuranceCertificateNo string `json:"InsuranceCertificateNo"`
	MemberCompanyName      string `json:"MemberCompanyName"`
	RegistrationNumber     string `json:"RegistrationNumber"`
	ChassisNumber          string `json:"ChassisNumber"`
	CertificateStatus      string `json:"CertificateStatus"`
	InsurancePolicyNo      string `json:"InsurancePolicyNo"`
}

// CallbackURL contains callback URL information for asynchronous operations.
type CallbackURL struct {
	URL string `json:"URL"` // The callback URL
}

// InsuranceValidationRequest represents a request to validate insurance information.
type InsuranceValidationRequest struct {
	VehicleRegistrationNumber string `json:"vehicleRegistrationnumber"` // Vehicle registration number
	ChassisNumber             string `json:"Chassisnumber"`             // Vehicle chassis number
	CertificateNumber         string `json:"certificateNumber"`         // Insurance certificate number
}

// InsuranceValidationResponse represents the response from insurance validation operations.
type InsuranceValidationResponse struct {
	Inputs           InsuranceValidationRequest `json:"inputs"`           // Original request parameters
	Error            FlexibleDmvicError         `json:"error,omitempty"`  // Error details if operation failed
	Success          bool                       `json:"success"`          // Indicates if the operation was successful
	APIRequestNumber string                     `json:"apiRequestNumber"` // Unique API request identifier
	CallbackObj      InsuranceCallbackObj       `json:"callbackObj"`      // Insurance validation results
}

// InsuranceCallbackObj contains insurance validation results.
type InsuranceCallbackObj struct {
	ValidateInsurance InsuranceDetails `json:"validateInsurance"` // Detailed insurance information
}

// InsuranceDetails contains comprehensive insurance certificate information.
type InsuranceDetails struct {
	CertificateNumber     string `json:"CertificateNumber"`     // Insurance certificate number
	InsurancePolicyNumber string `json:"InsurancePolicyNumber"` // Insurance policy number
	ValidFrom             string `json:"ValidFrom"`             // Policy validity start date
	ValidTill             string `json:"ValidTill"`             // Policy validity end date
	RegistrationNumber    string `json:"Registrationnumber"`    // Vehicle registration number
	InsuredBy             string `json:"InsuredBy"`             // Insurance company name
	ChassisNumber         string `json:"Chassisnumber"`         // Vehicle chassis number
	InsuredName           string `json:"sInsuredName"`          // Name of the insured party
	Intermediary          string `json:"Intermediary"`          // Insurance intermediary name
	IntermediaryIRA       string `json:"IntermediaryIRA"`       // Intermediary IRA number
	CertificateStatus     string `json:"CertificateStatus"`     // Current status of the certificate
}

// CancellationRequest represents a request to cancel an insurance certificate.
type CancellationRequest struct {
	CertificateNumber string `json:"CertificateNumber"` // Certificate number to cancel
	CancelReasonID    int    `json:"cancelreasonid"`    // Reason code for cancellation
}

// CancellationResponse represents the response from certificate cancellation operations.
type CancellationResponse struct {
	Error            FlexibleDmvicError      `json:"error,omitempty"`  // Error details if operation failed
	Success          bool                    `json:"success"`          // Indicates if the operation was successful
	APIRequestNumber string                  `json:"apiRequestNumber"` // Unique API request identifier
	Inputs           CancellationRequest     `json:"Inputs"`           // Original request parameters
	CallbackObj      CancellationCallbackObj `json:"callbackObj"`      // Cancellation operation results
}

// CancellationCallbackObj contains cancellation operation results.
type CancellationCallbackObj struct {
	TransactionReferenceNumber string `json:"TransactionReferenceNumber"` // Reference number for the cancellation transaction
}

// DoubleInsuranceRequest represents a request to validate for duplicate insurance coverage.
type DoubleInsuranceRequest struct {
	PolicyStartDate           string `json:"policystartdate"`           // Policy start date
	PolicyEndDate             string `json:"policyenddate"`             // Policy end date
	VehicleRegistrationNumber string `json:"vehicleregistrationnumber"` // Vehicle registration number
	ChassisNumber             string `json:"chassisnumber"`             // Vehicle chassis number
}

// DoubleInsuranceResponse represents the response from double insurance validation operations.
type DoubleInsuranceResponse struct {
	Inputs           string                     `json:"Inputs"`           // Original request parameters as string
	CallbackObj      DoubleInsuranceCallbackObj `json:"callbackObj"`      // Double insurance validation results
	Error            FlexibleDmvicError         `json:"error,omitempty"`  // Error details if operation failed
	Success          bool                       `json:"success"`          // Indicates if the operation was successful
	APIRequestNumber string                     `json:"apiRequestNumber"` // Unique API request identifier
}

// DoubleInsuranceList is a flexible type that can unmarshal from either an
// array of DoubleInsuranceDetails or an object/map representation. Internally
// it stores a slice for predictable iteration.
type DoubleInsuranceList []DoubleInsuranceDetails

// UnmarshalJSON supports three shapes:
// - JSON array: [{...}, {...}]
// - JSON object/map: {"key": {...}, ...}
// - Single object: {...}
func (d *DoubleInsuranceList) UnmarshalJSON(data []byte) error {
	// Try as array
	var arr []DoubleInsuranceDetails
	if err := json.Unmarshal(data, &arr); err == nil {
		*d = arr
		return nil
	}

	// Try as map[string]DoubleInsuranceDetails
	var m map[string]DoubleInsuranceDetails
	if err := json.Unmarshal(data, &m); err == nil {
		res := make([]DoubleInsuranceDetails, 0, len(m))
		for _, v := range m {
			res = append(res, v)
		}
		*d = res
		return nil
	}

	// Try single object
	var single DoubleInsuranceDetails
	if err := json.Unmarshal(data, &single); err == nil {
		*d = []DoubleInsuranceDetails{single}
		return nil
	}

	return fmt.Errorf("DoubleInsurance: unsupported JSON format")
}

// DoubleInsuranceCallbackObj contains double insurance validation results.
type DoubleInsuranceCallbackObj struct {
	DoubleInsurance DoubleInsuranceList `json:"doubleInsurance"` // Flexible list of double insurance details
}

// UnmarshalJSON makes DoubleInsuranceCallbackObj tolerant to different casing
// of the "doubleInsurance" key (e.g. "DoubleInsurance"). It will look up any
// key that case-insensitively matches "doubleinsurance" and unmarshal its value
// into the DoubleInsuranceList.
func (d *DoubleInsuranceCallbackObj) UnmarshalJSON(data []byte) error {
	// Generic map to find the key regardless of case
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	for k, v := range m {
		if strings.EqualFold(k, "doubleinsurance") {
			var list DoubleInsuranceList
			if err := json.Unmarshal(v, &list); err != nil {
				return fmt.Errorf("failed to unmarshal DoubleInsurance: %w", err)
			}
			d.DoubleInsurance = list
			return nil
		}
	}
	// If not found, try direct unmarshal into structure (in case data is the inner array)
	var direct DoubleInsuranceList
	if err := json.Unmarshal(data, &direct); err == nil {
		d.DoubleInsurance = direct
		return nil
	}
	// No recognized content
	return fmt.Errorf("DoubleInsuranceCallbackObj: missing doubleInsurance key")
}

// BaseIssuanceFields contains common fields for insurance certificate issuance requests.
// It includes vehicle and policyholder information, coverage details, and contact information.
type BaseIssuanceFields struct {
	MemberCompanyID    int    `json:"MemberCompanyID"`    // Identifier for the member company
	TypeOfCover        int    `json:"Typeofcover"`        // Type of coverage (e.g., comprehensive, third-party)
	PolicyHolder       string `json:"Policyholder"`       // Name of the policyholder
	PolicyNumber       string `json:"policynumber"`       // Insurance policy number
	CommencingDate     string `json:"Commencingdate"`     // Policy start date
	ExpiringDate       string `json:"Expiringdate"`       // Policy end date
	RegistrationNumber string `json:"Registrationnumber"` // Vehicle registration number
	ChassisNumber      string `json:"Chassisnumber"`      // Vehicle chassis number
	PhoneNumber        string `json:"Phonenumber"`        // Contact phone number
	BodyType           string `json:"Bodytype"`           // Type of vehicle body (e.g., sedan, SUV)
	VehicleMake        string `json:"Vehiclemake"`        // Make of the vehicle
	VehicleModel       string `json:"Vehiclemodel"`       // Model of the vehicle
	EngineNumber       string `json:"Enginenumber"`       // Engine number of the vehicle
	Email              string `json:"Email"`              // Contact email address
	SumInsured         int    `json:"SumInsured"`         // Total insured amount
	InsuredPIN         string `json:"InsuredPIN"`         // Personal Identification Number of the insured
}

// TypeAIssuanceRequest represents a request for issuing a Type A insurance certificate.
// It includes additional fields specific to Type A certificates, such as the type of certificate and licensing information.
type TypeAIssuanceRequest struct {
	*BaseIssuanceFields `json:",inline"` // Embed base fields
	TypeOfCertificate   int              `json:"TypeOfCertificate"` // Type of certificate (e.g., original, duplicate)
	LicensedToCarry     int              `json:"Licensedtocarry"`   // Indicates if the vehicle is licensed to carry passengers or goods
}

// TypeBIssuanceRequest represents a request for issuing a Type B insurance certificate.
// It includes additional fields specific to Type B certificates, such as vehicle type, tonnage, and licensing information.
type TypeBIssuanceRequest struct {
	*BaseIssuanceFields `json:",inline"` // Embed base fields
	VehicleType         int              `json:"VehicleType"`     // Type of vehicle (e.g., private, commercial)
	Tonnage             int              `json:"Tonnage"`         // Tonnage of the vehicle for commercial vehicles
	LicensedToCarry     int              `json:"Licensedtocarry"` // Indicates if the vehicle is licensed to carry passengers or goods
}

// TypeCIssuanceRequest represents a request for issuing a Type C insurance certificate.
// It includes additional fields specific to Type C certificates.
type TypeCIssuanceRequest struct {
	*BaseIssuanceFields `json:",inline"` // Embed base fields
}

// TypeDIssuanceRequest represents a request for issuing a Type D insurance certificate.
// It includes additional fields specific to Type D certificates, such as the type of certificate, licensing information, and tonnage.
type TypeDIssuanceRequest struct {
	*BaseIssuanceFields
	TypeOfCertificate int `json:"TypeOfCertificate"` // Type of certificate (e.g., original, duplicate)
	LicensedToCarry   int `json:"Licensedtocarry"`   // Indicates if the vehicle is licensed to carry passengers or goods
	Tonnage           int `json:"Tonnage"`           // Tonnage of the vehicle for commercial vehicles
}

// DmvicError represents an error response from the DMVIC API.
// It contains error code and error text providing details about the error.
type DmvicError struct {
	ErrorCode string `json:"errorCode"` // Error code indicating the type of error
	ErrorText string `json:"errorText"` // Descriptive error message
}

// FlexibleDmvicError is a slice of DmvicError, allowing for multiple error details to be returned.
type FlexibleDmvicError []DmvicError

// InsuranceResponse represents the response from insurance certificate issuance requests.
// It contains details about the issued certificate or errors encountered during the process.
type InsuranceResponse struct {
	Inputs           interface{}         `json:"Inputs"`           // Original request parameters
	Error            FlexibleDmvicError  `json:"Error,omitempty"`  // Error details if operation failed
	Success          bool                `json:"success"`          // Indicates if the operation was successful
	APIRequestNumber string              `json:"apiRequestNumber"` // Unique API request identifier
	CallbackObj      IssuanceCallbackObj `json:"CallbackObj"`      // Issuance operation results
}

// IssuanceCallbackObj contains the results of the insurance certificate issuance operation.
// It includes details about the issued certificate such as transaction number, actual certificate number, and email.
type IssuanceCallbackObj struct {
	IssueCertificate IssuanceDetails `json:"issueCertificate"` // Details of the issued certificate
}

// IssuanceDetails contains detailed information about an issued insurance certificate.
// It includes transaction number, actual certificate number, and email of the certificate holder.
type IssuanceDetails struct {
	TransactionNo string `json:"TransactionNo"` // Transaction number for the issuance operation
	ActualCNo     string `json:"actualCNo"`     // Actual certificate number issued
	Email         string `json:"Email"`         // Email of the certificate holder
}

// StockResponse represents the response from stock retrieval operations.
// It contains details about the stock of insurance certificates available for issuance.
type StockResponse struct {
	CallbackObj      StockCallbackObj   `json:"callbackObj"`      // Stock information
	Error            FlexibleDmvicError `json:"error,omitempty"`  // Error details if operation failed
	Success          bool               `json:"success"`          // Indicates if the operation was successful
	APIRequestNumber string             `json:"apiRequestNumber"` // Unique API request identifier
}

// StockCallbackObj contains stock information for insurance certificates.
// It includes a list of stock details for each member company.
type StockCallbackObj struct {
	MemberCompanyStock []StockDetails `json:"MemberCompanyStock"` // List of stock details for each member company
}

// StockDetails contains information about the stock of a specific insurance certificate.
// It includes certificate classification ID, title, stock quantity, and certificate type ID.
type StockDetails struct {
	CertificateClassificationID int    `json:"CertificateClassificationID"` // Identifier for the certificate classification
	ClassificationTitle         string `json:"ClassificationTitle"`         // Title of the certificate classification
	Stock                       int    `json:"Stock"`                       // Quantity of certificates available in stock
	CertificateTypeID           int    `json:"CertificateTypeId"`           // Identifier for the certificate type
}

// ConfirmationRequest represents a request to confirm an insurance certificate issuance.
// It includes details about the issuance request ID, approval status, verification statuses, comments, and username.
type ConfirmationRequest struct {
	IssuanceRequestID  string `json:"IssuanceRequestID"`  // Identifier of the issuance request to confirm
	IsApproved         bool   `json:"IsApproved"`         // Approval status of the issuance request
	IsLogBookVerified  bool   `json:"IsLogBookVerified"`  // Log book verification status
	IsVehicleInspected bool   `json:"IsVehicleInspected"` // Vehicle inspection status
	AdditionalComments string `json:"AdditionalComments"` // Any additional comments
	UserName           string `json:"UserName"`           // Username of the person confirming the request
}
