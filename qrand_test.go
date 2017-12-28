package qrand

import (
	"fmt"
	"testing"
)

func TestGet(t *testing.T) {

	cases := []int{1, 10, 100, -1, 0, -10}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%v", tc), func(t *testing.T) {

			actual, err := Get(tc)
			if err != nil {
				fmt.Println(err)
			}

			fmt.Println(tc, ":", actual)
		})
	}
}
