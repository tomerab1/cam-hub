package models

// CameraCreds represents camera authentication credentials.
// It contains the unique identifier and login information for a camera.
//
// Fields:
//   - UUID: Unique identifier for the camera credentials, mapped to "id" in database
//   - Username: Camera's login username
//   - Password: Camera's login password (not exposed in JSON)
type CameraCreds struct {
	UUID     string `json:"uuid" db:"id"`
	Username string `json:"username" db:"username"`
	Password string `db:"password"`
}
