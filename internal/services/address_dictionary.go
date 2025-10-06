package services

import (
	"strings"

	"github.com/agnivade/levenshtein"
)

var USStates = map[string]string{
	"al": "alabama", "ak": "alaska", "az": "arizona", "ar": "arkansas",
	"ca": "california", "co": "colorado", "ct": "connecticut", "de": "delaware",
	"fl": "florida", "ga": "georgia", "hi": "hawaii", "id": "idaho",
	"il": "illinois", "in": "indiana", "ia": "iowa", "ks": "kansas",
	"ky": "kentucky", "la": "louisiana", "me": "maine", "md": "maryland",
	"ma": "massachusetts", "mi": "michigan", "mn": "minnesota", "ms": "mississippi",
	"mo": "missouri", "mt": "montana", "ne": "nebraska", "nv": "nevada",
	"nh": "new hampshire", "nj": "new jersey", "nm": "new mexico", "ny": "new york",
	"nc": "north carolina", "nd": "north dakota", "oh": "ohio", "ok": "oklahoma",
	"or": "oregon", "pa": "pennsylvania", "ri": "rhode island", "sc": "south carolina",
	"sd": "south dakota", "tn": "tennessee", "tx": "texas", "ut": "utah",
	"vt": "vermont", "va": "virginia", "wa": "washington", "wv": "west virginia",
	"wi": "wisconsin", "wy": "wyoming", "dc": "district of columbia",
}

var StateAbbreviations = map[string]string{
	"alabama": "AL", "alaska": "AK", "arizona": "AZ", "arkansas": "AR",
	"california": "CA", "colorado": "CO", "connecticut": "CT", "delaware": "DE",
	"florida": "FL", "georgia": "GA", "hawaii": "HI", "idaho": "ID",
	"illinois": "IL", "indiana": "IN", "iowa": "IA", "kansas": "KS",
	"kentucky": "KY", "louisiana": "LA", "maine": "ME", "maryland": "MD",
	"massachusetts": "MA", "michigan": "MI", "minnesota": "MN", "mississippi": "MS",
	"missouri": "MO", "montana": "MT", "nebraska": "NE", "nevada": "NV",
	"new hampshire": "NH", "new jersey": "NJ", "new mexico": "NM", "new york": "NY",
	"north carolina": "NC", "north dakota": "ND", "ohio": "OH", "oklahoma": "OK",
	"oregon": "OR", "pennsylvania": "PA", "rhode island": "RI", "south carolina": "SC",
	"south dakota": "SD", "tennessee": "TN", "texas": "TX", "utah": "UT",
	"vermont": "VT", "virginia": "VA", "washington": "WA", "west virginia": "WV",
	"wisconsin": "WI", "wyoming": "WY", "district of columbia": "DC",
}

var CommonStreetTypes = []string{
	"street", "avenue", "boulevard", "road", "drive", "lane", "court",
	"place", "way", "circle", "parkway", "terrace", "trail", "highway",
	"plaza", "alley", "bridge", "expressway", "freeway", "walk", "square",
}

var CommonCityNames = []string{
	"new york", "los angeles", "chicago", "houston", "phoenix", "philadelphia",
	"san antonio", "san diego", "dallas", "san jose", "austin", "jacksonville",
	"fort worth", "columbus", "charlotte", "francisco", "indianapolis", "seattle",
	"denver", "washington", "boston", "el paso", "nashville", "detroit", "oklahoma",
	"portland", "las vegas", "memphis", "louisville", "baltimore", "milwaukee",
	"albuquerque", "tucson", "fresno", "mesa", "sacramento", "atlanta", "kansas",
	"colorado springs", "omaha", "raleigh", "miami", "long beach", "virginia beach",
	"oakland", "minneapolis", "tulsa", "tampa", "arlington", "new orleans",
}

var StreetAbbreviations = map[string]string{
	"st": "street", "st.": "street",
	"ave": "avenue", "ave.": "avenue", "av": "avenue",
	"blvd": "boulevard", "blvd.": "boulevard",
	"rd": "road", "rd.": "road",
	"dr": "drive", "dr.": "drive",
	"ln": "lane", "ln.": "lane",
	"ct": "court", "ct.": "court",
	"pl": "place", "pl.": "place",
	"pkwy": "parkway", "pkwy.": "parkway",
	"ter": "terrace", "ter.": "terrace",
	"trl": "trail", "trl.": "trail",
	"hwy": "highway", "hwy.": "highway",
	"cir": "circle", "cir.": "circle",
	"sq": "square", "sq.": "square",
	"aly": "alley", "aly.": "alley",
	"expy": "expressway", "expy.": "expressway",
	"fwy": "freeway", "fwy.": "freeway",
}

var DirectionAbbreviations = map[string]string{
	"n": "north", "n.": "north",
	"s": "south", "s.": "south",
	"e": "east", "e.": "east",
	"w": "west", "w.": "west",
	"ne": "northeast", "ne.": "northeast",
	"nw": "northwest", "nw.": "northwest",
	"se": "southeast", "se.": "southeast",
	"sw": "southwest", "sw.": "southwest",
}

func FindClosestMatch(word string, dictionary []string, maxDistance int) (string, bool) {
	word = strings.ToLower(word)
	bestMatch := ""
	bestDistance := maxDistance + 1

	for _, candidate := range dictionary {
		distance := levenshtein.ComputeDistance(word, strings.ToLower(candidate))
		if distance < bestDistance && distance <= maxDistance {
			bestDistance = distance
			bestMatch = candidate
		}
	}

	return bestMatch, bestMatch != ""
}

func IsValidUSState(state string) bool {
	state = strings.ToLower(strings.TrimSpace(state))

	if _, exists := USStates[state]; exists {
		return true
	}

	if _, exists := StateAbbreviations[state]; exists {
		return true
	}

	return false
}

func NormalizeUSState(state string) (string, bool) {
	state = strings.ToLower(strings.TrimSpace(state))

	if fullName, exists := USStates[state]; exists {
		return StateAbbreviations[fullName], true
	}

	if abbr, exists := StateAbbreviations[state]; exists {
		return abbr, true
	}

	stateNames := make([]string, 0, len(StateAbbreviations))
	for name := range StateAbbreviations {
		stateNames = append(stateNames, name)
	}

	if match, found := FindClosestMatch(state, stateNames, 2); found {
		return StateAbbreviations[match], true
	}

	return "", false
}
