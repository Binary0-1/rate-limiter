package apistore

func GetApiKeys() map[string]bool {
	KEYS := map[string]bool{
		"apikey123": true,
		"apikey124": true,
	}

	return KEYS
}