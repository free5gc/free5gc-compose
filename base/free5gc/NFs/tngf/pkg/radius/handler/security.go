package handler

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"math/big"

	"github.com/free5gc/tngf/pkg/context"
	"github.com/free5gc/tngf/pkg/radius/message"
)

func GenerateRandomUint8() (uint8, error) {
	number := make([]byte, 1)
	_, err := io.ReadFull(rand.Reader, number)
	if err != nil {
		radiusLog.Errorf("Read random failed: %+v", err)
		return 0, errors.New("read failed")
	}
	return number[0], nil
}

func GetMessageAuthenticator(message *message.RadiusMessage) []byte {
	tngfSelf := context.TNGFSelf()

	radiusMessageData := make([]byte, 4)

	radiusMessageData[0] = message.Code
	radiusMessageData[1] = message.PktID
	radiusMessageData = append(radiusMessageData, message.Auth...)

	radiusMessagePayloadData, err := message.Payloads.Encode()
	if err != nil {
		return nil
	}
	radiusMessageData = append(radiusMessageData, radiusMessagePayloadData...)
	binary.BigEndian.PutUint16(radiusMessageData[2:4], uint16(len(radiusMessageData)))

	hmacFun := hmac.New(md5.New, []byte(tngfSelf.RadiusSecret))
	hmacFun.Write(radiusMessageData)
	return hmacFun.Sum(nil)
}

func GenerateSalt() (uint16, error) {
	maxValue := big.NewInt(0xFFFF)
	number, err := rand.Int(rand.Reader, maxValue)
	if err != nil {
		radiusLog.Errorf("Read random failed: %+v", err)
		return 0, errors.New("read failed")
	}
	// Set the most significant bit to (1)
	number.Or(number, big.NewInt(0x8000))
	return uint16(number.Uint64()), nil
}

func EncryptMppeKey(key, secret, authenticator []byte, saltVal uint16) ([]byte, error) {
	padlen := (md5.Size - (len(key)+1)%md5.Size) % md5.Size
	pad := make([]byte, padlen)

	plain := make([]byte, 1)
	plain[0] = uint8(len(key))
	plain = append(plain, key...)
	plain = append(plain, pad...)

	first := true
	result := []byte{}
	salt := make([]byte, 2)
	binary.BigEndian.PutUint16(salt, saltVal)

	for i := 0; i < len(plain); i += md5.Size {
		var block []byte
		if first {
			block = secret
			block = append(block, authenticator...)
			block = append(block, salt...)
			first = false
		} else {
			block = secret
			block = append(block, result[i-md5.Size:i]...)
		}

		b := md5.Sum(block)
		result = append(result, b[:]...)
		for j := 0; j < md5.Size; j++ {
			result[i+j] ^= plain[i+j]
		}
	}

	radiusLog.Debugln("Reslut", hex.Dump(result))
	return result, nil
}
