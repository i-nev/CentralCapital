// main
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Country struct {
	capital     string
	countryCode string
	countryName string
	lat         float64
	lng         float64
}

type Countries []Country

type CentralCapital struct {
	lat  float64
	lng  float64
	dist float64
}

func GetURL(rUrl *url.URL) ([]byte, error) {
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

func (cs *Countries) GetCapitals() {
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
	for i := range geonamesCounties["geonames"] {
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
			for j := range capitals {
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

func (cs *Countries) LoadCapitals(fromWeb bool) {
	if fromWeb {
		cs.GetCapitals()
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

func (cc *CentralCapital) distToCC(cs *Countries, ch chan CentralCapital) {
	var c Country
	c.lat = cc.lat
	c.lng = cc.lng
	cc.dist = 0
	for i := range *cs {
		cc.dist += dist(c, (*cs)[i])
	}
	ch <- *cc
}

func (cc *CentralCapital) String() string {
	return fmt.Sprintf("dist = %f (lat=%f, lng=%f)\n", cc.dist, cc.lat, cc.lng)
}

func (cs *Countries) FindCC(sLat, fLat, stepLat, sLng, fLng, stepLng float64) {
	var (
		lat      float64
		lng      float64
		startCnt int
		stopCnt  int
		minCC    CentralCapital
		flEnd    bool
	)
	chIn := make(chan CentralCapital, 100)
	chOut := make(chan CentralCapital)

	startCnt = 0
	flEnd = false
	go func() {
		for lat = sLat; lat <= fLat; lat += stepLat {
			for lng = sLng; lng <= fLng; lng += stepLng {
				var cc CentralCapital
				cc.lat = lat
				cc.lng = lng
				startCnt++
				chIn <- cc
			}
		}
		fmt.Print("#")
		flEnd = true
	}()

	stopCnt = 0
	for {
		select {
		case c, ok := <-chIn:
			if ok {
				go c.distToCC(cs, chOut)
				fmt.Print("-")
			}
		case c := <-chOut:
			stopCnt++
			if (stopCnt == 1) || (c.dist < minCC.dist) {
				minCC = c
				fmt.Print("m")
			}
			fmt.Print("+")
		default:
			if (flEnd) && (startCnt == stopCnt) {
				fmt.Println("stop")
				fmt.Println(minCC.String())
				return
			}
		}
	}

}

func main() {
	var countries Countries

	countries.LoadCapitals(true)

	st := time.Now().String()
	countries.FindCC(42.99341, 42.994, 0.00001, 16.2893, 16.29, 0.00001)
	fmt.Printf("%s\n%s\n", st, time.Now().String())
}
