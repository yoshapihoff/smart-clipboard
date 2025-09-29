package tray

import (
	"os/exec"
	"runtime"
	"strings"
)

// isDarkTrayBackground checks if the tray/taskbar background is dark
func isDarkTrayBackground() (bool, error) {
	switch runtime.GOOS {
	case "darwin":
		return checkMacOSTrayBackground()
	case "windows":
		return checkWindowsTrayBackground()
	case "linux":
		return checkLinuxTrayBackground()
	default:
		return false, nil
	}
}

// isDarkTheme checks the general OS theme (fallback)
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

func checkMacOSTrayBackground() (bool, error) {
	// Check if the menu bar is dark (macOS specific)
	cmd := exec.Command("defaults", "read", "-g", "AppleInterfaceStyle")
	output, err := cmd.Output()
	if err == nil {
		return strings.Contains(strings.ToLower(string(output)), "dark"), nil
	}
	
	// Fallback: check if we're in dark mode by checking the system appearance
	cmd = exec.Command("defaults", "read", "com.apple.controlcenter", "NSStatusItem Visible Item")
	output, err = cmd.Output()
	if err == nil {
		// This is a heuristic - in practice, we might need a more sophisticated approach
		return strings.Contains(string(output), "Dark"), nil
	}
	
	return false, nil
}

func checkMacOSTheme() (bool, error) {
	cmd := exec.Command("defaults", "read", "-g", "AppleInterfaceStyle")
	output, err := cmd.Output()
	if err == nil {
		return strings.Contains(strings.ToLower(string(output)), "dark"), nil
	}
	return false, nil
}

func checkWindowsTrayBackground() (bool, error) {
	// Check the taskbar color specifically
	cmd := exec.Command("reg", "QUERY", "HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Themes\\Personalize", "/v", "SystemUsesLightTheme")
	output, err := cmd.Output()

	if err != nil {
		// Fallback to apps theme
		return checkWindowsTheme()
	}
	
	// If SystemUsesLightTheme is 0x0, it means dark theme
	return strings.Contains(string(output), "0x0"), nil
}

func checkWindowsTheme() (bool, error) {
	cmd := exec.Command("reg", "QUERY", "HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Themes\\Personalize", "/v", "AppsUseLightTheme")
	output, err := cmd.Output()

	if err != nil {
		return false, err
	}
	return strings.Contains(string(output), "0x0"), nil
}

func checkLinuxTrayBackground() (bool, error) {
	// Try to get panel background color
	cmd := exec.Command("gsettings", "get", "org.gnome.desktop.interface", "gtk-theme")
	output, err := cmd.Output()

	if err == nil {
		theme := strings.ToLower(strings.TrimSpace(string(output)))
		if strings.Contains(theme, "dark") || strings.Contains(theme, "adwaita-dark") {
			return true, nil
		}
		if strings.Contains(theme, "light") || strings.Contains(theme, "adwaita") {
			return false, nil
		}
	}
	
	// Fallback to color scheme
	cmd = exec.Command("gsettings", "get", "org.gnome.desktop.interface", "color-scheme")
	output, err = cmd.Output()

	if err != nil {
		return false, err
	}
	return strings.Contains(string(output), "dark"), nil
}

func checkLinuxTheme() (bool, error) {
	cmd := exec.Command("gsettings", "get", "org.gnome.desktop.interface", "color-scheme")
	output, err := cmd.Output()

	if err != nil {
		return false, err
	}
	return strings.Contains(string(output), "dark"), nil
}
