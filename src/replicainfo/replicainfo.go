// Package replicainfo holds the information that replicas send to each other
// within messages.
package replicainfo

type ReplicaInfo struct {
	// Can't marshal strings.
	Hostname []byte
	Port     int
	Id       int
}

// HostnameToString converts the hostname from a byte array to a string.
func HostnameToString(hostname []byte) string {
	return string(hostname)
}

// HostnameFromString converts the hostname from a string to a byte array.
func HostnameFromString(hostname string) []byte {
	return []byte(hostname)
}
