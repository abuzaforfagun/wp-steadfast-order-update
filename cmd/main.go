package main

import (
	"log"

	"github.com/spf13/cobra"
)

func main() {
	var wpSfO = &cobra.Command{
		Use:   "wpsfo",
		Short: "WordPress Steadfast order Status update",
		Run:   func(cmd *cobra.Command, args []string) {},
	}
	wpSfO.AddCommand(UpdateOrderCmd())

	if err := wpSfO.Execute(); err != nil {
		log.Println(err)
	}

}
