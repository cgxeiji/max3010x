package max3010x

const (
	maxIntStatus  = 0x00
	maxIntEnable  = 0x01
	maxFifoWrPtr  = 0x02
	maxOvfCounter = 0x03
	maxFifoRdPtr  = 0x04
	maxFifoData   = 0x05
	maxModeCfg    = 0x06
	maxSpO2Cfg    = 0x07
	maxLedCfg     = 0x09
	maxTempInt    = 0x16
	maxTempFrac   = 0x17
	maxRevID      = 0xFE
	maxPartID     = 0xFF

	maxID00 = 0x11
	maxID02 = 0x15

	maxAddr = 0x57
)

const (
	sr50 = iota
	sr100
	sr167
	sr200
	sr400
	sr600
	sr800
	sr1000
)

const (
	modeHR   = 0b0000010
	modeSPO2 = 0b0000011
	modeTemp = 0b0001000
	modeRST  = 0b0100000
	modeSHDN = 0b1000000
)

const (
	pw200 = iota
	pw400
	pw800
	pw1600
)

const (
	mA0 = iota
	mA44
	mA76
	mA110
	mA142
	mA174
	mA208
	mA24
	mA271
	mA306
	mA338
	mA370
	mA402
	mA436
	mA468
	mA500
)
