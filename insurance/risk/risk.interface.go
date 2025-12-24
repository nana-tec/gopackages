package risk

import (
	"context"
	"fmt"
)

type VehicleType string
type BodyType string

const (
	PSVBus                     VehicleType = "PSV-BUS"
	PSVMatatu                  VehicleType = "PSV-MATATU"
	PSVTaxi                    VehicleType = "PSV-TAXI"
	PSVPrivateHire             VehicleType = "PSV-PRIVATE HIRE"
	Private                    VehicleType = "PRIVATE"
	MotorCommercialOwnGoods    VehicleType = "MOTOR COMMERCIAL:OWN GOODS"
	MotorCommercialInstitution VehicleType = "MOTOR COMMERCIAL:INSTITUTIONAL"
	MotorCommercialPrimeMover  VehicleType = "MOTOR COMMERCIAL:Prime  Mover"
	MotorCommercialTrailer     VehicleType = "MOTOR COMMERCIAL:Trailer"
	MotorCommercialTankers     VehicleType = "MOTOR COMMERCIAL:Tankers & Specified Trailers"
	MotorCyclePrivate          VehicleType = "MOTOR CYCLE: PRIVATE"
	MotorCyclePSV              VehicleType = "MOTOR CYCLE: PSV"
	MotorCommercialCartage     VehicleType = "MOTOR COMMERCIAL:GENERAL CARTAGE"
	MotorCommercialTractor     VehicleType = "MOTOR COMMERCIAL:TRACTOR"
)

func (v VehicleType) IsValid() bool {
	switch v {
	case PSVBus, PSVMatatu, PSVTaxi, PSVPrivateHire, Private,
		MotorCommercialOwnGoods, MotorCommercialInstitution, MotorCommercialPrimeMover,
		MotorCommercialTrailer, MotorCommercialTankers, MotorCyclePrivate, MotorCyclePSV,
		MotorCommercialCartage, MotorCommercialTractor:
		return true
	}
	return false
}
func (v VehicleType) String() string {
	return string(v)
}

var VehicleTypeMap = map[VehicleType]int{
	PSVBus:                     1,
	PSVMatatu:                  2,
	PSVTaxi:                    3,
	PSVPrivateHire:             4,
	Private:                    5,
	MotorCommercialOwnGoods:    6,
	MotorCommercialInstitution: 7,
	MotorCommercialPrimeMover:  8,
	MotorCommercialTrailer:     9,
	MotorCommercialTankers:     10,
	MotorCyclePrivate:          11,
	MotorCyclePSV:              12,
	MotorCommercialCartage:     13,
	MotorCommercialTractor:     14,
}

var ToVehicleTypeMap = map[int]VehicleType{
	1:  PSVBus,
	2:  PSVMatatu,
	3:  PSVTaxi,
	4:  PSVPrivateHire,
	5:  Private,
	6:  MotorCommercialOwnGoods,
	7:  MotorCommercialInstitution,
	8:  MotorCommercialPrimeMover,
	9:  MotorCommercialTrailer,
	10: MotorCommercialTankers,
	11: MotorCyclePrivate,
	12: MotorCyclePSV,
	13: MotorCommercialCartage,
	14: MotorCommercialTractor,
}

const (
	Bus                BodyType = "Bus"
	StationWagon       BodyType = "Station Wagon"
	PickUp             BodyType = "Pick-up"
	Van                BodyType = "Van"
	MiniBus            BodyType = "Mini-Bus"
	Saloon             BodyType = "Saloon"
	SelfDriveSW        BodyType = "SelfDrive:SW"
	SelfDrivePU        BodyType = "SelfDrive:PU"
	SelfDriveVan       BodyType = "SelfDrive: Van"
	SelfDriveMBus      BodyType = "SelfDrive:MBus"
	SelfDriveBus       BodyType = "SelfDrive: Bus"
	ChauffeurSW        BodyType = "Chauffeur:SW"
	ChauffeurPU        BodyType = "Chauffeur:PU"
	ChauffeurVan       BodyType = "Chauffeur:Van"
	ChauffeurMBus      BodyType = "Chauffeur:MBus"
	ChauffeurBus       BodyType = "Chauffeur: Bus"
	ChauffeurTV        BodyType = "Chauffeur: TV"
	SUV                BodyType = "SUV"
	SubaruSW           BodyType = `Subaru: S\Wagon`
	SubaruSaloon       BodyType = "Subaru: Saloon"
	OldSW              BodyType = "> 15 Years S/W"
	Truck              BodyType = "Truck"
	PrimeMover         BodyType = "Prime  Mover"
	Trailer            BodyType = "Trailer"
	Tanker             BodyType = "Tanker"
	LPGTanker          BodyType = "LPG TANKER"
	PetroleumTanker    BodyType = "PETROLEUM TANKER"
	MotorCycle         BodyType = "Motor Cycle"
	ElectricMotorCycle BodyType = "Electric Motor Cycle"
	SamnelTruck        BodyType = "Samnel: Truck"
	Tractor            BodyType = "Tractor"
	PSVMatatuPickup    BodyType = "Pick up"
)

func (b BodyType) IsValid() bool {
	switch b {
	case Bus, StationWagon, PickUp, Van, MiniBus, Saloon,
		SelfDriveSW, SelfDrivePU, SelfDriveVan, SelfDriveMBus, SelfDriveBus,
		ChauffeurSW, ChauffeurPU, ChauffeurVan, ChauffeurMBus, ChauffeurBus, ChauffeurTV,
		SUV, SubaruSW, SubaruSaloon, OldSW, Truck, PrimeMover, Trailer, Tanker,
		LPGTanker, PetroleumTanker, MotorCycle, ElectricMotorCycle, SamnelTruck, Tractor, PSVMatatuPickup:
		return true
	}

	return false
}

func (b BodyType) String() string {
	return string(b)
}

var VehicleTypeToBodyType = map[int][]string{
	1: {"Bus"},
	2: {"Station Wagon", "Pick up", "Van", "Mini-Bus"},
	3: {"Station Wagon", "Saloon"},
	4: {
		"SelfDrive:SW", "SelfDrive:PU", "SelfDrive: Van", "SelfDrive:MBus", "SelfDrive: Bus",
		"Chauffeur:SW", "Chauffeur:PU", "Chauffeur:Van", "Chauffeur:MBus", "Chauffeur: Bus", "Chauffeur: TV",
	},
	5:  {"Station Wagon", "Saloon", "SUV", "Pick-up", `Subaru: S\Wagon`, "Subaru: Saloon", "> 15 Years S/W"},
	6:  {"Pick-up", "Truck", "Van"},
	7:  {"Pick-up", "Truck", "Van", "Bus", "Saloon"},
	8:  {"Prime Mover"},
	9:  {"Trailer"},
	10: {"Trailer", "Tanker", "LPG TANKER", "PETROLEUM TANKER"},
	11: {"Motor Cycle"},
	12: {"Motor Cycle", "Electric Motor Cycle"},
	13: {"Pick-up", "Truck", "Van", "Samnel: Truck"},
	14: {"Tractor"},
}

func ValidateBodyTypeAgainstVehicleType(vehicleType int, bodyType string) (string, error) {
	bodyTypes, ok := VehicleTypeToBodyType[vehicleType]
	if !ok {
		return "", fmt.Errorf("vehicle type %d not found in VehicleTypeToBodyType map", vehicleType)
	}

	for _, b := range bodyTypes {
		if b == bodyType {
			return b, nil
		}
	}

	return "", fmt.Errorf("invalid body type %s for vehicle type %v", bodyType, ToVehicleTypeMap[vehicleType])
}

type MotorRiskModel struct {
	RegistrationNumber string      `json:"registration_number" bson:"registration_number" `
	ChassisNumber      string      `json:"chassis_number" bson:"chassis_number" `
	CarMake            string      `json:"car_make" bson:"car_make" `
	CarModel           string      `json:"car_model" bson:"car_model" `
	SeatingCapacity    int         `json:"seating_capacity" bson:"sitting_capacity" `
	Tonnage            float64     `json:"tonnage" bson:"tonnage" `
	YearOfManufacture  string      `json:"year_of_manufacture" bson:"year_of_manufacture"`
	CubicCapacity      string      `json:"cubic_capacity" bson:"cubic_capacity"`
	VehicleType        VehicleType `json:"vehicle_type" bson:"vehicle_type"`
	BodyType           BodyType    `json:"body_type" bson:"body_type"`
	NameOfSacco        string      `json:"name_of_sacco" bson:"name_of_sacco"`
	RiskSystemRef      string      `json:"risk_system_ref" bson:"risk_system_ref"`
}

type MotorRisk struct {
	RegistrationNumber string
	ChassisNumber      string
	CarMake            string
	CarModel           string
	SeatingCapacity    int
	Tonnage            float64
	YearOfManufacture  string
	CubicCapacity      string
	VehicleType        VehicleType
	BodyType           BodyType
	NameOfSacco        string
}

type RiskRepository interface {

	// GetMotorRisk returns a MotorRisk by registration number
	GetMotorRiskByRegistrationNumber(ctx context.Context, registrationNumber string) (*MotorRiskModel, error)

	// GetMotorRisk returns a MotorRisk by chassis number
	GetMotorRiskByChassisNumber(ctx context.Context, chassisNumber string) (*MotorRiskModel, error)

	GetMotorRiskByRegistrationNumberOrChassis(ctx context.Context, registrationNumber string, chassisNumber string) (*MotorRiskModel, error)
	// GetMotorRisk returns a MotorRisk by risk system ref
	GetMotorRiskByRiskSystemRef(ctx context.Context, riskSystemRef string) (*MotorRiskModel, error)

	GetMotorRiskByRef(ctx context.Context, riskRef string) (*MotorRiskModel, error)

	// SaveMotorRisk saves a MotorRisk
	SaveMotorRisk(ctx context.Context, motorRisk *MotorRiskModel) error

	// UpdateMotorRisk updates a MotorRisk
	UpdateMotorRisk(ctx context.Context, motorRisk *MotorRiskModel) error

	// DeleteMotorRisk deletes a MotorRisk
	DeleteMotorRisk(ctx context.Context, motorRisk *MotorRiskModel) error
}

type riskValidateDoubleInsuranceResponse struct {
	IsInsured         bool
	ExistingPolicyRef string
	UnderwriterName   string
}
type RiskUsecase interface {
	CreateUpdateRisk(ctx context.Context, motorRisk *MotorRisk) (string, error)
	ValidateRiskDoubleInsurance(ctx context.Context, riskRef string, PolicyStartDate string, PolicyEndDate string) (riskValidateDoubleInsuranceResponse, error)
	GetRiskByRef(ctx context.Context, riskRef string) (*MotorRiskModel, error)
	UpdateRisk(ctx context.Context, motorRisk *MotorRiskModel) error
}
