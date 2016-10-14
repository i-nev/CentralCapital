// main
package main

import (
	"encoding/json"
	"fmt"
	//"io"
	"io/ioutil"
	"net/http"
	//"strings"
	"net/url"
	"strconv"
)

type Country struct {
	capital     string
	countryCode string
	countryName string
	lat         float64
	lng         float64
}

func GetURL(rUrl *url.URL) ([]byte, error) {
	//fmt.Printf("%s", rUrl.String())
	res, err := http.Get(rUrl.String())
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, err
	}
	return body, nil
}

func main() {
	var (
		mUrl             url.URL
		query            url.Values
		err              error
		jsonCounties     []byte
		jsonCapital      []byte
		geonamesCounties map[string][]map[string]interface{}
		geonamesCapital  map[string]interface{}
		capitals         []interface{}
		countries        []Country
		cap              map[string]interface{}
	)
	mUrl.Scheme = "http"
	mUrl.Host = "api.geonames.org"
	mUrl.Path = "countryInfoJSON"
	query = mUrl.Query()

	query.Set("username", "nevedrov")
	query.Set("style", "full")
	query.Set("formatted", "true")
	mUrl.RawQuery = query.Encode()
	jsonCounties, err = GetURL(&mUrl)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(jsonCounties, &geonamesCounties)
	if err != nil {
		panic(err)
	}
	mUrl.Path = "search"
	query.Del("formatted")

	query.Set("type", "json")
	query.Set("style", "short")
	coCount := 0
	caCount := 0
	coordCount := 0
	for i, _ := range geonamesCounties["geonames"] {
		var c Country
		c.capital = geonamesCounties["geonames"][i]["capital"].(string)
		c.countryCode = geonamesCounties["geonames"][i]["countryCode"].(string)
		c.countryName = geonamesCounties["geonames"][i]["countryName"].(string)
		if (len(c.capital) > 0) && (len(c.countryCode) > 0) {
			query.Set("name_equals", c.capital)
			query.Set("country", c.countryCode)
			query.Set("featureCode", "PPLC")
			mUrl.RawQuery = query.Encode()
			jsonCapital, err = GetURL(&mUrl)
			if err != nil {
				panic(err)
			}
			//fmt.Printf("%v\n", c)
			err = json.Unmarshal(jsonCapital, &geonamesCapital)
			if err != nil {
				panic(err)
			}
			capitals = geonamesCapital["geonames"].([]interface{})
			findCoord := false
			for j, _ := range capitals {
				cap = capitals[j].(map[string]interface{})
				if (cap["fcode"] != nil) && (cap["fcode"].(string) == "PPLC") {
					c.lat, _ = strconv.ParseFloat(cap["lat"].(string), 64)
					c.lng, _ = strconv.ParseFloat(cap["lng"].(string), 64)
					findCoord = true
				}
			}
			if findCoord {
				coordCount++
			} else {
				fmt.Printf("Oops: no PPLC for capital %s!\n", c.capital)
			}
			caCount++
		} else {
			fmt.Printf("Oops: country %s has no capital!\n", c.countryName)
		}

		countries = append(countries, c)
		coCount++
	}
	fmt.Printf("Total: %d countries, %d capitals, %d coordinates\n", coCount, caCount, coordCount)

	//for i, c := range countries {
	//	fmt.Printf("%v -> %v\n", i, c)
	//}

}
