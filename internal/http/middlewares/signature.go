package middlewares

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"os"
)


func VerifyRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Get the Signature
		signature := r.Header.Get("LT-SIGNATURE")
		if signature == "" {
			http.Error(w,"Missing Header for Verification!",http.StatusBadRequest)
			return
		}
		 
		// Verify with the request
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			return
		}
		// Reset r.Body for downstream handlers
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		secret := os.Getenv("HMAC_SECRET")
		if secret == "" {
			http.Error(w,"Internal Server Error",http.StatusInternalServerError)
			return
		}

		ok := verifyHMAC(bodyBytes,[]byte(secret),signature)
		if !ok {
			http.Error(w,"Invalid Request",http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}


func verifyHMAC(request []byte, secret []byte, signature string) bool {
	// Convert the signature to bytes
	byteSig, err := hex.DecodeString(signature)
	if err !=  nil {
		return false
	}
	
	// Create a HMAC object
	hmacObject := hmac.New(sha256.New, secret)
	hmacObject.Write(request)

	// Equate both request and signature
	computed := hmacObject.Sum(nil)
	res := hmac.Equal(byteSig, computed)

	return res
}