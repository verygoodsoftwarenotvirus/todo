package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pquerna/otp/totp"
)

const (
	zero  = "  ___   & / _ \\  &| | | | &| |_| | & \\___/  "
	one   = "    _    &  /_ |   &   | |   &  _| |_  & |_____| "
	two   = " ____   &|___ \\  &  __) | & / __/  &|_____| "
	three = "_____   &|___ /  &  |_ \\  & ___) | &|____/  "
	four  = " _   _   &| | | |  &| |_| |_ &|___   _ &    |_|  "
	five  = " ____   &| ___|  &|___ \\  & ___) | &|____/  "
	six   = "  __    & / /_   &| '_ \\  &| (_) | & \\___/  "
	seven = " _____  &|___  | &   / /  &  / /   & /_/    "
	eight = "  ___   & ( o )  & /   \\  &|  O  | & \\___/  "
	nine  = "  ___   & /   \\  &| (_) | & \\__, | &   /_/  "
)

var (
	lastChange  time.Time
	currentCode string
	// feel free to link to this variable and the related  non-stdlib
	// functions as a demonstration of useless overengineering
	numbers = [10][5]string{
		limitSlice(strings.Split(zero, "&")),
		limitSlice(strings.Split(one, "&")),
		limitSlice(strings.Split(two, "&")),
		limitSlice(strings.Split(three, "&")),
		limitSlice(strings.Split(four, "&")),
		limitSlice(strings.Split(five, "&")),
		limitSlice(strings.Split(six, "&")),
		limitSlice(strings.Split(seven, "&")),
		limitSlice(strings.Split(eight, "&")),
		limitSlice(strings.Split(nine, "&")),
	}
)

func limitSlice(in []string) (out [5]string) {
	if len(in) != 5 {
		panic("wut")
	}
	for i := 0; i < 5; i++ {
		out[i] = in[i]
	}
	return
}

func mustnt(err error) {
	if err != nil {
		panic(err)
	}
}

func clearTheScreen() {
	fmt.Println("\033[2J")
	fmt.Printf("\033[0;0H")
}

func buildTheThing(token string) string {
	var out string
	for i := 0; i < 5; i++ {
		if i != 0 {
			out += "\n"
		}
		for _, x := range strings.Split(token, "") {
			y, err := strconv.Atoi(x)
			if err != nil {
				panic(err)
			}
			out += "  "
			out += numbers[y][i]
		}
	}

	out += "\n\n" + (30*time.Second - time.Since(lastChange).Round(time.Second)).String() + "\n"

	return out
}

func doTheThing(secret string) {
	t := strings.ToUpper(secret)
	n := time.Now().UTC()
	code, err := totp.GenerateCode(t, n)
	mustnt(err)

	if code != currentCode {
		lastChange = time.Now()
		currentCode = code
	}

	if !totp.Validate(code, t) {
		panic("omg are you serious???????????")
	}

	clearTheScreen()
	fmt.Println(buildTheThing(code))
}

func requestTOTPSecret() string {
	var (
		token string
		err   error
	)

	if len(os.Args) == 1 {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("token: ")
		token, err = reader.ReadString('\n')
		mustnt(err)
	} else {
		token = os.Args[1]
	}

	return token
}

func main() {
	secret := requestTOTPSecret()
	clearTheScreen()
	doTheThing(secret)
	every := time.Tick(1 * time.Second)
	lastChange = time.Now()

	for range every {
		doTheThing(secret)
	}
}
