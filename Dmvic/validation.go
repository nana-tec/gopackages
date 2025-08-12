package dmvic

import "fmt"

// ValidateTypeARequest validates a Type A certificate issuance request
func ValidateTypeARequest(req *TypeAIssuanceRequest) error {
	if req.MemberCompanyID <= 0 {
		return fmt.Errorf("MemberCompanyID is required")
	}
	if req.TypeOfCertificate != CertTypeClassAPSVUnmarked && req.TypeOfCertificate != CertTypeTypeATaxi {
		return fmt.Errorf("invalid TypeOfCertificate for Type A: %d", req.TypeOfCertificate)
	}
	if req.TypeOfCover != CoverTypeComprehensive && req.TypeOfCover != CoverTypeThirdParty && req.TypeOfCover != CoverTypeTPTF {
		return fmt.Errorf("invalid TypeOfCover: %d", req.TypeOfCover)
	}
	if (req.TypeOfCover == CoverTypeComprehensive || req.TypeOfCover == CoverTypeTPTF) && req.SumInsured <= 0 {
		return fmt.Errorf("SumInsured is required for COMP and TPTF cover types")
	}
	if req.PolicyHolder == "" {
		return fmt.Errorf("Policyholder is required")
	}
	if req.PolicyNumber == "" {
		return fmt.Errorf("PolicyNumber is required")
	}
	if req.RegistrationNumber == "" {
		return fmt.Errorf("RegistrationNumber is required")
	}
	if req.ChassisNumber == "" {
		return fmt.Errorf("ChassisNumber is required")
	}
	return nil
}

// ValidateTypeBRequest validates a Type B certificate issuance request
func ValidateTypeBRequest(req *TypeBIssuanceRequest) error {
	if req.MemberCompanyID <= 0 {
		return fmt.Errorf("MemberCompanyID is required")
	}
	if req.VehicleType < VehicleTypeOwnGoods || req.VehicleType > VehicleTypeMotorTrade {
		return fmt.Errorf("invalid VehicleType: %d", req.VehicleType)
	}
	if req.TypeOfCover != CoverTypeComprehensive && req.TypeOfCover != CoverTypeThirdParty && req.TypeOfCover != CoverTypeTPTF {
		return fmt.Errorf("invalid TypeOfCover: %d", req.TypeOfCover)
	}
	if (req.TypeOfCover == CoverTypeComprehensive || req.TypeOfCover == CoverTypeTPTF) && req.SumInsured <= 0 {
		return fmt.Errorf("SumInsured is required for COMP and TPTF cover types")
	}
	if req.PolicyHolder == "" {
		return fmt.Errorf("Policyholder is required")
	}
	if req.PolicyNumber == "" {
		return fmt.Errorf("PolicyNumber is required")
	}
	if req.RegistrationNumber == "" {
		return fmt.Errorf("RegistrationNumber is required")
	}
	if req.ChassisNumber == "" {
		return fmt.Errorf("ChassisNumber is required")
	}
	return nil
}

// ValidateTypeCRequest validates a Type C certificate issuance request
func ValidateTypeCRequest(req *TypeCIssuanceRequest) error {
	if req.MemberCompanyID <= 0 {
		return fmt.Errorf("MemberCompanyID is required")
	}
	if req.TypeOfCover != CoverTypeComprehensive && req.TypeOfCover != CoverTypeThirdParty && req.TypeOfCover != CoverTypeTPTF {
		return fmt.Errorf("invalid TypeOfCover: %d", req.TypeOfCover)
	}
	if (req.TypeOfCover == CoverTypeComprehensive || req.TypeOfCover == CoverTypeTPTF) && req.SumInsured <= 0 {
		return fmt.Errorf("SumInsured is required for COMP and TPTF cover types")
	}
	if req.PolicyHolder == "" {
		return fmt.Errorf("Policyholder is required")
	}
	if req.PolicyNumber == "" {
		return fmt.Errorf("PolicyNumber is required")
	}
	if req.RegistrationNumber == "" {
		return fmt.Errorf("RegistrationNumber is required")
	}
	if req.ChassisNumber == "" {
		return fmt.Errorf("ChassisNumber is required")
	}
	return nil
}

// ValidateTypeDRequest validates a Type D certificate issuance request
func ValidateTypeDRequest(req *TypeDIssuanceRequest) error {
	if req.MemberCompanyID <= 0 {
		return fmt.Errorf("MemberCompanyID is required")
	}
	if req.TypeOfCertificate != CertTypeTypeDMotorCycle &&
		req.TypeOfCertificate != CertTypeTypeDPSVMotorCycle &&
		req.TypeOfCertificate != CertTypeTypeDMotorCycleComm {
		return fmt.Errorf("invalid TypeOfCertificate for Type D: %d", req.TypeOfCertificate)
	}
	if req.TypeOfCover != CoverTypeComprehensive && req.TypeOfCover != CoverTypeThirdParty && req.TypeOfCover != CoverTypeTPTF {
		return fmt.Errorf("invalid TypeOfCover: %d", req.TypeOfCover)
	}
	if (req.TypeOfCover == CoverTypeComprehensive || req.TypeOfCover == CoverTypeTPTF) && req.SumInsured <= 0 {
		return fmt.Errorf("SumInsured is required for COMP and TPTF cover types")
	}
	if req.PolicyHolder == "" {
		return fmt.Errorf("Policyholder is required")
	}
	if req.PolicyNumber == "" {
		return fmt.Errorf("PolicyNumber is required")
	}
	if req.RegistrationNumber == "" {
		return fmt.Errorf("RegistrationNumber is required")
	}
	if req.ChassisNumber == "" {
		return fmt.Errorf("ChassisNumber is required")
	}
	return nil
}
