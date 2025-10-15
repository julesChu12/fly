package main

import (
	"log"
	"os"

	"github.com/julesChu12/fly/kratos/cmd/kratos/cmd"
)

// @title Kratos Order Service API
// @version 1.0
// @description Order service for Fly platform
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8082
// @BasePath /

func main() {
	if err := cmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
