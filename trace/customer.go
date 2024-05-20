package trace

import (
	"encoding/json"
	"fmt"

	"github.com/sabariramc/goserverbase/v6/utils"
)

// UserIdentifier defines context object for a customer identification
type UserIdentifier struct {
	UserID    *string `header:"x-user-id,omitempty" body:"userId,omitempty"`
	AppUserID *string `header:"x-app-user-id,omitempty" body:"appUserID,omitempty"`
	EntityID  *string `header:"x-entity-id,omitempty" body:"entityId,omitempty"`
}

// GetPayload encodes UserIdentifier into map[string]string with body struct tag
func (c *UserIdentifier) GetPayload() map[string]string {
	encodedData, _ := utils.BodyJSON.Marshal(c)
	res := map[string]string{}
	json.Unmarshal(encodedData, &res)
	return res
}

// GetHeader encodes UserIdentifier into map[string]string with header struct tag
func (c *UserIdentifier) GetHeader() map[string]string {
	encodedData, _ := utils.HeaderJSON.Marshal(c)
	res := map[string]string{}
	json.Unmarshal(encodedData, &res)
	return res
}

// ExtractFromHeader extracts UserIdentifier from map[string]string with header struct tag
func (c *UserIdentifier) ExtractFromHeader(header map[string]string) error {
	data, err := json.Marshal(header)
	if err != nil {
		return fmt.Errorf("CustomerIdentifier.LoadFromHeader: error marshalling header: %w", err)
	}
	err = utils.HeaderJSON.Unmarshal(data, c)
	if err != nil {
		return fmt.Errorf("CustomerIdentifier.LoadFromHeader: error unmarshalling header: %w", err)
	}
	return nil
}
