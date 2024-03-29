// Package qrand provides true random numbers generated from the ANU Quantum Random Numbers Server, https://qrng.anu.edu.au, to which you must have connectivity for true randomness.
// Randomness from the quantum beyond!!! Fallback to Go's crypto/rand package in the event of no connectivity, but also return a PseudoRandomError.
package qrand

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var webSite = "https://qrng.anu.edu.au/API/jsonI.php"

//The way their site's api works...
var ILength = 10    //Number of "packages" to receive
var ISize = 2       //Number of "items" in those packages
var iType = "hex16" //Type of those "items". Not public as changing will definitely break.

// Attempts is the number of times to retry() the GET request if an error occurs.
var Attempts int = 2

// SleepTime is the time to wait between retry() attempts.
var SleepTime time.Duration = time.Second * 1

// PseudoRandomError is the error type returned if no complete interaction with the WebSite occurs and a pseudo-random []byte is returned instead.
// Check for it with "if _, ok := x.(qrand.PseudoRandomError); ok {..."
type PseudoRandomError struct{}

// Reader is a drop-in, true random number generator replacement for crypto/rand's Reader.
type Reader struct{}

func (f PseudoRandomError) Error() string {
	return fmt.Sprintf("No connectivity to %v. Generating pseudo-random number instead.", webSite)
}

func (r *Reader) Read(p []byte) (n int, err error) {

	l := len(p)

	if l == 0 {
		return 0, nil
	}

	x, err := Get(l)

	for i, b := range x {
		p[i] = b
	}

	return len(x), err
}

// func Get returns a quantum random []byte of size and a nil error, or a pseudo-random []byte of size and an error of type PseudoRandomError, or nil and a regular, old error.
func Get(size int) (out []byte, err error) {

	if size < 1 {
		return nil, errors.New("size parameter must be a positive integer.")
	}

	out = make([]byte, 0, size)

	endPoint := webSite + fmt.Sprintf("?length=%v&type=%v&size=%v", ILength, iType, ISize)

	var resp *http.Response

	type Data struct {
		RType    string   `json:"type"`
		RLenght  int      `json:"length"`
		RSize    int      `json:"size"`
		RData    []string `json:"data"`
		RSuccess bool     `json:"success"`
	}

	for i := 0; i < size; {

		err := retry(Attempts, SleepTime, func() (err error) {

			resp, err = http.Get(endPoint)
			return
		})
		if err != nil {
			fmt.Println("Error in GET request.", err)
			break //Fall back to pseudo-random
		}

		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Println("Error reading HTTPS response.", err)
			break // Fall back to pseudo-random
		}

		var jData Data

		jData.RData = make([]string, ILength)

		err = json.Unmarshal(body, &jData)
		if err != nil {
			fmt.Println("Error unmarshaling HTTPS response JSON.", err)
			break // Fall back to pseudo-random
		}

		if jData.RSuccess == false {
			fmt.Println("Error in data returned from: ", webSite)
			break // Fall back to pseudo-random
		}

		for j := 0; j < ILength; j++ {

			jDataBytes, err := hex.DecodeString(jData.RData[j])
			if err != nil {
				fmt.Println("Error decoding HTTPS response hex data.", err)
				break // Fall back to pseudo-random
			}

			out = append(out, jDataBytes...)
			i = i + ISize
		}
		// Propogate any error from loop the loop above.
		if err != nil {
			break
		}
	}

	if len(out) == size {
		return out, nil
	}

	if len(out) >= size {
		return out[:size], nil // Truncate to remove excess data.
	}

	// Falling back to math/rand

	fmt.Println("Falling back to pseudo-random generation...")

	out = out[0:size]

	n, err := rand.Read(out)
	if err != nil || n != size {
		fmt.Println("Something went wrong with generating the pseudo-random.", err)
		if err != nil {
			return nil, err
		}
	}

	return out, PseudoRandomError{}
}

// Credit to Alexandre Bourget...
func retry(attempts int, sleep time.Duration, callback func() error) (err error) {
	for i := 0; ; i++ {
		err = callback()
		if err == nil {
			return
		}

		if i >= (attempts - 1) {
			break
		}

		time.Sleep(sleep)

		fmt.Println("retrying after error:", err)
	}
	return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}
