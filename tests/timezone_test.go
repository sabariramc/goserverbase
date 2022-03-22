package tests

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestTimeZone(t *testing.T) {
	fmt.Println(os.Environ())
	fmt.Println(os.Getenv("TZ"))
	fmt.Println(time.Now().UTC())
	fmt.Println(time.Now())
}
