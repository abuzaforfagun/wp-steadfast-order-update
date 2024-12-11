package main

import (
	"fmt"

	"github.com/abuzaforfagun/wp-steadfast-order-update/pkg"
	"github.com/spf13/cobra"
)

func UpdateOrderCmd() *cobra.Command {
	var wpHostAddress string
	var wpConsumerKey string
	var wpConsumerSecret string

	var statuses []string
	var destinationStatus map[string]string

	var steadFastApiKey string
	var steadFastApiSecret string
	var updateOrderCmd = &cobra.Command{
		Use:   "update-order",
		Short: "Update wordpress order status according to Steadfast status",
		Run: func(cmd *cobra.Command, args []string) {
			err := pkg.VerifyWp(wpHostAddress, wpConsumerKey, wpConsumerSecret, statuses, destinationStatus)
			if err != nil {
				fmt.Println("Please verify Wordpress config")
				return
			}

			err = pkg.VerifySteadfast(steadFastApiKey, steadFastApiSecret, destinationStatus)
			if err != nil {
				fmt.Println("Please verify Steadfast config")
				return
			}

			pkg.ExecuteProcessOrder(wpHostAddress, wpConsumerKey, wpConsumerSecret, steadFastApiKey, steadFastApiSecret, statuses, destinationStatus)
		},
	}

	updateOrderCmd.Flags().StringVarP(&wpHostAddress, "host", "", "", "Wordpress website address")
	updateOrderCmd.Flags().StringVarP(&wpConsumerKey, "consumer-key", "", "", "Wordpress consumer key")
	updateOrderCmd.Flags().StringVarP(&wpConsumerSecret, "consumer-secret", "", "", "Wordpress consumer secret")
	updateOrderCmd.Flags().StringVarP(&steadFastApiKey, "steadfast-key", "", "", "Stead Fast API key")
	updateOrderCmd.Flags().StringVarP(&steadFastApiSecret, "steadfast-secrect", "", "", "Stead Fast API secrect")
	updateOrderCmd.Flags().StringSliceVarP(&statuses, "statuses", "", []string{}, "Order Statuses to check")
	updateOrderCmd.Flags().StringToStringVarP(&destinationStatus, "destinations", "", map[string]string{}, "Steadfast status map to order status map")

	updateOrderCmd.MarkFlagRequired("host")
	updateOrderCmd.MarkFlagRequired("consumer-key")
	updateOrderCmd.MarkFlagRequired("consumer-secret")
	updateOrderCmd.MarkFlagRequired("steadfast-key")
	updateOrderCmd.MarkFlagRequired("steadfast-secrect")
	updateOrderCmd.MarkFlagRequired("statuses")
	updateOrderCmd.MarkFlagRequired("destinations")

	return updateOrderCmd
}
