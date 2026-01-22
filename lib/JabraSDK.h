/**
 * JabraSDK.h - Headers mínimos para CGO wrapper
 *
 * Este arquivo define os tipos e funções necessários para integração
 * com libjabra.dll via CGO no Go.
 *
 * Baseado na documentação oficial do Jabra SDK.
 */

#ifndef JABRA_SDK_H
#define JABRA_SDK_H

#ifdef __cplusplus
extern "C" {
#endif

// ============================================================================
// Tipos Básicos
// ============================================================================

typedef unsigned short Jabra_DeviceID;
typedef int Jabra_ReturnCode;

// Códigos de retorno
#define JABRA_SUCCESS 0
#define JABRA_ERROR_INVALID_PARAMETER 1
#define JABRA_ERROR_NO_DEVICE 2
#define JABRA_ERROR_NOT_SUPPORTED 3
#define JABRA_ERROR_FAILED 4

// ============================================================================
// Estruturas
// ============================================================================

typedef struct {
    int levelInPercent;     // 0-100
    int charging;           // 0 = não, 1 = sim
    int batteryLow;         // 0 = não, 1 = sim
} Jabra_BatteryStatus;

typedef struct {
    Jabra_DeviceID deviceID;
    char* deviceName;
    char* serialNumber;
    unsigned short vendorID;
    unsigned short productID;
    int isDongle;
} Jabra_DeviceInfo;

// IDs de botões (TranslatedButtonInput)
typedef enum {
    CYCLIC,
    CYCLIC_END,
    DECLINE,
    DIAL_NEXT,
    DIAL_PREV,
    ENDCALL,
    FIRE_ALARM,
    FLASH,
    FLEXIBLE_BOOT_MUTE,
    GN_BUTTON_1,
    GN_BUTTON_2,
    GN_BUTTON_3,
    GN_BUTTON_4,
    GN_BUTTON_5,
    GN_BUTTON_6,
    HOOK_SWITCH,
    JABRA_BUTTON,
    KEY_0,
    KEY_1,
    KEY_2,
    KEY_3,
    KEY_4,
    KEY_5,
    KEY_6,
    KEY_7,
    KEY_8,
    KEY_9,
    KEY_CLEAR,
    KEY_POUND,
    KEY_STAR,
    LINE_BUSY,
    MUTE,
    OFFLINE,
    OFFHOOK,
    ONLINE,
    PSEUDO_OFFHOOK,
    REDIAL,
    REJECT_CALL,
    SPEED_DIAL,
    TRANSFER,
    VOICE_MAIL,
    VOLUME_DOWN,
    VOLUME_UP
} Jabra_ButtonID;

// ============================================================================
// Callbacks
// ============================================================================

/**
 * Callback chamado quando um dispositivo é conectado.
 * @param deviceID ID único do dispositivo
 */
typedef void (*Jabra_DeviceAttachedCallback)(Jabra_DeviceID deviceID);

/**
 * Callback chamado quando um dispositivo é desconectado.
 * @param deviceID ID único do dispositivo
 */
typedef void (*Jabra_DeviceDetachedCallback)(Jabra_DeviceID deviceID);

/**
 * Callback para eventos de botão traduzidos (alto nível).
 * @param deviceID ID do dispositivo
 * @param translatedInData ID do botão (Jabra_ButtonID)
 * @param value Estado do botão (1 = pressionado, 0 = liberado)
 */
typedef void (*Jabra_ButtonInDataTranslatedCallback)(
    Jabra_DeviceID deviceID,
    Jabra_ButtonID translatedInData,
    int value
);

/**
 * Callback para eventos de botão raw (baixo nível HID).
 * @param deviceID ID do dispositivo
 * @param usagePage Página HID
 * @param usage Uso HID
 * @param value Valor do evento
 */
typedef void (*Jabra_ButtonInDataRawHidCallback)(
    Jabra_DeviceID deviceID,
    unsigned short usagePage,
    unsigned short usage,
    int value
);

/**
 * Callback para atualização de status da bateria.
 * @param deviceID ID do dispositivo
 * @param levelInPercent Nível da bateria (0-100)
 * @param charging Se está carregando
 * @param batteryLow Se bateria está baixa
 */
typedef void (*Jabra_BatteryStatusUpdateCallback)(
    Jabra_DeviceID deviceID,
    int levelInPercent,
    int charging,
    int batteryLow
);

// ============================================================================
// Funções de Inicialização
// ============================================================================

/**
 * Inicializa o SDK da Jabra.
 * Deve ser chamado antes de qualquer outra função.
 *
 * @param appID ID da aplicação (GUID fornecido pela Jabra)
 * @return JABRA_SUCCESS em caso de sucesso
 */
Jabra_ReturnCode Jabra_Initialize(const char* appID);

/**
 * Finaliza o SDK e libera recursos.
 */
void Jabra_Uninitialize(void);

/**
 * Verifica se o SDK está inicializado.
 * @return 1 se inicializado, 0 caso contrário
 */
int Jabra_IsInitialized(void);

// ============================================================================
// Funções de Registro de Callbacks
// ============================================================================

/**
 * Registra callback para dispositivo conectado.
 */
void Jabra_RegisterDeviceAttachedCallback(Jabra_DeviceAttachedCallback callback);

/**
 * Registra callback para dispositivo desconectado.
 */
void Jabra_RegisterDeviceDetachedCallback(Jabra_DeviceDetachedCallback callback);

/**
 * Registra callback para eventos de botão traduzidos.
 */
void Jabra_RegisterButtonInDataTranslatedCallback(Jabra_ButtonInDataTranslatedCallback callback);

/**
 * Registra callback para eventos de botão raw HID.
 */
void Jabra_RegisterButtonInDataRawHidCallback(Jabra_ButtonInDataRawHidCallback callback);

/**
 * Registra callback para atualização de bateria.
 */
void Jabra_RegisterBatteryStatusUpdateCallback(Jabra_BatteryStatusUpdateCallback callback);

// ============================================================================
// Funções de Dispositivo
// ============================================================================

/**
 * Obtém lista de dispositivos conectados.
 * @param count Ponteiro para receber quantidade de dispositivos
 * @return Array de Jabra_DeviceInfo (deve ser liberado com Jabra_FreeDeviceList)
 */
Jabra_DeviceInfo* Jabra_GetAttachedDevices(int* count);

/**
 * Libera lista de dispositivos.
 */
void Jabra_FreeDeviceList(Jabra_DeviceInfo* devices);

/**
 * Obtém nome do dispositivo.
 * @param deviceID ID do dispositivo
 * @return Nome do dispositivo (não liberar, gerenciado pelo SDK)
 */
const char* Jabra_GetDeviceName(Jabra_DeviceID deviceID);

/**
 * Obtém número serial do dispositivo.
 * @param deviceID ID do dispositivo
 * @return Serial (deve ser liberado com Jabra_FreeString)
 */
char* Jabra_GetSerialNumber(Jabra_DeviceID deviceID);

/**
 * Libera string alocada pelo SDK.
 */
void Jabra_FreeString(char* str);

/**
 * Verifica se dispositivo é um dongle.
 * @param deviceID ID do dispositivo
 * @return 1 se dongle, 0 se headset
 */
int Jabra_IsDongle(Jabra_DeviceID deviceID);

// ============================================================================
// Funções de Bateria
// ============================================================================

/**
 * Obtém status da bateria.
 * @param deviceID ID do dispositivo
 * @param status Ponteiro para receber status
 * @return JABRA_SUCCESS em caso de sucesso
 */
Jabra_ReturnCode Jabra_GetBatteryStatus(
    Jabra_DeviceID deviceID,
    Jabra_BatteryStatus* status
);

// ============================================================================
// Funções de Controle
// ============================================================================

/**
 * Define estado do mute.
 * @param deviceID ID do dispositivo
 * @param mute 1 para mute, 0 para unmute
 * @return JABRA_SUCCESS em caso de sucesso
 */
Jabra_ReturnCode Jabra_SetMute(Jabra_DeviceID deviceID, int mute);

/**
 * Obtém estado do mute.
 * @param deviceID ID do dispositivo
 * @param mute Ponteiro para receber estado
 * @return JABRA_SUCCESS em caso de sucesso
 */
Jabra_ReturnCode Jabra_GetMute(Jabra_DeviceID deviceID, int* mute);

/**
 * Define estado do ringer (toque).
 * @param deviceID ID do dispositivo
 * @param ringer 1 para ligar, 0 para desligar
 * @return JABRA_SUCCESS em caso de sucesso
 */
Jabra_ReturnCode Jabra_SetRinger(Jabra_DeviceID deviceID, int ringer);

/**
 * Define estado do hook (gancho).
 * @param deviceID ID do dispositivo
 * @param offHook 1 para off-hook (atendendo), 0 para on-hook
 * @return JABRA_SUCCESS em caso de sucesso
 */
Jabra_ReturnCode Jabra_SetHookState(Jabra_DeviceID deviceID, int offHook);

/**
 * Define estado do busylight (LED de ocupado).
 * @param deviceID ID do dispositivo
 * @param on 1 para ligar, 0 para desligar
 * @return JABRA_SUCCESS em caso de sucesso
 */
Jabra_ReturnCode Jabra_SetBusylightState(Jabra_DeviceID deviceID, int on);

/**
 * Define estado do hold (chamada em espera).
 * @param deviceID ID do dispositivo
 * @param hold 1 para hold, 0 para resume
 * @return JABRA_SUCCESS em caso de sucesso
 */
Jabra_ReturnCode Jabra_SetHold(Jabra_DeviceID deviceID, int hold);

// ============================================================================
// Funções de Áudio
// ============================================================================

/**
 * Define volume do dispositivo.
 * @param deviceID ID do dispositivo
 * @param volume Nível de volume (0-100)
 * @return JABRA_SUCCESS em caso de sucesso
 */
Jabra_ReturnCode Jabra_SetVolume(Jabra_DeviceID deviceID, int volume);

/**
 * Obtém volume do dispositivo.
 * @param deviceID ID do dispositivo
 * @param volume Ponteiro para receber volume
 * @return JABRA_SUCCESS em caso de sucesso
 */
Jabra_ReturnCode Jabra_GetVolume(Jabra_DeviceID deviceID, int* volume);

#ifdef __cplusplus
}
#endif

#endif // JABRA_SDK_H
