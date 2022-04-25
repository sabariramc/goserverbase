package tests

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestTimeZone(t *testing.T) {
	fmt.Println(time.Now().Zone())
	fmt.Println(os.Getenv("TZ"))
	fmt.Println(time.Now().UTC())
	fmt.Println(time.Now())
}
