//go:build windows

package jabra

/*
#cgo CFLAGS: -I../../lib
#cgo LDFLAGS: -L../../lib -ljabra

#include "JabraSDK.h"
#include <stdlib.h>

// Forward declarations para callbacks Go
extern void goOnDeviceAttached(unsigned short deviceID);
extern void goOnDeviceDetached(unsigned short deviceID);
extern void goOnButtonEvent(unsigned short deviceID, int buttonID, int value);
extern void goOnBatteryUpdate(unsigned short deviceID, int level, int charging, int low);

// Wrappers C para registrar callbacks
static void registerCallbacks() {
    Jabra_RegisterDeviceAttachedCallback(goOnDeviceAttached);
    Jabra_RegisterDeviceDetachedCallback(goOnDeviceDetached);
    Jabra_RegisterButtonInDataTranslatedCallback(
        (Jabra_ButtonInDataTranslatedCallback)goOnButtonEvent
    );
    Jabra_RegisterBatteryStatusUpdateCallback(
        (Jabra_BatteryStatusUpdateCallback)goOnBatteryUpdate
    );
}
*/
import "C"

import (
	"errors"
	"fmt"
	"sync"
	"time"
	"unsafe"
)

// Variável global para o singleton do driver (necessário para callbacks CGO)
var globalSDKDriver *SDKDriver
var globalSDKDriverMu sync.RWMutex

// SDKDriver implementa Driver usando o Jabra SDK nativo (Windows)
type SDKDriver struct {
	mu      sync.RWMutex
	config  DriverConfig
	running bool

	// Dispositivos conectados
	devices map[uint16]*DeviceInfo

	// Callbacks registrados
	onDeviceConnected    func(event DeviceEvent)
	onDeviceDisconnected func(event DeviceEvent)
	onButtonEvent        func(event ButtonEvent)
	onBatteryUpdate      func(deviceID uint16, status BatteryStatus)
}

// NewSDKDriver cria uma nova instância do driver SDK para Windows
func NewSDKDriver(config DriverConfig) (*SDKDriver, error) {
	driver := &SDKDriver{
		config:  config,
		devices: make(map[uint16]*DeviceInfo),
	}

	return driver, nil
}

// Start inicializa o Jabra SDK
func (d *SDKDriver) Start() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.running {
		return errors.New("driver already running")
	}

	// Inicializa o SDK
	appID := C.CString(d.config.AppID)
	defer C.free(unsafe.Pointer(appID))

	result := C.Jabra_Initialize(appID)
	if result != C.JABRA_SUCCESS {
		return fmt.Errorf("Jabra_Initialize failed with code: %d", result)
	}

	// Registra este driver como global para callbacks
	globalSDKDriverMu.Lock()
	globalSDKDriver = d
	globalSDKDriverMu.Unlock()

	// Registra callbacks
	C.registerCallbacks()

	// Enumera dispositivos já conectados
	d.enumerateDevices()

	d.running = true
	return nil
}

// Stop finaliza o SDK
func (d *SDKDriver) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.running {
		return nil
	}

	C.Jabra_Uninitialize()

	globalSDKDriverMu.Lock()
	globalSDKDriver = nil
	globalSDKDriverMu.Unlock()

	d.running = false
	d.devices = make(map[uint16]*DeviceInfo)

	return nil
}

// IsRunning retorna se o driver está ativo
func (d *SDKDriver) IsRunning() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.running
}

// enumerateDevices obtém lista de dispositivos já conectados
func (d *SDKDriver) enumerateDevices() {
	var count C.int
	devices := C.Jabra_GetAttachedDevices(&count)
	if devices == nil || count == 0 {
		return
	}
	defer C.Jabra_FreeDeviceList(devices)

	// Itera sobre os dispositivos
	deviceSlice := unsafe.Slice(devices, int(count))
	for _, dev := range deviceSlice {
		info := &DeviceInfo{
			ID:          uint16(dev.deviceID),
			Name:        C.GoString(dev.deviceName),
			VendorID:    uint16(dev.vendorID),
			ProductID:   uint16(dev.productID),
			IsDongle:    dev.isDongle != 0,
			Connected:   true,
			ConnectedAt: time.Now(),
		}

		// Obtém serial separadamente
		serial := C.Jabra_GetSerialNumber(dev.deviceID)
		if serial != nil {
			info.SerialNumber = C.GoString(serial)
			C.Jabra_FreeString(serial)
		}

		d.devices[info.ID] = info
	}
}

// GetDevices retorna lista de dispositivos conectados
func (d *SDKDriver) GetDevices() []DeviceInfo {
	d.mu.RLock()
	defer d.mu.RUnlock()

	devices := make([]DeviceInfo, 0, len(d.devices))
	for _, dev := range d.devices {
		devices = append(devices, *dev)
	}
	return devices
}

// GetDevice retorna informações de um dispositivo específico
func (d *SDKDriver) GetDevice(deviceID uint16) (*DeviceInfo, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	dev, ok := d.devices[deviceID]
	if !ok {
		return nil, fmt.Errorf("device %d not found", deviceID)
	}
	return dev, nil
}

// GetBatteryStatus obtém status da bateria
func (d *SDKDriver) GetBatteryStatus(deviceID uint16) (*BatteryStatus, error) {
	d.mu.RLock()
	if !d.running {
		d.mu.RUnlock()
		return nil, errors.New("driver not running")
	}
	d.mu.RUnlock()

	var status C.Jabra_BatteryStatus
	result := C.Jabra_GetBatteryStatus(C.Jabra_DeviceID(deviceID), &status)
	if result != C.JABRA_SUCCESS {
		return nil, fmt.Errorf("GetBatteryStatus failed: %d", result)
	}

	return &BatteryStatus{
		Level:      int(status.levelInPercent),
		IsCharging: status.charging != 0,
		IsLow:      status.batteryLow != 0,
	}, nil
}

// SetMute define estado do mute
func (d *SDKDriver) SetMute(deviceID uint16, mute bool) error {
	muteVal := C.int(0)
	if mute {
		muteVal = 1
	}

	result := C.Jabra_SetMute(C.Jabra_DeviceID(deviceID), muteVal)
	if result != C.JABRA_SUCCESS {
		return fmt.Errorf("SetMute failed: %d", result)
	}
	return nil
}

// GetMute obtém estado do mute
func (d *SDKDriver) GetMute(deviceID uint16) (bool, error) {
	var mute C.int
	result := C.Jabra_GetMute(C.Jabra_DeviceID(deviceID), &mute)
	if result != C.JABRA_SUCCESS {
		return false, fmt.Errorf("GetMute failed: %d", result)
	}
	return mute != 0, nil
}

// SetRinger define estado do ringer
func (d *SDKDriver) SetRinger(deviceID uint16, ring bool) error {
	ringVal := C.int(0)
	if ring {
		ringVal = 1
	}

	result := C.Jabra_SetRinger(C.Jabra_DeviceID(deviceID), ringVal)
	if result != C.JABRA_SUCCESS {
		return fmt.Errorf("SetRinger failed: %d", result)
	}
	return nil
}

// SetHookState define estado do hook
func (d *SDKDriver) SetHookState(deviceID uint16, offHook bool) error {
	hookVal := C.int(0)
	if offHook {
		hookVal = 1
	}

	result := C.Jabra_SetHookState(C.Jabra_DeviceID(deviceID), hookVal)
	if result != C.JABRA_SUCCESS {
		return fmt.Errorf("SetHookState failed: %d", result)
	}
	return nil
}

// SetBusylight define estado do LED
func (d *SDKDriver) SetBusylight(deviceID uint16, on bool) error {
	onVal := C.int(0)
	if on {
		onVal = 1
	}

	result := C.Jabra_SetBusylightState(C.Jabra_DeviceID(deviceID), onVal)
	if result != C.JABRA_SUCCESS {
		return fmt.Errorf("SetBusylightState failed: %d", result)
	}
	return nil
}

// SetHold define estado de hold
func (d *SDKDriver) SetHold(deviceID uint16, hold bool) error {
	holdVal := C.int(0)
	if hold {
		holdVal = 1
	}

	result := C.Jabra_SetHold(C.Jabra_DeviceID(deviceID), holdVal)
	if result != C.JABRA_SUCCESS {
		return fmt.Errorf("SetHold failed: %d", result)
	}
	return nil
}

// SetVolume define volume do dispositivo
func (d *SDKDriver) SetVolume(deviceID uint16, volume int) error {
	if volume < 0 {
		volume = 0
	}
	if volume > 100 {
		volume = 100
	}

	result := C.Jabra_SetVolume(C.Jabra_DeviceID(deviceID), C.int(volume))
	if result != C.JABRA_SUCCESS {
		return fmt.Errorf("SetVolume failed: %d", result)
	}
	return nil
}

// GetVolume obtém volume do dispositivo
func (d *SDKDriver) GetVolume(deviceID uint16) (int, error) {
	var volume C.int
	result := C.Jabra_GetVolume(C.Jabra_DeviceID(deviceID), &volume)
	if result != C.JABRA_SUCCESS {
		return 0, fmt.Errorf("GetVolume failed: %d", result)
	}
	return int(volume), nil
}

// OnDeviceConnected registra callback
func (d *SDKDriver) OnDeviceConnected(handler func(event DeviceEvent)) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.onDeviceConnected = handler
}

// OnDeviceDisconnected registra callback
func (d *SDKDriver) OnDeviceDisconnected(handler func(event DeviceEvent)) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.onDeviceDisconnected = handler
}

// OnButtonEvent registra callback
func (d *SDKDriver) OnButtonEvent(handler func(event ButtonEvent)) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.onButtonEvent = handler
}

// OnBatteryUpdate registra callback
func (d *SDKDriver) OnBatteryUpdate(handler func(deviceID uint16, status BatteryStatus)) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.onBatteryUpdate = handler
}

// =============================================================================
// Callbacks exportados para C
// =============================================================================

//export goOnDeviceAttached
func goOnDeviceAttached(deviceID C.ushort) {
	globalSDKDriverMu.RLock()
	driver := globalSDKDriver
	globalSDKDriverMu.RUnlock()

	if driver == nil {
		return
	}

	driver.mu.Lock()

	// Obtém informações do dispositivo
	name := C.Jabra_GetDeviceName(C.Jabra_DeviceID(deviceID))
	serial := C.Jabra_GetSerialNumber(C.Jabra_DeviceID(deviceID))
	isDongle := C.Jabra_IsDongle(C.Jabra_DeviceID(deviceID))

	info := &DeviceInfo{
		ID:          uint16(deviceID),
		Name:        C.GoString(name),
		IsDongle:    isDongle != 0,
		Connected:   true,
		ConnectedAt: time.Now(),
	}

	if serial != nil {
		info.SerialNumber = C.GoString(serial)
		C.Jabra_FreeString(serial)
	}

	driver.devices[info.ID] = info
	handler := driver.onDeviceConnected
	driver.mu.Unlock()

	if handler != nil {
		handler(DeviceEvent{
			DeviceID:  info.ID,
			Connected: true,
			Device:    info,
		})
	}
}

//export goOnDeviceDetached
func goOnDeviceDetached(deviceID C.ushort) {
	globalSDKDriverMu.RLock()
	driver := globalSDKDriver
	globalSDKDriverMu.RUnlock()

	if driver == nil {
		return
	}

	driver.mu.Lock()
	info, ok := driver.devices[uint16(deviceID)]
	if ok {
		info.Connected = false
		delete(driver.devices, uint16(deviceID))
	}
	handler := driver.onDeviceDisconnected
	driver.mu.Unlock()

	if handler != nil && info != nil {
		handler(DeviceEvent{
			DeviceID:  uint16(deviceID),
			Connected: false,
			Device:    info,
		})
	}
}

//export goOnButtonEvent
func goOnButtonEvent(deviceID C.ushort, buttonID C.int, value C.int) {
	globalSDKDriverMu.RLock()
	driver := globalSDKDriver
	globalSDKDriverMu.RUnlock()

	if driver == nil {
		return
	}

	driver.mu.RLock()
	handler := driver.onButtonEvent
	driver.mu.RUnlock()

	if handler != nil {
		handler(ButtonEvent{
			DeviceID: uint16(deviceID),
			ButtonID: ButtonID(buttonID),
			Pressed:  value != 0,
		})
	}
}

//export goOnBatteryUpdate
func goOnBatteryUpdate(deviceID C.ushort, level C.int, charging C.int, low C.int) {
	globalSDKDriverMu.RLock()
	driver := globalSDKDriver
	globalSDKDriverMu.RUnlock()

	if driver == nil {
		return
	}

	driver.mu.RLock()
	handler := driver.onBatteryUpdate
	driver.mu.RUnlock()

	if handler != nil {
		handler(uint16(deviceID), BatteryStatus{
			Level:      int(level),
			IsCharging: charging != 0,
			IsLow:      low != 0,
		})
	}
}
