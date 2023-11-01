package domain

import (
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/platform"
	"github.com/utilitywarehouse/energy-pkg/domain"
)

// Occupancy contains all the data and rules for evaluating the eligibility of an occupancy.
type Occupancy struct {
	ID               string
	Account          Account
	Site             *Site
	Services         []Service
	EvaluationResult OccupancyEvaluation
}

// Account customer account of the occupancy.
type Account struct {
	ID       string
	OptOut   bool
	PSRCodes []string
}

type Site struct {
	ID          string
	Postcode    string
	WanCoverage bool
}

type Meterpoint struct {
	Mpxn         string
	AltHan       bool
	ProfileClass platform.ProfileClass
	SSC          string
}

type Meter struct {
	ID         string
	Mpxn       string
	MSN        string
	SupplyType domain.SupplyType
	Capacity   *float32
	MeterType  string
}

type Service struct {
	ID               string
	Mpxn             string
	SupplyType       domain.SupplyType
	Meterpoint       *Meterpoint
	Meter            *Meter
	BookingReference string
}

type OccupancyEvaluation struct {
	OccupancyID              string
	EligibilityEvaluated     bool
	Eligibility              IneligibleReasons
	SuppliabilityEvaluated   bool
	Suppliability            IneligibleReasons
	CampaignabilityEvaluated bool
	Campaignability          IneligibleReasons
}

var unsupportedSSCs = map[string]bool{
	"0003": true,
	"0006": true,
	"0007": true,
	"0008": true,
	"0009": true,
	"0011": true,
	"0013": true,
	"0015": true,
	"0016": true,
	"0020": true,
	"0022": true,
	"0023": true,
	"0024": true,
	"0025": true,
	"0026": true,
	"0028": true,
	"0029": true,
	"0030": true,
	"0032": true,
	"0033": true,
	"0034": true,
	"0035": true,
	"0036": true,
	"0037": true,
	"0038": true,
	"0042": true,
	"0043": true,
	"0044": true,
	"0046": true,
	"0049": true,
	"0050": true,
	"0051": true,
	"0052": true,
	"0055": true,
	"0056": true,
	"0057": true,
	"0058": true,
	"0063": true,
	"0064": true,
	"0065": true,
	"0066": true,
	"0067": true,
	"0071": true,
	"0072": true,
	"0079": true,
	"0081": true,
	"0082": true,
	"0083": true,
	"0084": true,
	"0085": true,
	"0086": true,
	"0093": true,
	"0095": true,
	"0099": true,
	"0100": true,
	"0101": true,
	"0102": true,
	"0103": true,
	"0104": true,
	"0105": true,
	"0108": true,
	"0109": true,
	"0110": true,
	"0112": true,
	"0113": true,
	"0115": true,
	"0116": true,
	"0117": true,
	"0118": true,
	"0120": true,
	"0121": true,
	"0122": true,
	"0129": true,
	"0136": true,
	"0140": true,
	"0141": true,
	"0142": true,
	"0143": true,
	"0149": true,
	"0154": true,
	"0159": true,
	"0178": true,
	"0228": true,
	"0231": true,
	"0242": true,
	"0243": true,
	"0246": true,
	"0251": true,
	"0252": true,
	"0254": true,
	"0257": true,
	"0259": true,
	"0260": true,
	"0261": true,
	"0262": true,
	"0264": true,
	"0265": true,
	"0266": true,
	"0267": true,
	"0268": true,
	"0269": true,
	"0270": true,
	"0271": true,
	"0272": true,
	"0274": true,
	"0276": true,
	"0277": true,
	"0278": true,
	"0281": true,
	"0283": true,
	"0284": true,
	"0288": true,
	"0300": true,
	"0310": true,
	"0312": true,
	"0313": true,
	"0316": true,
	"0317": true,
	"0318": true,
	"0319": true,
	"0320": true,
	"0321": true,
	"0322": true,
	"0323": true,
	"0324": true,
	"0325": true,
	"0326": true,
	"0327": true,
	"0328": true,
	"0329": true,
	"0330": true,
	"0331": true,
	"0332": true,
	"0334": true,
	"0335": true,
	"0343": true,
	"0346": true,
	"0350": true,
	"0351": true,
	"0353": true,
	"0354": true,
	"0357": true,
	"0358": true,
	"0359": true,
	"0360": true,
	"0361": true,
	"0362": true,
	"0363": true,
	"0364": true,
	"0365": true,
	"0366": true,
	"0367": true,
	"0368": true,
	"0369": true,
	"0370": true,
	"0371": true,
	"0372": true,
	"0373": true,
	"0374": true,
	"0375": true,
	"0376": true,
	"0381": true,
	"0382": true,
	"0386": true,
	"0387": true,
	"0388": true,
	"0389": true,
	"0390": true,
	"0391": true,
	"0392": true,
	"0394": true,
	"0395": true,
	"0396": true,
	"0397": true,
	"0399": true,
	"0400": true,
	"0401": true,
	"0403": true,
	"0405": true,
	"0427": true,
	"0435": true,
	"0436": true,
	"0443": true,
	"0444": true,
	"0447": true,
	"0479": true,
	"0481": true,
	"0702": true,
	"0711": true,
	"0715": true,
	"0717": true,
	"0722": true,
	"0723": true,
	"0724": true,
	"0725": true,
	"0726": true,
	"0727": true,
	"0728": true,
	"0729": true,
	"0730": true,
	"0731": true,
	"0732": true,
	"0733": true,
	"0734": true,
	"0735": true,
	"0736": true,
	"0737": true,
	"0738": true,
	"0739": true,
	"0740": true,
	"0741": true,
	"0742": true,
	"0743": true,
	"0744": true,
	"0745": true,
	"0746": true,
	"0747": true,
	"0748": true,
	"0749": true,
	"0752": true,
	"0753": true,
	"0754": true,
	"0755": true,
	"0756": true,
	"0757": true,
	"0758": true,
	"0759": true,
	"0760": true,
	"0761": true,
	"0762": true,
	"0763": true,
	"0764": true,
	"0765": true,
	"0767": true,
	"0768": true,
	"0769": true,
	"0770": true,
	"0777": true,
	"0803": true,
	"0805": true,
	"0809": true,
	"0811": true,
	"0813": true,
	"0815": true,
	"0817": true,
	"0819": true,
	"0821": true,
	"0823": true,
	"0825": true,
	"0829": true,
	"0833": true,
	"0835": true,
	"0837": true,
	"0839": true,
	"0841": true,
	"0843": true,
	"0845": true,
	"0847": true,
	"0849": true,
	"0851": true,
	"0852": true,
	"0857": true,
	"0858": true,
	"0859": true,
	"0860": true,
	"0861": true,
	"0862": true,
	"0863": true,
	"0864": true,
	"0871": true,
	"0872": true,
	"0890": true,
	"0891": true,
	"0894": true,
	"0895": true,
	"0896": true,
	"0899": true,
	"0901": true,
	"0902": true,
	"0908": true,
	"0911": true,
	"0912": true,
	"0913": true,
	"0914": true,
	"0915": true,
	"0918": true,
	"0934": true,
	"0935": true,
	"0936": true,
	"0937": true,
	"0938": true,
	"0939": true,
	"0942": true,
	"0944": true,
	"0945": true,
	"0946": true,
	"0948": true,
	"0949": true,
	"0950": true,
	"0954": true,
	"0956": true,
	"0967": true,
	"0968": true,
	"0970": true,
	"0971": true,
	"0975": true,
	"0976": true,
	"0980": true,
	"0981": true,
}

func (m Meterpoint) HasComplexTariff() bool {
	if m.ProfileClass == platform.ProfileClass_PROFILE_CLASS_02 ||
		m.ProfileClass == platform.ProfileClass_PROFILE_CLASS_04 {
		if _, found := unsupportedSSCs[m.SSC]; found {
			return true
		}
	}

	return false
}

func (m Meter) IsSmart() bool {
	switch m.SupplyType {
	case domain.SupplyTypeGas:
		switch m.MeterType {
		case platform.MeterTypeGas_METER_TYPE_GAS_SMETS1.String(),
			platform.MeterTypeGas_METER_TYPE_GAS_SMETS2.String():
			return true
		}
		return false
	case domain.SupplyTypeElectricity:
		IsElectricitySmartMeter(m.MeterType)
	}

	return false
}

func IsElectricitySmartMeter(meterType string) bool {
	switch meterType {
	case platform.MeterTypeElec_METER_TYPE_ELEC_SMETS1.String(),
		platform.MeterTypeElec_METER_TYPE_ELEC_S2A.String(),
		platform.MeterTypeElec_METER_TYPE_ELEC_S2B.String(),
		platform.MeterTypeElec_METER_TYPE_ELEC_S2C.String(),
		platform.MeterTypeElec_METER_TYPE_ELEC_S2AD.String(),
		platform.MeterTypeElec_METER_TYPE_ELEC_S2BD.String(),
		platform.MeterTypeElec_METER_TYPE_ELEC_S2CD.String(),
		platform.MeterTypeElec_METER_TYPE_ELEC_S2ADE.String(),
		platform.MeterTypeElec_METER_TYPE_ELEC_S2BDE.String(),
		platform.MeterTypeElec_METER_TYPE_ELEC_S2CDE.String():
		return true
	}
	return false
}
