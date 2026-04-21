package routes

import (
	"fmt"
	"net/http"

	"github.com/mycelo-dev/mycelo/backend/destination_management"
	stream_routes "github.com/mycelo-dev/mycelo/backend/routes/stream"
	topics_routes "github.com/mycelo-dev/mycelo/backend/topics_management"
)

func HandleRequests() {

	// stream routes
	http.HandleFunc("/publish", stream_routes.Publish)
	http.HandleFunc("/events", stream_routes.GetEvents)

	// topics routes
	http.HandleFunc("/create_topic", topics_routes.CreateTopicRoute)
	http.HandleFunc("/update_topic", topics_routes.UpdateTopicRoute)

	// destination routes
	http.HandleFunc("/create_destination", destination_management.CreateDestinationRoute)

	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		fmt.Println("error occurred while starting the web server: ", err)
		return
	}

	fmt.Println("successfully started the web server")
}
