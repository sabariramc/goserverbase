package utils

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"

	"sabariram.com/goserverbase/errors"
)

func GetenvMust(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(errors.NewCustomError(http.StatusFailedDependency, fmt.Sprintf("mandatory environment variable is not set %v", key), nil))
	}
	return value
}

func Getenv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		value = defaultValue
	}
	return value
}

func ExecuteCommand(command string, arg ...string) {
	cmd := exec.Command(command, arg...)
	stdout, err := cmd.Output()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// Print the output
	fmt.Println(string(stdout))
}

func PrintDir(path string) {

	files, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Printf("Error in searcing dir %v : %v\n", path, err)
		return
	}
	fmt.Printf("Contents of dir %v:\n", path)
	for _, f := range files {
		fmt.Println(f.Name())
	}
}
