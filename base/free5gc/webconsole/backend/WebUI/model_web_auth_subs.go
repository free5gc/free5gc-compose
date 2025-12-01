package WebUI

import (
	"github.com/free5gc/openapi/models"
)

type WebAuthenticationSubscription struct {
	AuthenticationMethod models.AuthMethod `json:"authenticationMethod" bson:"authenticationMethod"`
	PermanentKey         *PermanentKey     `json:"permanentKey" bson:"permanentKey"`
	SequenceNumber       string            `json:"sequenceNumber" bson:"sequenceNumber"`
	Milenage             *Milenage         `json:"milenage,omitempty" bson:"milenage"`
	Opc                  *Opc              `json:"opc,omitempty" bson:"opc"`
	//nolint
	AuthenticationManagementField string `json:"authenticationManagementField,omitempty" bson:"authenticationManagementField"`
}

type PermanentKey struct {
	PermanentKeyValue   string `json:"permanentKeyValue" bson:"permanentKeyValue"`
	EncryptionKey       int32  `json:"encryptionKey" bson:"encryptionKey"`
	EncryptionAlgorithm int32  `json:"encryptionAlgorithm" bson:"encryptionAlgorithm"`
}

type Milenage struct {
	Op *Op `json:"op,omitempty" bson:"op"`
}

type Op struct {
	OpValue             string `json:"opValue" bson:"opValue"`
	EncryptionKey       int32  `json:"encryptionKey" bson:"encryptionKey"`
	EncryptionAlgorithm int32  `json:"encryptionAlgorithm" bson:"encryptionAlgorithm"`
}

type Opc struct {
	OpcValue            string `json:"opcValue" bson:"opcValue"`
	EncryptionKey       int32  `json:"encryptionKey" bson:"encryptionKey"`
	EncryptionAlgorithm int32  `json:"encryptionAlgorithm" bson:"encryptionAlgorithm"`
}
