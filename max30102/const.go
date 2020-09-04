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

// Device constants
const (
	Addr   = 0x57
	PartID = 0x15
)

// Settings
const (
	TempEna = 0b0000_0001
)
