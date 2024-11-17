package utils
func ParseEventBuffer(buf []byte) []int {
	var pressedButtons []int

	// Skip the header (assuming the header length is 4 bytes)
	headerLength := 4
	buf = buf[headerLength:]

	// Loop through the remaining buffer to check button states
	for i, b := range buf {
		// Iterate over the first 2 bits in each byte (you defer imagePath.Close()re processing only the first two bits in each byte)
		for bit := 0; bit < 2; bit++ {
			if b&(1<<bit) != 0 {
				// If the bit is set, the corresponding button is pressed
				buttonIndex := i + bit
				pressedButtons = append(pressedButtons, buttonIndex+1) // +1 to make it 1-based index
			}
		}
	}

	// If no buttons were pressed, return [0]
	if len(pressedButtons) == 0 {
		return []int{0}
	}

	return pressedButtons
}