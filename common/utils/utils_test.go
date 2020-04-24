package utils

import (
	"fmt"
	"testing"
)

func TestSha256(t *testing.T) {
	fmt.Println(Sha256("admin", "superUser"))
}
