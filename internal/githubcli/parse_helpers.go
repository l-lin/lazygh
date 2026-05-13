package githubcli

import "fmt"

func decodeEndpointJSONResponse(data []byte, target any, endpointErr error) error {
	if err := (responseDecoder{}).DecodeJSON(data, target); err != nil {
		return fmt.Errorf("%w: %v", endpointErr, err)
	}
	return nil
}

func decodeEndpointPaginatedOrFlatJSONResponse(data []byte, target any, endpointErr error) error {
	if err := (paginator{}).DecodeSlurpedJSON(data, target); err == nil {
		return nil
	}
	if err := (responseDecoder{}).DecodeJSON(data, target); err != nil {
		return fmt.Errorf("%w: %v", endpointErr, err)
	}
	return nil
}

func decodeEndpointGraphQLResponse(data []byte, target any, endpointErr error) error {
	if err := (responseDecoder{}).DecodeGraphQL(data, target); err != nil {
		return fmt.Errorf("%w: %v", endpointErr, err)
	}
	return nil
}
