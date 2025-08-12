package dmvic

import "fmt"

// Utility functions for human-readable descriptions

func GetCancelReasonDescription(reasonID int) string {
	switch reasonID {
	case CancelReasonInsuredRequest:
		return "Insured person requested cancellation"
	case CancelReasonAmendPassengers:
		return "Amending no of passengers"
	case CancelReasonChangeScopeOfCover:
		return "Change of scope of cover"
	case CancelReasonPolicyNotTaken:
		return "Policy Not taken up"
	case CancelReasonVehicleSold:
		return "Vehicle sold"
	case CancelReasonAmendInsuredDetails:
		return "Amending Insured's Details"
	case CancelReasonAmendVehicleDetails:
		return "Amending vehicle details"
	case CancelReasonSuspectedFraud:
		return "Suspected Fraud"
	case CancelReasonNonPayment:
		return "Non-payment of premium"
	case CancelReasonFailureToProvideKYC:
		return "Failure to provide KYCs"
	case CancelReasonGovernmentRequest:
		return "Request by a government body"
	case CancelReasonSubjectMatterCeased:
		return "Subject matter ceased to exist"
	case CancelReasonChangePeriod:
		return "Change Period of Insurance"
	case CancelReasonCoverDeclined:
		return "Cover declined by Insurer"
	case CancelReasonVehicleWrittenOff:
		return "Motor Vehicle was written off"
	case CancelReasonVehicleStolen:
		return "Motor Vehicle was stolen"
	default:
		return fmt.Sprintf("Unknown cancel reason: %d", reasonID)
	}
}

func GetCoverTypeDescription(coverType int) string {
	switch coverType {
	case CoverTypeComprehensive:
		return "Comprehensive (COMP)"
	case CoverTypeThirdParty:
		return "Third-party (TPO)"
	case CoverTypeTPTF:
		return "Third-party, Theft & Fire (TPTF)"
	default:
		return fmt.Sprintf("Unknown cover type: %d", coverType)
	}
}

func GetVehicleTypeDescription(vehicleType int) string {
	switch vehicleType {
	case VehicleTypeOwnGoods:
		return "MOTOR COMMERCIAL OWN GOODS"
	case VehicleTypeGeneralCartage:
		return "MOTOR COMMERCIAL GENERAL CARTAGE"
	case VehicleTypeInstitutional:
		return "MOTOR INSTITUTIONAL VEHICLE"
	case VehicleTypeSpecial:
		return "MOTOR SPECIAL VEHICLES"
	case VehicleTypeTankers:
		return "TANKERS (LIQUID CARRYING)"
	case VehicleTypeMotorTrade:
		return "MOTOR TRADE (ROAD RISK)"
	default:
		return fmt.Sprintf("Unknown vehicle type: %d", vehicleType)
	}
}

func GetCertificateTypeDescription(certType int) string {
	switch certType {
	case CertTypeClassAPSVUnmarked:
		return "Class A - PSV Unmarked"
	case CertTypeTypeATaxi:
		return "Type A Taxi"
	case CertTypeTypeDMotorCycle:
		return "Type D Motor Cycle"
	case CertTypeTypeDPSVMotorCycle:
		return "Type D PSV Motor Cycle"
	case CertTypeTypeDMotorCycleComm:
		return "Type D – Motor Cycle Commercial"
	default:
		return fmt.Sprintf("Unknown certificate type: %d", certType)
	}
}
