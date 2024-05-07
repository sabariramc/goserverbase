package log

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sabariramc/goserverbase/v5/utils"
)

type CustomerIdentifier struct {
	UserID    *string `header:"x-user-id,omitempty" body:"userId,omitempty"`
	AppUserID *string `header:"x-app-user-id,omitempty" body:"appUserID,omitempty"`
	EntityID  *string `header:"x-entity-id,omitempty" body:"entityId,omitempty"`
}

func (c *CustomerIdentifier) GetPayload() map[string]string {
	encodedData, _ := utils.BodyJSON.Marshal(c)
	res := map[string]string{}
	json.Unmarshal(encodedData, &res)
	return res
}

func (c *CustomerIdentifier) GetHeader() map[string]string {
	encodedData, _ := utils.HeaderJSON.Marshal(c)
	res := map[string]string{}
	json.Unmarshal(encodedData, &res)
	return res
}

func (c *CustomerIdentifier) LoadFromHeader(header map[string]string) error {
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

func GetCustomerIdentifier(ctx context.Context) *CustomerIdentifier {
	iVal := ctx.Value(ContextKeyCustomerIdentifier)
	if iVal == nil {
		return &CustomerIdentifier{}
	}
	val, ok := iVal.(*CustomerIdentifier)
	if !ok {
		return &CustomerIdentifier{}
	}
	return val
}
