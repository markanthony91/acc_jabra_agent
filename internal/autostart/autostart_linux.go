//go:build linux

package autostart

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	desktopFileName = "acc-jabra-agent.desktop"
	appComment      = "Jabra Telemetry Agent for ACC"
)

// Enable cria arquivo .desktop no diretório autostart do Linux
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

	autostartDir, err := getAutostartDir()
	if err != nil {
		return err
	}

	// Cria diretório se não existir
	if err := os.MkdirAll(autostartDir, 0755); err != nil {
		return err
	}

	content := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=ACC Jabra Agent
Comment=%s
Exec=%s
Terminal=false
Icon=audio-headset
Categories=System;Utility;
StartupNotify=false
X-GNOME-Autostart-enabled=true
`, appComment, absPath)

	desktopPath := filepath.Join(autostartDir, desktopFileName)
	return os.WriteFile(desktopPath, []byte(content), 0644)
}

// Disable remove o arquivo .desktop do autostart
func Disable() error {
	autostartDir, err := getAutostartDir()
	if err != nil {
		return err
	}

	desktopPath := filepath.Join(autostartDir, desktopFileName)
	err = os.Remove(desktopPath)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// IsEnabled verifica se o autostart está habilitado
func IsEnabled() bool {
	autostartDir, err := getAutostartDir()
	if err != nil {
		return false
	}

	desktopPath := filepath.Join(autostartDir, desktopFileName)
	_, err = os.Stat(desktopPath)
	return err == nil
}

// GetPath retorna o caminho do executável configurado no autostart
func GetPath() (string, error) {
	autostartDir, err := getAutostartDir()
	if err != nil {
		return "", err
	}

	desktopPath := filepath.Join(autostartDir, desktopFileName)
	content, err := os.ReadFile(desktopPath)
	if err != nil {
		return "", err
	}

	// Parse simples do campo Exec=
	lines := string(content)
	for _, line := range filepath.SplitList(lines) {
		if len(line) > 5 && line[:5] == "Exec=" {
			return line[5:], nil
		}
	}

	return "", fmt.Errorf("Exec not found in desktop file")
}

// getAutostartDir retorna o diretório de autostart do usuário
func getAutostartDir() (string, error) {
	// Primeiro tenta XDG_CONFIG_HOME
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome != "" {
		return filepath.Join(configHome, "autostart"), nil
	}

	// Fallback para ~/.config/autostart
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".config", "autostart"), nil
}
