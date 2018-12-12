package mmgpio

//MMGPIO represents a boards memorymapped GPIOs
type MMGPIO interface {
	Init() error
	SetFilename(in string)
	DeInit() error
	OutGpio(gpio int)
	SetGpio(gpio int)
	ClrGpio(gpio int)
}
