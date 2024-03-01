package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type apiPhoneSettings struct {
	LineKeys         map[int]apiLineKey `json:"lineKeys"`
	AdvancedSettings []apiPhoneSetting  `json:"advancedSettings"`
}

type apiLineKey struct {
	Label       string
	Line        string
	PickupValue string
	Type        string
	Value       string
}

type apiPhoneSetting struct {
	SettingName  string  `json:"settingName"`
	SettingValue string  `json:"settingValue"`
	SettingNotes *string `json:"settingNotes"`
}

func generateConfigContent() (string, error) {

	var resultString string
	resultString = `#!version:1.0.0.1
## the file header "#!version:1.0.0.1" can not be edited or deleted. ##`

	s, err := getPhoneConfigSettings()
	if err != nil {
		return "", err
	}

	for i := 0; i < len(s.LineKeys); i++ {
		resultString += fmt.Sprintf("\nlinekey.%d.label = %s", i, s.LineKeys[i].Label)
		resultString += fmt.Sprintf("\nlinekey.%d.value = %s", i, s.LineKeys[i].Value)
		resultString += fmt.Sprintf("\nlinekey.%d.line = %s", i, s.LineKeys[i].Line)
		resultString += fmt.Sprintf("\nlinekey.%d.pickup_value = %s", i, s.LineKeys[i].PickupValue)
		resultString += fmt.Sprintf("\nlinekey.%d.type = %s", i, s.LineKeys[i].Type)
	}

	for _, setting := range s.AdvancedSettings {
		resultString += fmt.Sprintf("\n%s = %s", setting.SettingName, setting.SettingValue)
	}

	return resultString, nil
}

func getPhoneConfigSettings() (apiPhoneSettings, error) {
	var settings apiPhoneSettings
	lineKeys := make([]apiPhoneSetting, 0)
	advancedSettings := make([]apiPhoneSetting, 0)

	readFile, err := os.Open("y000000000028.cfg")
	if err != nil {
		fmt.Println(err)
	}
	defer readFile.Close()

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		key, val, found := strings.Cut(fileScanner.Text(), " = ")
		if !found {
			continue
		}

		var s apiPhoneSetting
		s.SettingName = key
		s.SettingValue = val

		nameParts := strings.Split(s.SettingName, ".")
		if nameParts[0] == "linekey" {
			lineKeys = append(lineKeys, s)
		} else {
			advancedSettings = append(advancedSettings, s)
		}
	}

	settings.LineKeys = mapToLineKeys(lineKeys)
	settings.AdvancedSettings = advancedSettings

	return settings, nil
}

func mapToLineKeys(lineKeys []apiPhoneSetting) map[int]apiLineKey {
	sortedLineKeys := make(map[int]apiLineKey)

	for _, s := range lineKeys {
		nameParts := strings.Split(s.SettingName, ".")
		keyNum, err := strconv.Atoi(nameParts[1])
		if err != nil {
			panic(err)
		}
		if _, ok := sortedLineKeys[keyNum]; !ok {
			sortedLineKeys[keyNum] = apiLineKey{}
		}
		lk := sortedLineKeys[keyNum]
		switch nameParts[2] {
		case "label":
			lk.Label = s.SettingValue
		case "line":
			lk.Line = s.SettingValue
		case "pickup_value":
			lk.PickupValue = s.SettingValue
		case "type":
			lk.Type = s.SettingValue
		case "value":
			lk.Value = s.SettingValue
		}
		sortedLineKeys[keyNum] = lk
	}

	return sortedLineKeys
}

func (a *appContext) handleAPIPhoneSettingsFileGet(rw http.ResponseWriter, r *http.Request) {

	s, err := generateConfigContent()
	if err != nil {
		panic(err)
	}

	fmt.Fprint(rw, s)
	rw.Header().Set("content-type", "application/json")
}

func (a *appContext) handleAPIPhoneSettingsPut(rw http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		panic(err)
	}

	newSettings := make([]apiPhoneSetting, 0)
	for i := range r.Form {
		var s apiPhoneSetting
		s.SettingName = i
		s.SettingValue = r.Form[i][0]

		newSettings = append(newSettings, s)
	}

	err = updatePhoneConfigSettings(newSettings)
	if err != nil {
		panic(err)
	}

	fmt.Println("Settings saved.")
	fmt.Fprint(rw, "Settings saved.")
	rw.WriteHeader(200)
}

func updatePhoneConfigSettings(newSettings []apiPhoneSetting) error {

	settingsFile, err := os.OpenFile("y000000000028.cfg", os.O_TRUNC|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer settingsFile.Close()

	var resultString string
	resultString = `#!version:1.0.0.1
## the file header "#!version:1.0.0.1" can not be edited or deleted. ##`

	for _, setting := range newSettings {
		resultString += fmt.Sprintf("\n%s = %s", setting.SettingName, setting.SettingValue)
	}
	_, err = settingsFile.WriteString(resultString)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (a *appContext) handleAPIPhoneSettingsFrontendGet(rw http.ResponseWriter, r *http.Request) {
	settings, err := getPhoneConfigSettings()
	if err != nil {
		panic(err)
	}
	var lineKeysHtml, advancedSettingsHtml string
	for i := 1; i <= len(settings.LineKeys); i++ {
		lineKeysHtml += fmt.Sprintf(`<span class="linekey">
		<input type="text" name="linekey.%d.label" placeholder="Name" value="%s" autocomplete="off" />
		<input type="number" name="linekey.%d.value" placeholder="Ext#" value="%s" autocomplete="off" inputmode="numeric" />
		<input type="hidden" name="linekey.%d.type" value="%s" />
		<input type="hidden" name="linekey.%d.line" value="%s" />
		<input type="hidden" name="linekey.%d.pickup_value" value="%s" />
	</span>`, i, settings.LineKeys[i].Label, i, settings.LineKeys[i].Value, i, settings.LineKeys[i].Type, i, settings.LineKeys[i].Line, i, settings.LineKeys[i].PickupValue)
	}
	for _, advSetting := range settings.AdvancedSettings {
		advancedSettingsHtml += fmt.Sprintf(`<label>%s: <input type="text" name="%s" value="%s" autocomplete="off" /></label>`, advSetting.SettingName, advSetting.SettingName, advSetting.SettingValue)
	}

	data, err := os.ReadFile("index.html")
	if err != nil {
		panic(err)
	}
	s := strings.Replace(strings.Replace(string(data), "{{lineKeys}}", lineKeysHtml, 1), "{{advancedSettings}}", advancedSettingsHtml, 1)
	fmt.Fprint(rw, s)

	rw.Header().Set("content-type", "text/html")
}
