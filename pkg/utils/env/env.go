package env

import "os"

func Get(key string, def ...string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	if len(def) > 0 {
		return def[0]
	}
	return ""
}
