package models

import "github.com/google/uuid"

type PersonID string

func AccountIDToPersonID(accountID string) PersonID {
	var personNSUUID = uuid.NewSHA1(uuid.UUID{}, []byte("person"))
	personID := uuid.NewSHA1(personNSUUID, []byte(accountID+"-1"))

	return PersonID(personID.String())
}
