package main

import (
	"io/ioutil"
	"encoding/json"
	"bytes"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
)

// Abscissa : Structure for information about specific time
type Abscissa struct {
	ID int `json:"id"`
	Year int `json:"year"`
	Month int `json:"month"`
	Day int `json:"day"`
	Name string `json:"name"`
	Date string `json:"date"`
}

// Point : Structure for data at a specific time
type Point struct {
	Abscissa Abscissa `json:"abscissa"`
	Ordinate float32 `json:"ordinate"`
}

// AreaData : Structure for data about a specific area
type AreaData struct {
	ID int `json:"id"`
	Name string `json:"name"`
	Color string `json:"color"`
	Points []Point `json:"points"`
}

// DataInTime : Structure for total data at specific time
type DataInTime struct {
	Confirmed float32 `json:"confirmed"`
	NewConfirmed float32 `json:"newConfirmed"`
}

// ParsedAreaData : Structure for parsed area data
type ParsedAreaData struct {
	ID int `json:"id"`
	Name string `json:"name"`
	Color string `json:"color"`
	Timeline []DataInTime `json:"timeline"`
}

// ParsedData : Structure for parsed data
type ParsedData struct {
	Timestamps []string `json:"timestamps"`
	TotalData []DataInTime `json:"totalData"`
	AreaData []ParsedAreaData `json:"areaData"`
}

func getData(data *ParsedData) (error) {
	jsonBody := []byte(`{"dataSetId":1,"refCodes":[{"id":1,"code":"COVID-19 статистике заражени","values":[{"id":2,"name":"Заражено укупно"}]}],"territoryIds":[168,40,41,169,170,42,43,44,45,46,171,172,173,174,47,175,176,177,64,65,66,67,68,69,70,71,72,73,74,75,215,76,77,78,79,161,210,80,178,216,81,217,218,82,242,83,163,164,84,219,85,86,220,179,87,88,180,89,90,243,181,221,91,182,183,222,133,184,223,185,92,224,93,94,186,187,95,225,226,238,160,96,97,98,99,100,188,101,102,103,153,104,227,105,228,106,107,108,109,110,111,189,112,113,114,115,116,117,190,191,192,118,193,229,230,194,231,119,195,196,120,213,121,232,197,122,198,233,123,124,125,126,234,127,235,128,129,130,131,132,199,134,135,162,200,201,136,137,138,139,202,236,203,204,205,206,207,237,140,208,209,142,143,144,145,146,147,148,141,239,149,150,151,211,152,212,240,241]}`)

	resp, err := http.Post("https://covid19.data.gov.rs/api/datasets/statistic", "application/json;charset=UTF-8", bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var areaList []AreaData
	json.Unmarshal(body, &areaList)

	var parsedData ParsedData
	parsedData.TotalData = make([]DataInTime, len(areaList[0].Points))

	for _, area := range areaList {
		var parsedArea ParsedAreaData
		parsedArea.ID = area.ID
		parsedArea.Name = area.Name
		parsedArea.Color = area.Color
		for j, point := range area.Points {
			var newDataInTime DataInTime
			if len(parsedArea.Timeline) > 0 {
				newDataInTime.Confirmed = point.Ordinate
				newDataInTime.NewConfirmed = point.Ordinate - area.Points[j - 1].Ordinate
				parsedArea.Timeline = append(parsedArea.Timeline, newDataInTime)
			} else {
				newDataInTime.Confirmed = point.Ordinate
				newDataInTime.NewConfirmed = point.Ordinate
				parsedArea.Timeline = append(parsedArea.Timeline, newDataInTime)
			}

			parsedData.TotalData[j].Confirmed += newDataInTime.Confirmed
			parsedData.TotalData[j].NewConfirmed += newDataInTime.NewConfirmed
		}

		parsedData.AreaData = append(parsedData.AreaData, parsedArea)
	}

	for _, point := range areaList[0].Points {
		parsedData.Timestamps = append(parsedData.Timestamps, point.Abscissa.Date)
	}

	*data = parsedData
	return nil
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.Use(cors.Default())

	router.GET("/", func (ctx *gin.Context) {
		var data ParsedData
		if err := getData(&data); err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, data)
	})

	router.Run()
}
