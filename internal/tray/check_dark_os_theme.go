package tray

import (
	"os/exec"
	"runtime"
	"strings"
)

func isDarkTheme() (bool, error) {
	switch runtime.GOOS {
	case "darwin":
		return checkMacOSTheme()
	case "windows":
		return checkWindowsTheme()
	case "linux":
		return checkLinuxTheme()
	default:
		return false, nil
	}
}

func checkMacOSTheme() (bool, error) {
	cmd := exec.Command("defaults", "read", "-g", "AppleInterfaceStyle")
	output, err := cmd.Output()
	if err == nil {
		return strings.Contains(strings.ToLower(string(output)), "dark"), nil
	}
	return false, nil
}

func checkWindowsTheme() (bool, error) {
	cmd := exec.Command("reg", "QUERY", "HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Themes\\Personalize", "/v", "AppsUseLightTheme")
	output, err := cmd.Output()

	if err != nil {
		return false, err
	}
	return strings.Contains(string(output), "0x0"), nil
}

func checkLinuxTheme() (bool, error) {
	cmd := exec.Command("gsettings", "get", "org.gnome.desktop.interface", "color-scheme")
	output, err := cmd.Output()

	if err != nil {
		return false, err
	}
	return strings.Contains(string(output), "dark"), nil
}
