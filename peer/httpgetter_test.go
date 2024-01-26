package peer

import (
	"fmt"
	"testing"
)

func TestNewHttpGetter(t *testing.T) {
	getter := NewHttpGetter("http://localhost:8080/jie_cache")
	val, err := getter.Get("school", "Jack")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(val))
}
