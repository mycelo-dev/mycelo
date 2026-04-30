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

type DeleteDestination struct {
	Id string `json:"id"`
}

type AssignTopicToDestination struct {
	Destination_id string `json:"destination_id"`
	Topic_id       string `json:"topic_id"`
}

type DeleteDestinationTopicMapping struct {
	Destination_id string `json:"destination_id"`
	Topic_id       string `json:"topic_id"`
}
