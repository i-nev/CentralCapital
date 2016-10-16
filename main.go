// main
package main

import (
	"encoding/json"
	"fmt"
	//"io"
	"io/ioutil"
	"math"
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

func GetCapitals(cs *[]Country) {
	var (
		mUrl             url.URL
		query            url.Values
		err              error
		jsonCounties     []byte
		jsonCapital      []byte
		geonamesCounties map[string][]map[string]interface{}
		geonamesCapital  map[string]interface{}
		capitals         []interface{}
		cap              map[string]interface{}
		c                Country
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
				*cs = append(*cs, c)
			} else {
				fmt.Printf("Oops: no PPLC for capital %s!\n", c.capital)
			}
			caCount++
		} else {
			fmt.Printf("Oops: country %s has no capital!\n", c.countryName)
		}
		coCount++
	}
	fmt.Printf("Total: %d countries, %d capitals, %d coordinates\n", coCount, caCount, coordCount)

	jsonCapital, err = json.Marshal(*cs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", string(jsonCapital))
	err = ioutil.WriteFile("capitals.json", jsonCapital, 0777)
	if err != nil {
		panic(err)
	}
}

func LoadCapitals(cs *[]Country, fromWeb bool) {
	if fromWeb {
		GetCapitals(cs)
	}
}

func dist(a Country, b Country) float64 {
	const rad = 6372795

	aLat := a.lat * math.Pi / 180
	bLat := b.lat * math.Pi / 180
	aLng := a.lng * math.Pi / 180
	bLng := b.lng * math.Pi / 180

	cl1 := math.Cos(aLat)
	cl2 := math.Cos(bLat)
	sl1 := math.Sin(aLat)
	sl2 := math.Sin(bLat)
	delta := bLng - aLng
	cdelta := math.Cos(delta)
	sdelta := math.Sin(delta)

	y := math.Sqrt(math.Pow(cl2*sdelta, 2) + math.Pow(cl1*sl2-sl1*cl2*cdelta, 2))
	x := sl1*sl2 + cl1*cl2*cdelta
	ad := math.Atan2(y, x)
	return ad * rad
}

func distToCC(cc Country, cs *[]Country) float64 {
	var total float64 = 0
	for _, c := range *cs {
		total += dist(cc, c)
	}
	return total
}

func main() {
	var (
		countries []Country
		lat       float64
		lng       float64
	)
	LoadCapitals(&countries, true)

	for lat = -90; lat <= 90; lat += 20 {
		for lng = -180; lng <= 180; lng += 20 {
			var cc Country
			cc.lat = lat
			cc.lng = lng
			fmt.Printf("dist= %v\n", distToCC(cc, &countries))
		}
	}

}
