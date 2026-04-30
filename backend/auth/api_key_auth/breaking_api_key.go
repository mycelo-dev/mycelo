package api_key_auth

import "strings"

func GetTenantPublicIdFromApiKey(api_key string) string {

	return strings.Split(api_key, "_")[1]
}

func GetTeamPublicIdFromApiKey(api_key string) string {

	return strings.Split(api_key, "_")[2]
}

func GetHashStringFromApiKey(api_key string) string {

	return strings.Split(api_key, "_")[1]
}
