package outbound

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/mycelo-dev/mycelo/backend/core"
	"github.com/mycelo-dev/mycelo/backend/queries/select_queries"
	"github.com/mycelo-dev/mycelo/backend/queries/update_queries"
)

type DestinationTopicMapping struct {
	DestinationID        string
	TopicID              string
	LastDeliveredEventID int64
}

type OutboundMappingState struct {
	TopicName            string
	Endpoint             string
	DeliveryFlag         bool
	LastDeliveredEventID int64
	MappingExists        bool
}

func GetDestinationTopicMappings(ctx context.Context) ([]DestinationTopicMapping, error) {
	query := select_queries.GetOutboundMappingsQuery()

	rows, err := core.Get().Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	mappings := make([]DestinationTopicMapping, 0)

	for rows.Next() {
		var mapping DestinationTopicMapping

		if err := rows.Scan(&mapping.DestinationID, &mapping.TopicID, &mapping.LastDeliveredEventID); err != nil {
			return nil, err
		}

		mappings = append(mappings, mapping)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return mappings, nil
}

func GetOutboundMappingState(ctx context.Context, destinationID string, topicID string) (OutboundMappingState, error) {
	query := select_queries.GetOutboundMappingStateQuery()

	row := core.Get().QueryRow(ctx, query, destinationID, topicID)

	var state OutboundMappingState

	err := row.Scan(&state.TopicName, &state.Endpoint, &state.DeliveryFlag, &state.LastDeliveredEventID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return OutboundMappingState{MappingExists: false}, nil
		}

		return OutboundMappingState{}, err
	}

	state.MappingExists = true

	return state, nil
}

func UpdateOutboundMappingCursor(ctx context.Context, destinationID string, topicID string, lastDeliveredEventID int64) error {
	query := update_queries.GetUpdateDestinationTopicMappingCursorQuery()

	_, err := core.Get().Exec(ctx, query, destinationID, topicID, lastDeliveredEventID)
	return err
}
