package sample

import (
	"time"
)

type Address struct {
	AddressLine1 string `bson:"addressLine1"`
	AddressLine2 string `bson:"addressLine2"`
	AddressLine3 string `bson:"addressLine3"`
	State        string `bson:"state"`
	PIN          string `bson:"pin"`
	Country      string `bson:"country"`
}

type Name struct {
	First  string `bson:"first"`
	Middle string `bson:"middle"`
	Last   string `bson:"last"`
	Full   string `bson:"full"`
}

type PIITestVal struct {
	DOB     time.Time `bson:"dob"`
	Name    Name      `bson:"name"`
	Pan     string    `bson:"pan"`
	Email   string    `bson:"email"`
	Phone   []string  `bson:"phone"`
	Address Address   `bson:"address"`
	UUID    string    `bson:"UUID"`
}

func GetRandomData(uuid string) PIITestVal {
	dob, _ := time.Parse(time.DateOnly, "2001-01-01")
	return PIITestVal{
		UUID: uuid,
		DOB:  dob,
		Name: Name{
			First:  uuid + " first name",
			Middle: uuid + " middle name",
			Last:   uuid + " last name",
			Full:   uuid + " full name",
		},
		Pan:   "ABCDE1234F",
		Email: "abc@" + uuid + ".com",
		Phone: []string{"9600000000", "9600000001"},
		Address: Address{
			AddressLine1: uuid + "address first line",
			AddressLine2: uuid + "address first line",
			AddressLine3: uuid + "address first line",
			State:        "Delhi",
			PIN:          "100000",
			Country:      "India",
		},
	}
}
