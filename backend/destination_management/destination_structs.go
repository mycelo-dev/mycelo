package destination_management

type CreateDestination struct {
	Destination_name    string `json:"destination_name"`
	Destination_address string `json:"destination_address"`
}

type UpdateDestination struct {
	Destination_name    string `json:"destination_name"`
	Destination_address string `json:"destination_address"`
	Id                  string `json:"id"`
}

type DeliveryFlag struct {
	Delivery_flag bool
}

type UpdateDeliveryFlag struct {
	Id            string `json:"id"`
	Delivery_flag bool   `json:"delivery_flag"`
}

type DeleteDestination struct {
	Id string `json:"id"`
}

type AssignTopicToDestination struct {
	Destination_id string `json:"destination_id"`
	Topic_id       string `json:"topic_id"`
}

type DestinationRecord struct {
	Destination_id      string `json:"destination_id"`
	Destination_name    string `json:"destination_name"`
	Destination_address string `json:"destination_address"`
	Delivery_flag       bool   `json:"delivery_flag"`
}

type DestinationTopicMappingRecord struct {
	Destination_id          string `json:"destination_id"`
	Destination_name        string `json:"destination_name"`
	Destination_address     string `json:"destination_address"`
	Delivery_flag           bool   `json:"delivery_flag"`
	Last_delivered_event_id int64  `json:"last_delivered_event_id"`
	Topic_id                string `json:"topic_id"`
	Topic_name              string `json:"topic_name"`
}

type DeleteDestinationTopicMapping struct {
	Destination_id string `json:"destination_id"`
	Topic_id       string `json:"topic_id"`
}
