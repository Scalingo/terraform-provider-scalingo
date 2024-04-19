package scalingo

import (
	"context"

	scalingo "github.com/Scalingo/go-scalingo/v7"
)

func appEnvironment(ctx context.Context, client *scalingo.Client, appID string) (map[string]interface{}, error) {
	variables, err := client.VariablesList(ctx, appID)
	if err != nil {
		return nil, err
	}
	result := make(map[string]interface{}, len(variables))

	for _, variable := range variables {
		result[variable.Name] = variable.Value
	}
	return result, nil
}

func deleteVariablesByName(ctx context.Context, client *scalingo.Client, appID string, names []string) error {
	if len(names) == 0 {
		return nil
	}

	variables, err := client.VariablesList(ctx, appID)
	if err != nil {
		return err
	}

	for _, variable := range variables {
		if Contains(names, variable.Name) {
			err := client.VariableUnset(ctx, appID, variable.ID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
