package static

import "os"

func getEnv(key, fallback string) string {
	if env, ok := os.LookupEnv(key); ok {
		return env
	}
	return fallback
}
