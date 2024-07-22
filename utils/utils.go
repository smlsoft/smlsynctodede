package utils

import (
	"fmt"
	"strconv"
	"time"
)

func GetIntInput(prompt string) int {
	for {
		input := getUserInput(prompt)
		num, err := strconv.Atoi(input)
		if err == nil {
			return num
		}
		fmt.Println("Invalid input. Please enter a valid number.")
	}
}

func getUserInput(prompt string) string {
	fmt.Print(prompt)
	var input string
	fmt.Scanln(&input)
	return input
}

func StartLoadingAnimation(message string, stopChan chan bool, separator string) {
	loadingChars := []rune{'◐', '◓', '◑', '◒'}
	i := 0
	for {
		select {
		case <-stopChan:
			fmt.Print("\r   \r") // Clear the line after stopping the animation
			return
		default:
			fmt.Printf("\r%s %c %s", message, loadingChars[i], separator)
			i = (i + 1) % len(loadingChars)
			time.Sleep(100 * time.Millisecond)
		}
	}
}
