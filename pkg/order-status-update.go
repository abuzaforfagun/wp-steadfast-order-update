package pkg

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/abuzaforfagun/wp-steadfast-order-update/models"
)

var wpBaseAddress, wpApiKey, wpApiSecret, steadFastApiKey, steadFastApiSecret string
var steadFAstApiBaseAddress = "https://portal.packzy.com/api/v1"
var wg sync.WaitGroup

func VerifyWp(wpHostAddress string, wpConsumerKey string, wpConsumerSecret string, statuses []string, destinationStatuses map[string]string) error {
	var destinationStatus []string

	for _, wpStatus := range destinationStatuses {
		destinationStatus = append(destinationStatus, wpStatus)
	}

	wpStatuses := append(statuses, destinationStatus...)

	wpBaseAddress = fmt.Sprintf("%s/wp-json/wc/v3", wpHostAddress)
	wpApiKey = wpConsumerKey
	wpApiSecret = wpConsumerSecret

	_, err := getOrders(wpStatuses)

	return err
}

func VerifySteadfast(courierApiKey string, courierApiSecret string, destinationStatus map[string]string) error {
	steadFastApiKey = courierApiKey
	steadFastApiSecret = courierApiSecret
	_, err := getBalance()
	if err != nil {
		return err
	}

	availableStatuses := map[string]bool{
		"pending":                            true,
		"delivered_approval_pending":         true,
		"partial_delivered_approval_pending": true,
		"cancelled_approval_pending":         true,
		"unknown_approval_pending":           true,
		"delivered":                          true,
		"partial_delivered":                  true,
		"cancelled":                          true,
		"hold":                               true,
		"in_review":                          true,
		"unknown":                            true,
	}

	for steadFastStatus, _ := range destinationStatus {
		_, ok := availableStatuses[steadFastStatus]
		if !ok {
			return errors.New("Invalid steadfast status")
		}
	}

	return nil
}

func ExecuteProcessOrder(
	wpHostAddress string,
	wpConsumerKey string,
	wpConsumerSecret string,
	courierApiKey string,
	courierApiSecret string,
	statuses []string,
	destinationStatuses map[string]string,
) {
	wpBaseAddress = fmt.Sprintf("%s/wp-json/wc/v3", wpHostAddress)
	wpApiKey = wpConsumerKey
	wpApiSecret = wpConsumerSecret
	steadFastApiKey = courierApiKey
	steadFastApiSecret = courierApiSecret

	for _, status := range statuses {
		err := processOrders(status, destinationStatuses)
		if err != nil {
			log.Println("Unable to process orders from `%s` status", status)
		}
	}

	// err = processOrders("hold")
	// if err != nil {
	// 	log.Println("Unable to process orders from `hold` status")
	// }

	// err = processOrders("in-review")
	// if err != nil {
	// 	log.Println("Unable to process orders from `in-review` status")
	// }

	fmt.Println("Process completed!")
}

func processOrders(status string, destinationStatus map[string]string) error {
	fmt.Printf("Start processing orders from `%s` status\n", status)
	orders, err := getAllOrders([]string{status}, 100)
	if err != nil {
		return err
	}

	err = checkAndUpdateOrders(orders, destinationStatus)

	if err != nil {
		return err
	}

	fmt.Printf("Finished processing orders from `%s` status\n", status)
	return nil
}

func getOrders(status []string) ([]*models.Order, error) {
	statusParam := strings.Join(status, ",")

	var orders []*models.Order
	api := fmt.Sprintf("%s/orders?per_page=1&status=%s&nocache=%d", wpBaseAddress, statusParam, time.Now().UnixNano())
	req, err := http.NewRequest(http.MethodGet, api, nil)
	if err != nil {
		log.Fatal("failed to create reqeust")
		return nil, err
	}

	auth := base64.StdEncoding.EncodeToString([]byte(wpApiKey + ":" + wpApiSecret))
	req.Header.Add("Authorization", "Basic "+auth)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error making request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response: %v", err)
		return nil, err
	}

	var responsePayload []*models.Order
	err = json.Unmarshal(body, &responsePayload)
	if err != nil {
		return nil, err
	}
	orders = append(orders, responsePayload...)

	return orders, nil
}

func getAllOrders(status []string, perPage int) ([]*models.Order, error) {
	pageNumber := 1
	statusParam := strings.Join(status, ",")

	var orders []*models.Order
	for {
		api := fmt.Sprintf("%s/orders?per_page=%d&status=%s&page=%d&nocache=%d", wpBaseAddress, perPage, statusParam, pageNumber, time.Now().UnixNano())
		req, err := http.NewRequest(http.MethodGet, api, nil)
		if err != nil {
			log.Fatal("failed to create reqeust")
			return nil, err
		}

		auth := base64.StdEncoding.EncodeToString([]byte(wpApiKey + ":" + wpApiSecret))
		req.Header.Add("Authorization", "Basic "+auth)
		req.Header.Add("Content-Type", "application/json")

		client := &http.Client{}

		resp, err := client.Do(req)
		if err != nil {
			log.Fatalf("Error making request: %v", err)
			return nil, err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Error reading response: %v", err)
			return nil, err
		}

		var responsePayload []*models.Order
		err = json.Unmarshal(body, &responsePayload)
		orders = append(orders, responsePayload...)
		if len(responsePayload) == 0 {
			return orders, nil
		}
		pageNumber++
	}
}

func checkAndUpdateOrders(orders []*models.Order, destinationStatuses map[string]string) error {

	steadFastOrderApi := steadFAstApiBaseAddress + "/status_by_cid/"

	for _, order := range orders {

		wg.Add(1)
		go func(order *models.Order) {
			defer wg.Done()

			for _, metadata := range order.MetaData {
				if metadata.Key == "steadfast_consignment_id" {

					api := steadFastOrderApi + metadata.Value.(string)
					req, err := http.NewRequest(http.MethodGet, api, nil)
					if err != nil {
						log.Fatal("failed to create reqeust")
					}

					req.Header.Add("Api-Key", steadFastApiKey)
					req.Header.Add("Secret-Key", steadFastApiSecret)
					req.Header.Add("Content-Type", "application/json")

					client := &http.Client{}

					resp, err := client.Do(req)
					if err != nil {
						log.Fatalf("Error making request: %v", err)
					}
					defer resp.Body.Close()
					// Read response body
					body, err := io.ReadAll(resp.Body)
					if err != nil {
						log.Fatalf("Error reading response: %v", err)
					}

					var responsePayload models.OrderStatus
					err = json.Unmarshal(body, &responsePayload)
					if err != nil || responsePayload.Status != 200 {
						log.Fatal("Request failed")
					}

					destinationStatus, ok := destinationStatuses[responsePayload.DeliveryStatus]

					if !ok {
						continue
					}
					// switch responsePayload.DeliveryStatus {
					// case "delivered":
					// 	destinationStatus = "delivered"
					// case "partial_delivered":
					// 	destinationStatus = "partially-deliver"
					// case "cancelled":
					// 	destinationStatus = "cancelled"
					// case "hold":
					// 	destinationStatus = "hold"
					// case "in_review":
					// 	destinationStatus = "in-review"
					// case "unknown":
					// 	destinationStatus = "manual-verify"
					// }

					// if destinationStatus == "" {
					// 	continue
					// }

					if order.Status == destinationStatus {
						continue
					}

					err = changeWpStatus(order.Id, destinationStatus)
					if err == nil {
						fmt.Printf("%s\t\t%s\t\t%s\n", strconv.Itoa(order.Id), order.Status, destinationStatus)
					}

				}
			}
		}(order)

	}
	wg.Wait()
	return nil
}

func changeWpStatus(orderId int, status string) error {
	api := fmt.Sprintf("%s/orders/%s", wpBaseAddress, strconv.Itoa(orderId))

	payload := &models.ChangeStatus{
		Status: status,
	}

	payloadJson, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("unable to parse json", err)
	}

	req, err := http.NewRequest(http.MethodPost, api, bytes.NewBuffer(payloadJson))
	if err != nil {
		log.Fatal("failed to create reqeust")
	}

	auth := base64.StdEncoding.EncodeToString([]byte(wpApiKey + ":" + wpApiSecret))
	req.Header.Add("Authorization", "Basic "+auth)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error making request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("Unable to update status for order #%d", orderId)
		return errors.New("unable to update status")
	}

	return nil
}

func getBalance() (*models.Balance, error) {
	api := steadFAstApiBaseAddress + "/get_balance"
	req, err := http.NewRequest(http.MethodGet, api, nil)
	if err != nil {
		log.Fatal("failed to create reqeust")
	}

	req.Header.Add("Api-Key", steadFastApiKey)
	req.Header.Add("Secret-Key", steadFastApiSecret)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error making request: %v", err)
	}
	defer resp.Body.Close()
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response: %v", err)
	}

	var responsePayload models.Balance
	err = json.Unmarshal(body, &responsePayload)
	if err != nil || responsePayload.Status != 200 {
		return nil, errors.New("please verify steadfast configuration")
	}

	return &responsePayload, nil
}
