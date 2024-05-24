package correlation

import (
	"encoding/json"
	"fmt"

	"github.com/sabariramc/goserverbase/v6/utils"
)

// UserIdentifier defines a context object for customer identification.
type UserIdentifier struct {
	UserID    *string `header:"x-user-id,omitempty" body:"userId,omitempty"`
	AppUserID *string `header:"x-app-user-id,omitempty" body:"appUserID,omitempty"`
	EntityID  *string `header:"x-entity-id,omitempty" body:"entityId,omitempty"`
}

// GetPayload encodes UserIdentifier into a map[string]string with body struct tags.
func (c *UserIdentifier) GetPayload() map[string]string {
	encodedData, _ := utils.BodyJSON.Marshal(c)
	res := map[string]string{}
	json.Unmarshal(encodedData, &res)
	return res
}

// GetHeader encodes UserIdentifier into a map[string]string with header struct tags.
func (c *UserIdentifier) GetHeader() map[string]string {
	encodedData, _ := utils.HeaderJSON.Marshal(c)
	res := map[string]string{}
	json.Unmarshal(encodedData, &res)
	return res
}

// LoadFromHeader extracts UserIdentifier from a map[string]string with header struct tags.
func (c *UserIdentifier) LoadFromHeader(header map[string]string) error {
	data, err := json.Marshal(header)
	if err != nil {
		return fmt.Errorf("UserIdentifier.LoadFromHeader: error marshalling header: %w", err)
	}
	err = utils.HeaderJSON.Unmarshal(data, c)
	if err != nil {
		return fmt.Errorf("UserIdentifier.LoadFromHeader: error unmarshalling header: %w", err)
	}
	return nil
}
