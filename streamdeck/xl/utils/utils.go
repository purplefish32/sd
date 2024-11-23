package utils

import (
	"fmt"
	"log"
	"os"

	"github.com/h2non/bimg"
	"github.com/karalabe/hid"
)

const (
	ImageReportLength        = 1024  // The expected length of the HID report
	ImageReportPayloadLength = 1016  // The size of each payload chunk
	ImageReportHeaderLength  = 8     // The size of the header in the report
)

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


// func SetKey(device *hid.Device, keyId int, imagePath string) bool {
// 	buffer, err := bimg.Read(imagePath)
// 	if err != nil {
// 		fmt.Fprintln(os.Stderr, err)
// 	}

// 	image := bimg.NewImage(buffer)

// 	// first crop image
// 	_, err = image.Resize(96, 96)
// 	if err != nil {
// 		fmt.Fprintln(os.Stderr, err)
// 	}

// 	// then flip it
// 	newImage, err := image.Rotate(180)
// 	if err != nil {
// 		fmt.Fprintln(os.Stderr, err)
// 	}

// 	// Calculate the total length of the image data
// 	content := newImage

// 	remainingBytes := len(content)
// 	iteration := 0

// 	// Ensure the device is opened for communication
// 	// device.Open() may not be necessary since you're passing a device object already open, but you can modify it based on your code
// 	if device != nil {
// 		for remainingBytes > 0 {
// 			// Slice the image to fit into the payload size
// 			sliceLength := min(remainingBytes, ImageReportPayloadLength)
// 			bytesSent := iteration * ImageReportPayloadLength

// 			// Determine if this is the final chunk
// 			var finalizer byte
// 			if sliceLength == remainingBytes {
// 				finalizer = 1
// 			} else {
// 				finalizer = 0
// 			}

// 			// Prepare the header with bit manipulation
// 			bitmaskedLength := byte(sliceLength & 0xFF)
// 			shiftedLength := byte(sliceLength >> 8)
// 			bitmaskedIteration := byte(iteration & 0xFF)
// 			shiftedIteration := byte(iteration >> 8)

// 			// Create the header (This can be adjusted based on the actual protocol you are working with)
// 			header := []byte{
// 				0x02,  // Report ID for key setting
// 				0x07,  // Command for setting key image (check your device documentation)
// 				byte(keyId),  // Key ID
// 				finalizer,     // Final chunk indicator
// 				bitmaskedLength,
// 				shiftedLength,
// 				bitmaskedIteration,
// 				shiftedIteration,
// 			}

// 			// Slice the image data
// 			payload := append(header, content[bytesSent:bytesSent+sliceLength]...)
// 			padding := make([]byte, ImageReportLength-len(payload))

// 			// Final payload with padding
// 			finalPayload := append(payload, padding...)

// 			// Write the payload to the Stream Deck
// 			_, err := device.Write(finalPayload)
// 			if err != nil {
// 				log.Printf("Error writing to device: %v", err)
// 				return false
// 			}

// 			remainingBytes -= sliceLength
// 			iteration++
// 		}
// 		return true
// 	}
// 	return false
// }

func SetKeyFromBuffer(device *hid.Device, keyId int, buffer []byte) bool {
	image := bimg.NewImage(buffer)

	// first crop image
	_, err := image.Resize(96, 96)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	// then flip it
	newImage, err := image.Rotate(180)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	// Calculate the total length of the image data
	content := newImage

	remainingBytes := len(content)
	iteration := 0

	// Ensure the device is opened for communication
	// device.Open() may not be necessary since you're passing a device object already open, but you can modify it based on your code
	if device != nil {
		for remainingBytes > 0 {
			// Slice the image to fit into the payload size
			sliceLength := min(remainingBytes, ImageReportPayloadLength)
			bytesSent := iteration * ImageReportPayloadLength

			// Determine if this is the final chunk
			var finalizer byte
			if sliceLength == remainingBytes {
				finalizer = 1
			} else {
				finalizer = 0
			}

			// Prepare the header with bit manipulation
			bitmaskedLength := byte(sliceLength & 0xFF)
			shiftedLength := byte(sliceLength >> 8)
			bitmaskedIteration := byte(iteration & 0xFF)
			shiftedIteration := byte(iteration >> 8)

			// Create the header (This can be adjusted based on the actual protocol you are working with)
			header := []byte{
				0x02,  // Report ID for key setting
				0x07,  // Command for setting key image (check your device documentation)
				byte(keyId),  // Key ID
				finalizer,     // Final chunk indicator
				bitmaskedLength,
				shiftedLength,
				bitmaskedIteration,
				shiftedIteration,
			}

			// Slice the image data
			payload := append(header, content[bytesSent:bytesSent+sliceLength]...)
			padding := make([]byte, ImageReportLength-len(payload))

			// Final payload with padding
			finalPayload := append(payload, padding...)

			// Write the payload to the Stream Deck
			_, err := device.Write(finalPayload)
			if err != nil {
				log.Printf("Error writing to device: %v", err)
				return false
			}

			remainingBytes -= sliceLength
			iteration++
		}
		return true
	}
	return false
}