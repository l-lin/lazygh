package theme

import (
	"errors"
	"fmt"
	"testing"
)

func TestDetectSystemPolarityForOS_GivenDarwinDarkAppearance_WhenDetecting_ThenItReturnsDark(t *testing.T) {
	runner := given_systemCommandRunner(map[string]fakeSystemCommandResponse{
		"defaults read -g AppleInterfaceStyle": {output: "Dark\n"},
	})

	actual := detectSystemPolarityForOS("darwin", runner)

	if actual != systemPolarityDark {
		t.Fatalf("expected system polarity %v, actual %v", systemPolarityDark, actual)
	}
}

func TestDetectSystemPolarityForOS_GivenDarwinLightAppearance_WhenDetecting_ThenItReturnsLight(t *testing.T) {
	runner := given_systemCommandRunner(map[string]fakeSystemCommandResponse{
		"defaults read -g AppleInterfaceStyle": {
			output: "The domain/default pair of (kCFPreferencesAnyApplication, AppleInterfaceStyle) does not exist\n",
			err:    errors.New("exit status 1"),
		},
	})

	actual := detectSystemPolarityForOS("darwin", runner)

	if actual != systemPolarityLight {
		t.Fatalf("expected system polarity %v, actual %v", systemPolarityLight, actual)
	}
}

func TestDetectSystemPolarityForOS_GivenLinuxPreferDark_WhenDetecting_ThenItReturnsDark(t *testing.T) {
	runner := given_systemCommandRunner(map[string]fakeSystemCommandResponse{
		"gsettings get org.gnome.desktop.interface color-scheme": {output: "'prefer-dark'\n"},
	})

	actual := detectSystemPolarityForOS("linux", runner)

	if actual != systemPolarityDark {
		t.Fatalf("expected system polarity %v, actual %v", systemPolarityDark, actual)
	}
}

func TestDetectSystemPolarityForOS_GivenLinuxDarkGTKTheme_WhenDetecting_ThenItReturnsDark(t *testing.T) {
	runner := given_systemCommandRunner(map[string]fakeSystemCommandResponse{
		"gsettings get org.gnome.desktop.interface color-scheme": {
			output: "No such key\n",
			err:    errors.New("exit status 1"),
		},
		"gsettings get org.gnome.desktop.interface gtk-theme": {output: "'Adwaita-dark'\n"},
	})

	actual := detectSystemPolarityForOS("linux", runner)

	if actual != systemPolarityDark {
		t.Fatalf("expected system polarity %v, actual %v", systemPolarityDark, actual)
	}
}

func TestDetectSystemPolarityForOS_GivenWindowsDarkAppsTheme_WhenDetecting_ThenItReturnsDark(t *testing.T) {
	runner := given_systemCommandRunner(map[string]fakeSystemCommandResponse{
		`reg query HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Themes\\Personalize /v AppsUseLightTheme`: {output: "HKEY_CURRENT_USER\\Software\\Microsoft\\Windows\\CurrentVersion\\Themes\\Personalize\n    AppsUseLightTheme    REG_DWORD    0x0\n"},
	})

	actual := detectSystemPolarityForOS("windows", runner)

	if actual != systemPolarityDark {
		t.Fatalf("expected system polarity %v, actual %v", systemPolarityDark, actual)
	}
}

func TestDetectSystemPolarityForOS_GivenAnUnknownPlatform_WhenDetecting_ThenItReturnsUnknown(t *testing.T) {
	actual := detectSystemPolarityForOS("plan9", given_systemCommandRunner(nil))

	if actual != systemPolarityUnknown {
		t.Fatalf("expected system polarity %v, actual %v", systemPolarityUnknown, actual)
	}
}

type fakeSystemCommandResponse struct {
	output string
	err    error
}

func given_systemCommandRunner(responses map[string]fakeSystemCommandResponse) systemCommandRunner {
	return func(name string, args ...string) (string, error) {
		commandKey := fmt.Sprintf("%s %s", name, joinCommandArguments(args))
		response, ok := responses[commandKey]
		if !ok {
			return "", fmt.Errorf("unexpected system command %q", commandKey)
		}

		return response.output, response.err
	}
}

func joinCommandArguments(arguments []string) string {
	if len(arguments) == 0 {
		return ""
	}

	joined := arguments[0]
	for _, argument := range arguments[1:] {
		joined += " " + argument
	}
	return joined
}
