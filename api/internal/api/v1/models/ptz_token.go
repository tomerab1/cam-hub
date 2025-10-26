package models

// PtzToken represents a PTZ (Pan-Tilt-Zoom) camera token structure.
// It contains a unique identifier and an authentication token for PTZ operations.
//
// Fields:
//   - UUID: A unique identifier string for the PTZ token
//   - Token: The authentication token string used for PTZ camera operations
type PtzToken struct {
	UUID  string
	Token string
}
