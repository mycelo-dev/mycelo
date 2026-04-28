package api_key_auth

func CompareApiKeyHash(incoming_api_key_hash string, stored_api_key_hash string) bool {

	return incoming_api_key_hash == stored_api_key_hash
}
