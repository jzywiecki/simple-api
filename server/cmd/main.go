package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"server/types"
	"sort"
	"strings"
)

func main() {
	fs := http.FileServer(http.Dir("./html"))
	http.Handle("/", fs)
	mux := http.NewServeMux()

	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/api/get-listings", handleApiRequest)
	mux.HandleFunc("/api/auth", handleAuthRequest)

	http.ListenAndServe(":8080", mux)
}

func handleApiRequest(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://pro-api.coinmarketcap.com/v1/cryptocurrency/listings/latest", nil)
	if err != nil {
		log.Print(err)
		http.Error(w, "Error creating request", http.StatusInternalServerError)
		return
	}

	limit := r.FormValue("limit")
	order := r.FormValue("order")

	q := req.URL.Query()
	q.Add("sort", order)
	q.Add("limit", limit)
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Accepts", "application/json")
	req.Header.Add("X-CMC_PRO_API_KEY", "713a6b7d-6e93-4d59-88ea-038f57de2ae6")

	resp, err := client.Do(req)
	if err != nil {
		log.Print("Error sending request to server: ", err)
		http.Error(w, "Error sending request to server", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Received non-OK status code 1: %d", resp.StatusCode)
		http.Error(w, "Recieved status code 1:", resp.StatusCode)
		return
	}

	var responseData types.Response
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		log.Print("Error decoding JSON response: ", err)
		http.Error(w, "Error decoding JSON response", http.StatusInternalServerError)
		return
	}

	nameToCheckInOtherApi := responseData.Data[0].Name

	req2, err := http.NewRequest("GET", "https://api.coingecko.com/api/v3/coins/list", nil)
	if err != nil {
		log.Print(err)
		http.Error(w, "Error creating request", http.StatusInternalServerError)
		return
	}

	req2.Header.Add("x-cg-demo-api-key", "CG-x46kYuMHifPvVb46Qxj8WnRs")

	resp2, err := client.Do(req2)
	if err != nil {
		log.Print(err)
		http.Error(w, "Error sending request to server", http.StatusInternalServerError)
		return
	}

	if resp2.StatusCode != http.StatusOK {
		log.Printf("Received non-OK status code 2: %d", resp2.StatusCode)
		http.Error(w, "Recieved status code 2:", resp2.StatusCode)
		return
	}

	var responseData2 []types.Coin
	if err := json.NewDecoder(resp2.Body).Decode(&responseData2); err != nil {
		log.Print("Error decoding JSON response: ", err)
		http.Error(w, "Error decoding JSON response", http.StatusInternalServerError)
		return
	}

	for _, coin := range responseData2 {
		if strings.EqualFold(coin.Name, nameToCheckInOtherApi) {
			nameToCheckInOtherApi = coin.ID
			break
		}
		fmt.Println(coin.Name)
	}

	req3, err := http.NewRequest("GET", "https://api.coingecko.com/api/v3/simple/price?ids="+nameToCheckInOtherApi+"&vs_currencies=usd", nil)
	if err != nil {
		log.Print(err)
		http.Error(w, "Error creating request", http.StatusInternalServerError)
		return
	}

	req3.Header.Add("x-cg-demo-api-key", "CG-x46kYuMHifPvVb46Qxj8WnRs")

	resp3, err := client.Do(req3)
	if err != nil {
		log.Print(err)
		http.Error(w, "Error sending request to server", http.StatusInternalServerError)
		return
	}
	defer resp3.Body.Close()

	if resp3.StatusCode != http.StatusOK {
		log.Printf("Received non-OK status code 3: %d", resp3.StatusCode)
		http.Error(w, "Recieved status code 3:", resp3.StatusCode)
		return
	}

	var responseData3 map[string]map[string]float64
	if err := json.NewDecoder(resp3.Body).Decode(&responseData3); err != nil {
		log.Print("Error decoding JSON response: ", err)
		http.Error(w, "Error decoding JSON response", http.StatusInternalServerError)
		return
	}

	priceOfCoinOtherApi := responseData3[nameToCheckInOtherApi]["usd"]

	average := CalculateAverage(responseData)
	median := CalculateMedian(responseData)
	standardDeviation := CalculateStandardDeviation(responseData)

	tmpl := template.Must(template.ParseFiles("html/index.html"))

	templateData := types.ResponseToHttp{
		Response:            responseData,
		Average:             average,
		Median:              median,
		StandardDeviation:   standardDeviation,
		Max:                 CalculateMax(responseData),
		Min:                 CalculateMin(responseData),
		PriceOfCoinOtherApi: priceOfCoinOtherApi,
	}

	err = tmpl.Execute(w, templateData)

	if err != nil {
		log.Println("Error executing template:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

}

func CalculateAverage(listings types.Response) float64 {
	var sum float64
	for _, listing := range listings.Data {
		sum += listing.Quote["USD"].Price
	}
	return sum / float64(len(listings.Data))
}

func CalculateMedian(listings types.Response) float64 {
	prices := make([]float64, len(listings.Data))
	for i, listing := range listings.Data {
		prices[i] = listing.Quote["USD"].Price
	}
	sort.Float64s(prices)

	middle := len(prices) / 2
	median := prices[middle]
	if len(prices)%2 == 0 {
		median = (median + prices[middle-1]) / 2
	}
	return median
}

func CalculateStandardDeviation(listings types.Response) float64 {
	var sum float64
	average := CalculateAverage(listings)
	for _, listing := range listings.Data {
		deviation := listing.Quote["USD"].Price - average
		sum += deviation * deviation
	}
	variance := sum / float64(len(listings.Data))
	return math.Sqrt(variance)
}

func CalculateMax(listings types.Response) float64 {
	var max float64
	for _, listing := range listings.Data {
		if listing.Quote["USD"].Price > max {
			max = listing.Quote["USD"].Price
		}
	}
	return max
}

func CalculateMin(listings types.Response) float64 {
	min := listings.Data[0].Quote["USD"].Price
	for _, listing := range listings.Data {
		if listing.Quote["USD"].Price < min {
			min = listing.Quote["USD"].Price
		}
	}
	return min
}

func homeHandler(rw http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("html/index.html"))
	var data interface{} = nil

	err := tmpl.Execute(rw, data)
	if err != nil {
		panic(err)
	}
}

func handleAuthRequest(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Auth request received"))
}
