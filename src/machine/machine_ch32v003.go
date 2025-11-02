//go:build ch32v003

package machine

import (
	"device/wch"
)

const deviceName = wch.Device

// CPUFrequency returns the current CPU clock in Hz.
// The CH32V003 starts up running from the 24MHz internal oscillator.
func CPUFrequency() uint32 {
	return 24_000_000
}

const (
	portA Pin = iota * 16
	_         // portB not present on this device
	portC
	portD
)

const (
	PA0 Pin = portA + 0
	PA1 Pin = portA + 1
	PA2 Pin = portA + 2
	PA3 Pin = portA + 3
	PA4 Pin = portA + 4
	PA5 Pin = portA + 5
	PA6 Pin = portA + 6
	PA7 Pin = portA + 7

	PC0 Pin = portC + 0
	PC1 Pin = portC + 1
	PC2 Pin = portC + 2
	PC3 Pin = portC + 3
	PC4 Pin = portC + 4
	PC5 Pin = portC + 5
	PC6 Pin = portC + 6
	PC7 Pin = portC + 7

	PD0 Pin = portD + 0
	PD1 Pin = portD + 1
	PD2 Pin = portD + 2
	PD3 Pin = portD + 3
	PD4 Pin = portD + 4
	PD5 Pin = portD + 5
	PD6 Pin = portD + 6
	PD7 Pin = portD + 7
)

const (
	PinInput       PinMode = 0
	PinOutput10MHz PinMode = 1
	PinOutput2MHz  PinMode = 2
	PinOutput50MHz PinMode = 3
	PinOutput      PinMode = PinOutput2MHz

	PinInputModeAnalog     PinMode = 0
	PinInputModeFloating   PinMode = 4
	PinInputModePullUpDown PinMode = 8
	PinInputModeReserved   PinMode = 12

	PinOutputModeGPPushPull   PinMode = 0
	PinOutputModeGPOpenDrain  PinMode = 4
	PinOutputModeAltPushPull  PinMode = 8
	PinOutputModeAltOpenDrain PinMode = 12

	PinInputPulldown PinMode = PinInputModePullUpDown
	PinInputPullup   PinMode = PinInputModePullUpDown | 0x10
)

func init() {
	wch.RCC.APB2PCENR.SetBits(wch.RCC_APB2PCENR_AFIOEN)
}

// Configure configures a GPIO pin according to the supplied PinConfig.
func (p Pin) Configure(config PinConfig) {
	if p == NoPin {
		return
	}

	port := p.port()
	pin := p.pinNumber()
	if pin >= 8 {
		panic("machine: unsupported pin number")
	}
	mask := uint32(1) << pin

	p.enableClock()

	mode := config.Mode
	rawMode := uint32(mode & 0x0f)
	pos := uint8(pin * 4)
	port.CFGLR.ReplaceBits(rawMode, 0x0f, pos)

	if (mode & 0x0f) == PinInputModePullUpDown {
		if mode&0x10 != 0 {
			port.OUTDR.SetBits(mask)
		} else {
			port.OUTDR.ClearBits(mask)
		}
	}
}

// Set drives the pin high or low.
func (p Pin) Set(high bool) {
	if p == NoPin {
		return
	}
	port := p.port()
	mask := p.mask()
	if high {
		port.BSHR.Set(mask)
	} else {
		port.BCR.Set(mask)
	}
}

// Get retrieves the current logic level on the pin.
func (p Pin) Get() bool {
	if p == NoPin {
		return false
	}
	port := p.port()
	mask := p.mask()
	return port.INDR.Get()&mask != 0
}

// PortMaskSet returns a register pointer and mask to set the pin high.
func (p Pin) PortMaskSet() (*uint32, uint32) {
	if p == NoPin {
		return nil, 0
	}
	port := p.port()
	return &port.BSHR.Reg, p.mask()
}

// PortMaskClear returns a register pointer and mask to drive the pin low.
func (p Pin) PortMaskClear() (*uint32, uint32) {
	if p == NoPin {
		return nil, 0
	}
	port := p.port()
	return &port.BCR.Reg, p.mask()
}

func (p Pin) enableClock() {
	switch p.portIndex() {
	case 0:
		wch.RCC.APB2PCENR.SetBits(wch.RCC_APB2PCENR_IOPAEN)
	case 2:
		wch.RCC.APB2PCENR.SetBits(wch.RCC_APB2PCENR_IOPCEN)
	case 3:
		wch.RCC.APB2PCENR.SetBits(wch.RCC_APB2PCENR_IOPDEN)
	default:
		panic("machine: unsupported port")
	}
}

func (p Pin) port() *wch.GPIO_Type {
	switch p.portIndex() {
	case 0:
		return wch.GPIOA
	case 2:
		return wch.GPIOC
	case 3:
		return wch.GPIOD
	default:
		panic("machine: unsupported port")
	}
}

func (p Pin) portIndex() uint8 {
	return uint8(p) / 16
}

func (p Pin) pinNumber() uint8 {
	return uint8(p) % 16
}

func (p Pin) mask() uint32 {
	pin := p.pinNumber()
	if pin >= 8 {
		panic("machine: unsupported pin number")
	}
	return 1 << pin
}

// UART represents the on-chip USART1 peripheral.
type UART struct {
	Bus *wch.USART_Type
}

// Hardware UART instances.
var (
	UART1  = &_UART1
	_UART1 = UART{
		Bus: wch.USART1,
	}
)

var DefaultUART = UART1

// Default UART pins on the CH32V003.
const (
	UART_TX_PIN = PC4
	UART_RX_PIN = PC5
)

// Configure sets up the UART peripheral using the provided configuration.
func (uart *UART) Configure(config UARTConfig) error {
	if config.BaudRate == 0 {
		config.BaudRate = 115200
	}
	if config.TX == 0 {
		config.TX = UART_TX_PIN
	}
	if config.RX == 0 {
		config.RX = UART_RX_PIN
	}
	if config.TX == NoPin {
		return ErrInvalidOutputPin
	}
	if config.RX == NoPin {
		return ErrInvalidInputPin
	}

	// Enable USART1 peripheral clock.
	wch.RCC.APB2PCENR.SetBits(wch.RCC_APB2PCENR_USART1EN)

	// Configure pins for alternate function operation.
	config.TX.Configure(PinConfig{Mode: PinOutput50MHz + PinOutputModeAltPushPull})
	config.RX.Configure(PinConfig{Mode: PinInputModeFloating})

	// Reset control registers before configuration.
	uart.Bus.CTLR1.Set(0)
	uart.Bus.CTLR2.Set(0)
	uart.Bus.CTLR3.Set(0)

	uart.SetBaudRate(config.BaudRate)

	uart.Bus.CTLR1.SetBits(wch.USART_CTLR1_TE | wch.USART_CTLR1_RE)
	uart.Bus.CTLR1.SetBits(wch.USART_CTLR1_UE)

	return nil
}

// SetBaudRate updates the USART baud rate divider.
func (uart *UART) SetBaudRate(baud uint32) {
	if baud == 0 {
		return
	}
	divider := CPUFrequency() / baud
	if divider == 0 {
		divider = 1
	}
	uart.Bus.BRR.Set(divider)
}

// WriteByte transmits a single byte and waits for completion.
func (uart *UART) WriteByte(c byte) error {
	for uart.Bus.STATR.Get()&wch.USART_STATR_TXE == 0 {
	}
	uart.Bus.DATAR.Set(uint32(c))
	for uart.Bus.STATR.Get()&wch.USART_STATR_TC == 0 {
	}
	return nil
}

// Write transmits a slice of bytes.
func (uart *UART) Write(data []byte) (int, error) {
	for i, b := range data {
		if err := uart.WriteByte(b); err != nil {
			return i, err
		}
	}
	return len(data), nil
}

// ReadByte reads a single byte if one is available.
func (uart *UART) ReadByte() (byte, error) {
	if uart.Buffered() == 0 {
		return 0, errNoByte
	}
	return byte(uart.Bus.DATAR.Get() & 0xff), nil
}

// Read reads up to len(data) bytes from the UART without blocking for new data.
func (uart *UART) Read(data []byte) (int, error) {
	count := 0
	for count < len(data) {
		b, err := uart.ReadByte()
		if err != nil {
			if count == 0 {
				return 0, err
			}
			break
		}
		data[count] = b
		count++
	}
	return count, nil
}

// Buffered reports whether unread data is waiting in the hardware FIFO.
func (uart *UART) Buffered() int {
	if uart.Bus.STATR.Get()&wch.USART_STATR_RXNE != 0 {
		return 1
	}
	return 0
}
