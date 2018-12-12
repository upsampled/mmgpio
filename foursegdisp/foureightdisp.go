package foursegdisp

import (
	"sync/atomic"
	"time"

	"github.com/upsampled/mmgpio"
)

//segdisp maps a decimal digit to 7 segment display
var segdisp = [][]int{
	{1, 1, 1, 1, 1, 1, 0},
	{0, 1, 1, 0, 0, 0, 0},
	{1, 1, 0, 1, 1, 0, 1},
	{1, 1, 1, 1, 0, 0, 1},
	{0, 1, 1, 0, 0, 1, 1},
	{1, 0, 1, 1, 0, 1, 1},
	{1, 0, 1, 1, 1, 1, 1},
	{1, 1, 1, 0, 0, 0, 0},
	{1, 1, 1, 1, 1, 1, 1},
	{1, 1, 1, 0, 0, 1, 1},
}

//FourEightSegs controls a four digit, eight digit display
//with 7 segments dedicated to a digit and one decidate for a dot
type FourEightSegs struct {
	mmgpio.MMGPIO
	nums [4]uint32
	dots [4]uint32
	segs [7]int
	digs [4]int
	dot  int
	stop uint32
}

//NewFourEightSegs blank constructor, Init does the most work
func NewFourEightSegs(board mmgpio.MMGPIO) *FourEightSegs {
	return &FourEightSegs{MMGPIO: board}
}

//Init initialized the four digit seven seg display, by memory mapping the GPIOS and then
//initializing the gpios that correspond to th display.
//The segs input array contains the gpios corresponding to the segments in this order:
//top, top right, bottom right, bottom, bottom left, top left, middle. The digs input array
//corresonding the the gpio pins that control the lit digits from right to left. The dot input
//corresponds to the gpio pin that controls the dot.
func (m *FourEightSegs) Init(segs [7]int, digs [4]int, dot int) error {
	var err error

	err = m.MMGPIO.Init()
	if err != nil {
		return err
	}

	//store and set segment gpios as outputs
	for i, j := range segs {
		m.segs[i] = j
		m.OutGpio(j)
	}

	//store and set display gpios as outputs
	for i, j := range digs {
		m.digs[i] = j
		m.OutGpio(j)
	}

	//store and set dot gpio as ouput
	m.dot = dot
	m.OutGpio(dot)

	return err
}

//SetNumsDots sets the digits and dots to be displayed. Can be called when Run loop active
func (m *FourEightSegs) SetNumsDots(nums [4]uint32, dots [4]uint32) {
	for i := 0; i < 4; i++ {
		_ = atomic.SwapUint32(&m.nums[i], nums[i])
		_ = atomic.SwapUint32(&m.dots[i], dots[i])
	}
}

//SetDigsSegs sets the segment pins. Used with AllDigsOn on for hardware checks.
//Should not be call when Run loop active.
func (m *FourEightSegs) SetDigsSegs(i int) {
	disp := segdisp[i]
	for i = range disp {
		if disp[i] == 1 {
			m.SetGpio(m.segs[i])
		} else {
			m.ClrGpio(m.segs[i])
		}
	}
}

//AllDigsOn turns all the digits on (will display the same number).
//Used with SetDigsSegs to test hardware.
//Should not be call when Run loop active.
func (m *FourEightSegs) AllDigsOn() {
	for _, j := range m.digs {
		m.ClrGpio(j)
	}
}

//AllDigsOff makes sure the display is off
func (m *FourEightSegs) AllDigsOff() {
	//Turn off all digits
	for _, p := range m.digs {
		m.SetGpio(p)
	}
}

//Stop issues a stop to the active Run loop via an atomic int
func (m *FourEightSegs) Stop() {
	_ = atomic.AddUint32(&m.stop, uint32(1))
}

//Run starts the main display driving loop. the timout given (in ms) is
//for how long each digit remains on before moving to the next.
func (m *FourEightSegs) Run(ms int) chan struct{} {
	done := make(chan struct{}, 1)
	go m.run(ms, done)
	return done
}

func (m *FourEightSegs) run(ms int, done chan struct{}) {
	//Turn off all digits
	m.AllDigsOff()
	//j is the current digit
	//i is the current segment
	//dis is the int array that maps a number with the segments that need to be lit
	var j, i int
	var disp []int
	for {
		if atomic.LoadUint32(&m.stop) > 0 {
			done <- struct{}{}
			break
		}
		//for our 4 digits
		for j = 0; j < 4; j++ {
			//turn the last digit off
			//digit pins are acting as current sinks so high == off, low == on
			if j > 0 {
				m.SetGpio(m.digs[j-1])
			} else {
				m.SetGpio(m.digs[3])
			}

			//atomically load the number for the digit, then get the int array for it
			disp = segdisp[atomic.LoadUint32(&m.nums[j])]

			//turn on the segments that represent the number
			//segment pins act act current sources so high == on, low == off
			for i = range disp {
				if disp[i] == 1 {
					m.SetGpio(m.segs[i])
				} else {
					m.ClrGpio(m.segs[i])
				}
			}
			//turn on the dots on the digit
			//dot pins act act current sources so high == on, low == off
			if atomic.LoadUint32(&m.dots[j]) > 0 {
				m.SetGpio(m.dot)
			} else {
				m.ClrGpio(m.dot)
			}

			//all segments are ready, turn the digit on
			m.ClrGpio(m.digs[j])

			//keep the digit on for some time
			time.Sleep(time.Duration(ms) * time.Microsecond)
		}
	}
}
