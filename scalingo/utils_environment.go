package scalingo

import scalingo "github.com/Scalingo/go-scalingo/v4"

func appEnvironment(client *scalingo.Client, appId string) (map[string]interface{}, error) {
	variables, err := client.VariablesList(appId)
	if err != nil {
		return nil, err
	}
	result := make(map[string]interface{}, len(variables))

	for _, variable := range variables {
		result[variable.Name] = variable.Value
	}
	return result, nil
}

func deleteVariablesByName(client *scalingo.Client, appID string, names []string) error {
	if len(names) == 0 {
		return nil
	}

	variables, err := client.VariablesList(appID)
	if err != nil {
		return err
	}

	for _, variable := range variables {
		if Contains(names, variable.Name) {
			err := client.VariableUnset(appID, variable.ID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
