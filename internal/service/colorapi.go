package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

const (
	idEndpoint     = "https://www.thecolorapi.com/id?hex=%s"
	schemeEndpoint = "https://www.thecolorapi.com/scheme?hex=%s"
)

// ColorResponse is returned by CallColorApi.
type ColorResponse struct {
	Name    string // e.g. "Cobalt"
	Match   bool   // exact_match_name
	Palette string // comma-joined list of hex values, e.g. "#123,#456,#789"
}

// CallColorApi fetches the name & match-flag, then the palette.
// It will strip any leading “#” from the input before calling the API.
func CallColorApi(hex string) ColorResponse {
	cleanHex := normalizeHex(hex)

	name, match, err := fetchColorName(cleanHex)
	if err != nil {
		log.Printf("fetchColorName error: %v", err)
	}

	colors, err := fetchColorScheme(cleanHex)
	if err != nil {
		log.Printf("fetchColorScheme error: %v", err)
	}

	return ColorResponse{
		Name:    name,
		Match:   match,
		Palette: strings.Join(colors, ","),
	}
}

// normalizeHex removes a leading “#” if present.
func normalizeHex(hex string) string {
	return strings.TrimPrefix(hex, "#")
}

// fetchColorName calls the “id” endpoint and returns the name & exact_match_name.
func fetchColorName(hex string) (name string, match bool, err error) {
	url := fmt.Sprintf(idEndpoint, hex)
	resp, err := http.Get(url)
	if err != nil {
		return "", false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", false, err
	}

	var payload struct {
		Name struct {
			Value string `json:"value"`
		} `json:"name"`
		ExactMatchName bool `json:"exact_match_name"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", false, err
	}

	return payload.Name.Value, payload.ExactMatchName, nil
}

// fetchColorScheme calls the “scheme” endpoint and returns the list of hex values.
func fetchColorScheme(hex string) ([]string, error) {
	url := fmt.Sprintf(schemeEndpoint, hex)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var payload struct {
		Colors []struct {
			Hex struct {
				Value string `json:"value"`
			} `json:"hex"`
		} `json:"colors"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	hexes := make([]string, len(payload.Colors))
	for i, c := range payload.Colors {
		hexes[i] = c.Hex.Value
	}
	return hexes, nil
}
