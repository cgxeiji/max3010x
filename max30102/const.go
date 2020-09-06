package max30102

// Register addresses
const (
	IntStat1         = 0x00
	IntStat2         = 0x01
	IntEna1          = 0x02
	IntEna2          = 0x03
	FIFOWrPtr        = 0x04
	OvfCount         = 0x05
	FIFORdPtr        = 0x06
	FIFOData         = 0x07
	FIFOCfg          = 0x08
	ModeCfg          = 0x09
	SpO2Cfg          = 0x0A
	Led1PA           = 0x0C
	Led2PA           = 0x0D
	MultiLedModeS2S1 = 0x11
	MultiLedModeS4S3 = 0x12
	TempInt          = 0x1F
	TempFrac         = 0x20
	TempCfg          = 0x21
	RegRevID         = 0xFE
	RegPartID        = 0xFF
)

// Interrupt flags
const (
	// Status 1
	AlmostFull            byte = (1 << 7)
	NewFIFOData           byte = (1 << 6)
	AmbientLightCancelOvf byte = (1 << 5)
	PowerReady            byte = (1 << 0)

	// Status 2
	DieTempReady byte = (1 << 1)
)

// Device constants
const (
	Addr   = 0x57
	PartID = 0x15
)

// Settings
const (
	TempEna      byte = 0b0000_0001
	ModeHR       byte = 0b010
	ModeSpO2     byte = 0b011
	ModeMultiLed byte = 0b111
	modeMask     byte = 0b1111_1000

	ResetControl = 0b0100_0000
)

// SpO2 Sample Rate Control
const (
	SR50 = (iota << 2)
	SR100
	SR200
	SR400
	SR800
	SR1000
	SR1600
	SR3200

	srMask byte = 0b1_11_000_11
)

// LED Pulse Width Control
const (
	PW69 = iota
	PW118
	PW215
	PW411

	pwMask byte = 0b1_11_111_00
)
