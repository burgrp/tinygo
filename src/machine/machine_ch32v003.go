//go:build ch32v003

package machine

import (
	"runtime/volatile"
	"unsafe"
)

const deviceName = "CH32V003"

const (
	portA Pin = iota * 16
	portB
	portC
	portD
)

// Pin assignments.
const (
	PA0  = portA + 0
	PA1  = portA + 1
	PA2  = portA + 2
	PA3  = portA + 3
	PA4  = portA + 4
	PA5  = portA + 5
	PA6  = portA + 6
	PA7  = portA + 7
	PA8  = portA + 8
	PA9  = portA + 9
	PA10 = portA + 10
	PA11 = portA + 11
	PA12 = portA + 12
	PA13 = portA + 13
	PA14 = portA + 14
	PA15 = portA + 15

	PB0  = portB + 0
	PB1  = portB + 1
	PB2  = portB + 2
	PB3  = portB + 3
	PB4  = portB + 4
	PB5  = portB + 5
	PB6  = portB + 6
	PB7  = portB + 7
	PB8  = portB + 8
	PB9  = portB + 9
	PB10 = portB + 10
	PB11 = portB + 11
	PB12 = portB + 12
	PB13 = portB + 13
	PB14 = portB + 14
	PB15 = portB + 15

	PC0  = portC + 0
	PC1  = portC + 1
	PC2  = portC + 2
	PC3  = portC + 3
	PC4  = portC + 4
	PC5  = portC + 5
	PC6  = portC + 6
	PC7  = portC + 7
	PC8  = portC + 8
	PC9  = portC + 9
	PC10 = portC + 10
	PC11 = portC + 11
	PC12 = portC + 12
	PC13 = portC + 13
	PC14 = portC + 14
	PC15 = portC + 15

	PD0 = portD + 0
	PD1 = portD + 1
	PD2 = portD + 2
	PD3 = portD + 3
	PD4 = portD + 4
	PD5 = portD + 5
	PD6 = portD + 6
	PD7 = portD + 7
)

// GPIO pin configuration modes.
const (
	PinInput       PinMode = 0 // Input mode
	PinOutput10MHz PinMode = 1 // Output mode, max speed 10MHz
	PinOutput2MHz  PinMode = 2 // Output mode, max speed 2MHz
	PinOutput50MHz PinMode = 3 // Output mode, max speed 50MHz
	PinOutput      PinMode = PinOutput2MHz

	PinInputModeAnalog     PinMode = 0  // Input analog mode
	PinInputModeFloating   PinMode = 4  // Input floating mode
	PinInputModePullUpDown PinMode = 8  // Input pull up/down mode
	PinInputModeReserved   PinMode = 12 // Input mode (reserved)

	PinOutputModeGPPushPull   PinMode = 0  // Output mode general purpose push/pull
	PinOutputModeGPOpenDrain  PinMode = 4  // Output mode general purpose open drain
	PinOutputModeAltPushPull  PinMode = 8  // Output mode alternate function push/pull
	PinOutputModeAltOpenDrain PinMode = 12 // Output mode alternate function open drain

	PinInputPulldown PinMode = PinInputModePullUpDown
	PinInputPullup   PinMode = PinInputModePullUpDown | 0x10
)

const (
	gpioABase = 0x40010800
	gpioBBase = 0x40010C00
	gpioCBase = 0x40011000
	gpioDBase = 0x40011400

	rccBase = 0x40021000
	usart1  = 0x40013800
)

type gpioRegister struct {
	CFGLR volatile.Register32
	CFGHR volatile.Register32
	INFR  volatile.Register32
	OUTDR volatile.Register32
	BSHR  volatile.Register32
	BCR   volatile.Register32
	LCKR  volatile.Register32
}

var (
	gpioAPtr = (*gpioRegister)(unsafe.Pointer(uintptr(gpioABase)))
	gpioBPtr = (*gpioRegister)(unsafe.Pointer(uintptr(gpioBBase)))
	gpioCPtr = (*gpioRegister)(unsafe.Pointer(uintptr(gpioCBase)))
	gpioDPtr = (*gpioRegister)(unsafe.Pointer(uintptr(gpioDBase)))
)

type rccRegister struct {
	CTLR      volatile.Register32
	CFGR0     volatile.Register32
	INTR      volatile.Register32
	APB2PRSTR volatile.Register32
	APB1PRSTR volatile.Register32
	AHBPCENR  volatile.Register32
	APB2PCENR volatile.Register32
	APB1PCENR volatile.Register32
	BDCTLR    volatile.Register32
	RSTSCKR   volatile.Register32
	AHBPRSTR  volatile.Register32
	CFGR2     volatile.Register32
	CFGR3     volatile.Register32
}

var rcc = (*rccRegister)(unsafe.Pointer(uintptr(rccBase)))

const (
	rccAPB2IOPAEN   = 1 << 2
	rccAPB2IOPBEN   = 1 << 3
	rccAPB2IOPCEN   = 1 << 4
	rccAPB2IOPDEN   = 1 << 5
	rccAPB2AFIOEN   = 1 << 0
	rccAPB2USART1EN = 1 << 14
)

// CPUFrequency returns the current CPU frequency in hertz.
func CPUFrequency() uint32 {
	return 48000000
}

// Configure configures the GPIO pin with the provided settings.
func (p Pin) Configure(config PinConfig) {
	p.enableClock()
	port := p.getPort()
	pin := uint8(p) % 16
	pos := (pin % 8) * 4
	if pin < 8 {
		port.CFGLR.ReplaceBits(uint32(config.Mode)&0xf, 0xf, pos)
	} else {
		port.CFGHR.ReplaceBits(uint32(config.Mode)&0xf, 0xf, pos)
	}

	if (config.Mode & 0xf) == PinInputModePullUpDown {
		var pullup uint32
		if config.Mode&0x10 != 0 {
			pullup = 1
		}
		port.OUTDR.ReplaceBits(pullup, 0x1, pin)
	}
}

// Set sets the GPIO output level.
func (p Pin) Set(high bool) {
	port := p.getPort()
	pin := uint8(p) % 16
	if high {
		port.BSHR.Set(1 << pin)
	} else {
		port.BCR.Set(1 << pin)
	}
}

// Get returns the current value of a GPIO pin when configured as either input or output.
func (p Pin) Get() bool {
	port := p.getPort()
	pin := uint8(p) % 16
	return (port.INFR.Get() & (1 << pin)) != 0
}

// PortMaskSet returns the register and mask to enable a given GPIO pin.
func (p Pin) PortMaskSet() (*uint32, uint32) {
	port := p.getPort()
	pin := uint8(p) % 16
	return &port.BSHR.Reg, 1 << pin
}

// PortMaskClear returns the register and mask to disable a given GPIO pin.
func (p Pin) PortMaskClear() (*uint32, uint32) {
	port := p.getPort()
	pin := uint8(p) % 16
	return &port.BCR.Reg, 1 << pin
}

func (p Pin) getPort() *gpioRegister {
	switch p / 16 {
	case 0:
		return gpioAPtr
	case 1:
		return gpioBPtr
	case 2:
		return gpioCPtr
	case 3:
		return gpioDPtr
	default:
		return gpioAPtr
	}
}

func (p Pin) enableClock() {
	switch p / 16 {
	case 0:
		rcc.APB2PCENR.SetBits(rccAPB2IOPAEN)
	case 1:
		rcc.APB2PCENR.SetBits(rccAPB2IOPBEN)
	case 2:
		rcc.APB2PCENR.SetBits(rccAPB2IOPCEN)
	case 3:
		rcc.APB2PCENR.SetBits(rccAPB2IOPDEN)
	}
}

// UART represents a universal asynchronous receiver/transmitter peripheral.
type UART struct {
	Bus *usartRegister
}

var (
	UART0  = &_UART0
	_UART0 = UART{Bus: (*usartRegister)(unsafe.Pointer(uintptr(usart1)))}
)

// DefaultUART is the default UART peripheral.
var DefaultUART = UART0

// UART configuration constants.
const (
	usartSR_TXE  = 1 << 7
	usartSR_TC   = 1 << 6
	usartSR_RXNE = 1 << 5

	usartCR1_RE = 1 << 2
	usartCR1_TE = 1 << 3
	usartCR1_UE = 1 << 13
)

// usartRegister represents the memory layout of the USART peripheral.
type usartRegister struct {
	STATR volatile.Register32
	DATAR volatile.Register32
	BRR   volatile.Register32
	CTLR1 volatile.Register32
	CTLR2 volatile.Register32
	CTLR3 volatile.Register32
	GPR   volatile.Register32
}

// Configure configures the UART peripheral.
func (uart *UART) Configure(config UARTConfig) error {
	if config.BaudRate == 0 {
		config.BaudRate = 115200
	}
	if config.TX == 0 {
		config.TX = PA9
	}
	if config.RX == 0 {
		config.RX = PA10
	}

	rcc.APB2PCENR.SetBits(rccAPB2AFIOEN | rccAPB2USART1EN)

	config.TX.Configure(PinConfig{Mode: PinOutputModeAltPushPull | PinOutput50MHz})
	config.RX.Configure(PinConfig{Mode: PinInputModeFloating})

	divider := (CPUFrequency() + config.BaudRate/2) / config.BaudRate
	uart.Bus.BRR.Set(divider)
	uart.Bus.CTLR2.Set(0)
	uart.Bus.CTLR3.Set(0)
	uart.Bus.CTLR1.Set(usartCR1_RE | usartCR1_TE | usartCR1_UE)

	return nil
}

// Write sends the provided data buffer over UART.
func (uart *UART) Write(data []byte) (n int, err error) {
	for _, b := range data {
		if err := uart.WriteByte(b); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

// WriteByte writes a single byte over UART.
func (uart *UART) WriteByte(c byte) error {
	for uart.Bus.STATR.Get()&usartSR_TXE == 0 {
	}
	uart.Bus.DATAR.Set(uint32(c))
	return nil
}

// ReadByte reads a single byte from UART.
func (uart *UART) ReadByte() (byte, error) {
	for uart.Bus.STATR.Get()&usartSR_RXNE == 0 {
	}
	value := byte(uart.Bus.DATAR.Get() & 0xff)
	return value, nil
}

// Buffered reports how many bytes are pending in the UART receive FIFO.
func (uart *UART) Buffered() int {
	if uart.Bus.STATR.Get()&usartSR_RXNE != 0 {
		return 1
	}
	return 0
}

// flush waits until transmission is complete.
func (uart *UART) flush() {
	for uart.Bus.STATR.Get()&usartSR_TC == 0 {
	}
}

// DTR is not implemented on this UART.
func (uart *UART) DTR() bool {
	return false
}

// RTS is not implemented on this UART.
func (uart *UART) RTS() bool {
	return false
}
