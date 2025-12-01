package WebUI

type User struct {
	UserId            string `json:"userId"`
	TenantId          string `json:"tenantId"`
	Email             string `json:"email"`
	EncryptedPassword string `json:"encryptedPassword"`
}
