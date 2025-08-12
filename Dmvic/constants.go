package dmvic

// Constants for cover types, cancel reasons, certificate types, and vehicle types

// Cover Types
const (
	CoverTypeComprehensive = 100 // COMP
	CoverTypeThirdParty    = 200 // TPO
	CoverTypeTPTF          = 300 // Third-party, Theft & Fire
)

// Cancel Reasons
const (
	CancelReasonInsuredRequest      = 8
	CancelReasonAmendPassengers     = 12
	CancelReasonChangeScopeOfCover  = 13
	CancelReasonPolicyNotTaken      = 14
	CancelReasonVehicleSold         = 15
	CancelReasonAmendInsuredDetails = 18
	CancelReasonAmendVehicleDetails = 19
	CancelReasonSuspectedFraud      = 20
	CancelReasonNonPayment          = 21
	CancelReasonFailureToProvideKYC = 24
	CancelReasonGovernmentRequest   = 25
	CancelReasonSubjectMatterCeased = 26
	CancelReasonChangePeriod        = 27
	CancelReasonCoverDeclined       = 28
	CancelReasonVehicleWrittenOff   = 29
	CancelReasonVehicleStolen       = 30
)

// Certificate Types
const (
	CertTypeClassAPSVUnmarked   = 1
	CertTypeTypeATaxi           = 8
	CertTypeTypeDMotorCycle     = 4
	CertTypeTypeDPSVMotorCycle  = 9
	CertTypeTypeDMotorCycleComm = 10
)

// Vehicle Types (Type B)
const (
	VehicleTypeOwnGoods       = 1
	VehicleTypeGeneralCartage = 2
	VehicleTypeInstitutional  = 3
	VehicleTypeSpecial        = 4
	VehicleTypeTankers        = 5
	VehicleTypeMotorTrade     = 6
)
