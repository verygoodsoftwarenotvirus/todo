/*
Command two_factor is a CLI that takes in a secret as a positional argument
and draws the TOTP code for that secret in big ASCII numbers. This command is
helpful when you need to repeatedly test the logic of registering an account
and logging in.
*/
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
	fmt.Println("\x1b[2J")
	fmt.Printf("\x1b[0;0H")
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

	timeLeft := (30*time.Second - time.Since(lastChange).Round(time.Second)).String()
	out += fmt.Sprintf("\n\n%s\n", timeLeft)

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
		panic("this shouldn't happen")
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
