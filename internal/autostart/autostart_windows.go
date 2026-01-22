//go:build windows

package autostart

import (
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

const (
	registryKey = `Software\Microsoft\Windows\CurrentVersion\Run`
	appName     = "ACCJabraAgent"
)

// Enable adiciona o aplicativo ao autostart do Windows via Registry
func Enable(appPath string) error {
	// Se não foi fornecido path, usa o executável atual
	if appPath == "" {
		exe, err := os.Executable()
		if err != nil {
			return err
		}
		appPath = exe
	}

	// Resolve path absoluto
	absPath, err := filepath.Abs(appPath)
	if err != nil {
		return err
	}

	key, _, err := registry.CreateKey(
		registry.CURRENT_USER,
		registryKey,
		registry.SET_VALUE,
	)
	if err != nil {
		return err
	}
	defer key.Close()

	return key.SetStringValue(appName, absPath)
}

// Disable remove o aplicativo do autostart
func Disable() error {
	key, err := registry.OpenKey(
		registry.CURRENT_USER,
		registryKey,
		registry.SET_VALUE,
	)
	if err != nil {
		// Se a chave não existe, já está desabilitado
		if err == registry.ErrNotExist {
			return nil
		}
		return err
	}
	defer key.Close()

	err = key.DeleteValue(appName)
	if err == registry.ErrNotExist {
		return nil
	}
	return err
}

// IsEnabled verifica se o autostart está habilitado
func IsEnabled() bool {
	key, err := registry.OpenKey(
		registry.CURRENT_USER,
		registryKey,
		registry.QUERY_VALUE,
	)
	if err != nil {
		return false
	}
	defer key.Close()

	_, _, err = key.GetStringValue(appName)
	return err == nil
}

// GetPath retorna o caminho configurado no autostart
func GetPath() (string, error) {
	key, err := registry.OpenKey(
		registry.CURRENT_USER,
		registryKey,
		registry.QUERY_VALUE,
	)
	if err != nil {
		return "", err
	}
	defer key.Close()

	path, _, err := key.GetStringValue(appName)
	return path, err
}
