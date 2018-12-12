package mmgpio

import (
	"os"
	"reflect"
	"syscall"
	"unsafe"
)

type BoardType int

const (
	RASP_ZERO BoardType = iota
	RASP_2_3
)

//RaspMMGPIO represents the memory mapped GPIOs of the Radpberry Pi
type RaspMMGPIO struct {
	MMFilename string
	GPIOOffset int
	MMPageSize int
	mmfile     *os.File
	gpios      []int
}

//NewRaspMMGPIO returns a MMGPIO object corresponding to the board type
//OFFSET for RASP_ZERO 0x20200000 = 0x20000000 (peripheral offset) + 0x200000 (gpio offset)
//SIZE   4*1024 = we only care about the first 40 (10 *4) Bytes, but mapping the whole page anyhow (believe it is more efficient)
func NewRaspMMGPIO(rasp BoardType) *RaspMMGPIO {
	if rasp == RASP_2_3 {
		return &RaspMMGPIO{"/dev/mem", int(0x3f200000), 4 * 1024, nil, []int{}}
	}
	return &RaspMMGPIO{"/dev/mem", int(0x20200000), 4 * 1024, nil, []int{}}
}

//SetFilename overwrites the memory file (so you can use memgpio instead)
func (r *RaspMMGPIO) SetFilename(in string) {
	r.MMFilename = in
}

//Init opens the memory file and memory maps it at the given offset and pagesize
func (r *RaspMMGPIO) Init() error {
	var err error
	r.mmfile, err = os.OpenFile(r.MMFilename, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	//Magic number explanation
	//OFFSET 0x20200000 = 0x20000000 (peripheral offset) + 0x200000 (gpio offset)
	//SIZE   4*1024 = we only care about the first 40 Bytes, but mapping the whole page anyhow (believe it is more efficient)

	tmp, err := syscall.Mmap(int(r.mmfile.Fd()), r.GPIOOffset, r.MMPageSize, syscall.PROT_WRITE|syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		r.mmfile.Close()
		return err
	}

	//convert the []byte to []int
	r.gpios = bytesToInts(tmp)

	return nil
}

//Deinit deinitializes mmap and file
func (r *RaspMMGPIO) DeInit() error {
	orig := intsToBytes(r.gpios)
	err := syscall.Munmap(orig)
	if err != nil {
		return err
	}
	err = r.mmfile.Close()
	if err != nil {
		return err
	}
	return nil
}

//OutGpio mimics the macro below
//#define OUT_GPIO(g)   *(gpio.addr + ((g)/10)) |=  (1<<(((g)%10)*3))
func (r *RaspMMGPIO) OutGpio(g int) {
	r.gpios[(g)/10] |= 1 << uint((g%10)*3)
}

//SetGpio mimics the c macro below
//#define GPIO_SET  *(gpio.addr + 7)  // sets   bits which are 1 ignores bits which are 0
func (r *RaspMMGPIO) SetGpio(g int) {
	r.gpios[7] = 1 << uint(g)
}

//ClrGpio mimics the c macro below
//#define GPIO_CLR  *(gpio.addr + 10)  // sets   bits which are 1 ignores bits which are 0
func (r *RaspMMGPIO) ClrGpio(g int) {
	r.gpios[10] = 1 << uint(g)
}

//Converts a byte slice to an int slice without
//touching the internal data
//will only work on 32 bit machines
func bytesToInts(b []byte) []int {
	s := &reflect.SliceHeader{}
	s.Len = len(b) / 4
	s.Cap = len(b) / 4
	s.Data = (uintptr)(unsafe.Pointer(&b[0]))
	return *(*[]int)(unsafe.Pointer(s))

}

//Converts a int slice to a byte slice without
//touching the internal data
//will only work on 32 bit machines
func intsToBytes(i []int) []byte {
	s := &reflect.SliceHeader{}
	s.Len = len(i) * 4
	s.Cap = len(i) * 4
	s.Data = (uintptr)(unsafe.Pointer(&i[0]))
	return *(*[]byte)(unsafe.Pointer(s))
}
