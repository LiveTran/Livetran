package auth

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)



func GenerateStreamKey(streamId string) (token string,err error) {

	claims := jwt.MapClaims{
		"stream_id": streamId,
		"exp": time.Now().Add(time.Hour * 2).Unix(),
	}
	unsigned_token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secret := []byte(os.Getenv("JWT_SECRET"))
	token,err = unsigned_token.SignedString(secret)
	if err != nil {
		return "",err
	}

	return fmt.Sprintf("mode=publish,rid=%s,token=%s",streamId,token),nil 
}


func DecodeStreamKey(streamId string, streamkey string) (ok bool,data string) {

	secret := []byte(os.Getenv("JWT_SECRET"))

	values,err := parseStreamKey(streamkey)
	if err != nil {
		return false,err.Error()
	}

	if values["rid"] != streamId {
		return false,"Invalid StreamId"
	}
	if values["token"] == "" {
		return false,"Invalid Token"
	}
	

	token, err := jwt.Parse(values["token"], func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})
	if err != nil || !token.Valid {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return false, "Token expired"
		}
		return false,"Invalid Token"
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false,"Invalid Claim Structure"
	}

	sid, ok := claims["stream_id"].(string)
	if !ok {
		return false, "stream_id claim is not a string"
	}
	if sid != streamId {
		return false, "stream_id mismatch"
	}
	return true, sid

}


func parseStreamKey(streamkey string) (map[string]string, error) {
	keys := strings.Split(streamkey, ",")
	data := make(map[string]string)

	for _, key := range keys {
		pair := strings.SplitN(key, "=", 2) // Splits first = only Since jwt has =
		if len(pair) != 2 {
			return nil, fmt.Errorf("malformed key=value pair: %s", key)
		}
		data[strings.TrimSpace(pair[0])] = strings.TrimSpace(pair[1])
	}

	return data, nil
}
