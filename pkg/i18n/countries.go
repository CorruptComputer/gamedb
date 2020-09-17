package i18n

import (
	"github.com/gamedb/gamedb/pkg/log"
	"go.uber.org/zap"
)

func CountryCodeToName(code string) string {

	switch code {
	case "", "_":
		return "No Country"
	case "AX":
		return "Aland Islands"
	case "BQ":
		return "Bonaire, Sint Eustatius and Saba"
	case "SH":
		return "Saint Helena"
	case "XK":
		return "Kosovo"
	case "FX":
		return "France, Metropolitan"
	case "YU":
		return "Yugoslavia"
	}

	country, err := gountriesInstance.FindCountryByAlpha(code)
	if err != nil {
		log.Err(err.Error(), zap.String("code", code))
		return code
	}

	return country.Name.Common
}

func CountryCodeToContinent(code string) string {

	switch code {
	case "":
		return ""
	case "BQ":
		return ContinentSouthAmerica
	case "SH":
		return ContinentAfrica
	case "YU", "FX", "XK":
		return ContinentEurope
	}

	country, err := gountriesInstance.FindCountryByAlpha(code)
	if err != nil {
		log.ErrS(err, code)
		return ""
	}

	for _, v := range Continents {
		if v.Value == country.Continent {
			return v.Key
		}
	}

	return ""
}
