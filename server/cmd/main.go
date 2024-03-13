package main

import (
	"encoding/json"
	"html/template"
	"log"
	"math"
	"net/http"
	"server/types"
	"sort"
	"strings"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/api/get-listings", handleApiRequest)
	http.ListenAndServe(":8080", mux)
}

func handleApiRequest(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{}

	if r.FormValue("api-key") != "123" {
		http.Error(w, "Invalid API key", http.StatusUnauthorized)
		return
	}

	responseDataCh := make(chan types.Response)
	responseDataCh2 := make(chan []types.Coin)
	priceOfCoinOtherApiCh := make(chan float64)
	errorCh := make(chan error)

	go func() {
		responseData, err := createCoinMarketCapRequest(w, r, client)

		if err != nil {
			errorCh <- err
		} else {
			responseDataCh <- responseData
			responseDataCh <- responseData
		}
	}()

	go func() {
		responseData2, err := createCoinGeckoRequest(w, client)

		if err != nil {
			errorCh <- err
		} else {
			responseDataCh2 <- responseData2
		}
	}()

	go func() {
		responseData := <-responseDataCh
		responseData2 := <-responseDataCh2

		nameToCheckInOtherApi := responseData.Data[0].Name

		nameToCheckInOtherApi = findCoinID(responseData2, nameToCheckInOtherApi)

		priceOfCoinOtherApi := createCoinGeckoPriceRequest(w, nameToCheckInOtherApi, client)

		priceOfCoinOtherApiCh <- priceOfCoinOtherApi
	}()

	priceOfCoinOtherApi := <-priceOfCoinOtherApiCh
	responseData := <-responseDataCh

	averageCh := make(chan float64)
	medianCh := make(chan float64)
	standardDeviationCh := make(chan float64)

	go func() {
		average := CalculateAverage(responseData)
		averageCh <- average
	}()

	go func() {
		median := CalculateMedian(responseData)
		medianCh <- median
	}()

	go func() {
		standardDeviation := CalculateStandardDeviation(responseData)
		standardDeviationCh <- standardDeviation
	}()

	average := <-averageCh
	median := <-medianCh
	standardDeviation := <-standardDeviationCh

	tmpl := template.Must(template.ParseFiles("html/results.html"))

	templateData := types.ResponseToHttp{
		Response:            responseData,
		Average:             average,
		Median:              median,
		StandardDeviation:   standardDeviation,
		Max:                 CalculateMax(responseData),
		Min:                 CalculateMin(responseData),
		PriceOfCoinOtherApi: priceOfCoinOtherApi,
	}

	err := tmpl.Execute(w, templateData)

	if err != nil {
		log.Println("Error executing template:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

}

func createCoinMarketCapRequest(w http.ResponseWriter, r *http.Request, client *http.Client) (types.Response, error) {
	limit := r.FormValue("limit")
	order := r.FormValue("order")

	req, err := http.NewRequest("GET", "https://pro-api.coinmarketcap.com/v1/cryptocurrency/listings/latest", nil)
	if err != nil {
		return types.Response{}, err
	}

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
		return types.Response{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Received non-OK status code 1: %d", resp.StatusCode)
		http.Error(w, "Recieved status code 1:", resp.StatusCode)
		return types.Response{}, err
	}

	var responseData types.Response
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		log.Print("Error decoding JSON response: ", err)
		http.Error(w, "Error decoding JSON response", http.StatusInternalServerError)
		return types.Response{}, err
	}

	return responseData, nil
}

func createCoinGeckoRequest(w http.ResponseWriter, client *http.Client) ([]types.Coin, error) {
	req2, err := http.NewRequest("GET", "https://api.coingecko.com/api/v3/coins/list", nil)
	if err != nil {
		log.Print(err)
		http.Error(w, "Error creating request", http.StatusInternalServerError)
		return []types.Coin{}, err
	}

	req2.Header.Add("x-cg-demo-api-key", "CG-x46kYuMHifPvVb46Qxj8WnRs")

	resp2, err := client.Do(req2)
	if err != nil {
		log.Print(err)
		http.Error(w, "Error sending request to server", http.StatusInternalServerError)
		return []types.Coin{}, err
	}

	if resp2.StatusCode != http.StatusOK {
		log.Printf("Received non-OK status code 2: %d", resp2.StatusCode)
		http.Error(w, "Recieved status code 2:", resp2.StatusCode)
		return []types.Coin{}, err
	}

	var responseData2 []types.Coin
	if err := json.NewDecoder(resp2.Body).Decode(&responseData2); err != nil {
		log.Print("Error decoding JSON response: ", err)
		http.Error(w, "Error decoding JSON response", http.StatusInternalServerError)
		return []types.Coin{}, err
	}

	return responseData2, nil
}

func createCoinGeckoPriceRequest(w http.ResponseWriter, coinID string, client *http.Client) float64 {
	req3, err := http.NewRequest("GET", "https://api.coingecko.com/api/v3/simple/price?ids="+coinID+"&vs_currencies=usd", nil)
	if err != nil {
		log.Print(err)
		http.Error(w, "Error creating request", http.StatusInternalServerError)
		return 0
	}

	req3.Header.Add("x-cg-demo-api-key", "CG-x46kYuMHifPvVb46Qxj8WnRs")

	resp3, err := client.Do(req3)
	if err != nil {
		log.Print(err)
		http.Error(w, "Error sending request to server", http.StatusInternalServerError)
		return 0
	}
	defer resp3.Body.Close()

	if resp3.StatusCode != http.StatusOK {
		log.Printf("Received non-OK status code 3: %d", resp3.StatusCode)
		http.Error(w, "Recieved status code 3:", resp3.StatusCode)
		return 0
	}

	var responseData3 map[string]map[string]float64
	if err := json.NewDecoder(resp3.Body).Decode(&responseData3); err != nil {
		log.Print("Error decoding JSON response: ", err)
		http.Error(w, "Error decoding JSON response", http.StatusInternalServerError)
		return 0
	}

	return responseData3[coinID]["usd"]
}

func findCoinID(coins []types.Coin, nameToCheckInOtherApi string) string {
	for _, coin := range coins {
		if strings.EqualFold(coin.Name, nameToCheckInOtherApi) {
			return coin.ID
		}
	}
	return nameToCheckInOtherApi
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
