package util

import (
	"io"

	"github.com/h2non/bimg"
	"github.com/karalabe/hid"
	"github.com/rs/zerolog/log"

	"github.com/olekukonko/tablewriter"
)

const (
	ImageReportLength        = 1024 // The expected length of the HID report
	ImageReportPayloadLength = 1016 // The size of each payload chunk
	ImageReportHeaderLength  = 8    // The size of the header in the report
)

// NewTable creates a pre-configured tablewriter instance
func NewTable(output io.Writer) *tablewriter.Table {
	table := tablewriter.NewWriter(output)
	table.SetBorder(false)
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetTablePadding("   ") // Add spacing for readability
	table.SetNoWhiteSpace(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT}) // Left-align all columns
	return table
}

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

func SetKeyFromBuffer(device *hid.Device, keyId int, buffer []byte) (err error) {

	// Calculate the total length of the image data
	content := buffer

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
				0x02,            // Report ID for key setting
				0x07,            // Command for setting key image (check your device documentation)
				byte(keyId - 1), // Key ID
				finalizer,       // Final chunk indicator
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
				return err
			}

			remainingBytes -= sliceLength
			iteration++
		}
		return err
	}
	return nil
}

func SetKeyFromBufferPlus(device *hid.Device, keyId int, buffer []byte) (err error) {
	// Calculate the total length of the image data
	content := buffer

	remainingBytes := len(content)

	log.Debug().Int("remaining bytes", remainingBytes).Msg("toto")

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
				log.Debug().Int("finalizer", 1).Msg("titi")

			} else {
				finalizer = 0
				log.Debug().Int("finalizer", 0).Msg("titi")
			}

			// Prepare the header with bit manipulation
			bitmaskedLength := byte(sliceLength & 0xFF)
			shiftedLength := byte(sliceLength >> 8)
			bitmaskedIteration := byte(iteration & 0xFF)
			shiftedIteration := byte(iteration >> 8)

			// Create the header (This can be adjusted based on the actual protocol you are working with)
			header := []byte{
				0x02,            // Report ID for key setting
				0x07,            // Command for setting key image (check your device documentation)
				byte(keyId - 1), // Key ID
				finalizer,       // Final chunk indicator
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
				return err
			}

			remainingBytes -= sliceLength
			iteration++
		}
		return err
	}
	return nil
}

func ConvertImageToBuffer(imagePath string, size int) (buffer []byte, err error) {

	// Read the image file into a buffer using bimg
	buffer, err = bimg.Read(imagePath)

	if err != nil {
		return nil, err
	}

	// Create an image object
	image := bimg.NewImage(buffer)

	// Resize the image
	resizedImage, err := image.Resize(size, size)

	if err != nil {
		return nil, err
	}

	// Convert to JPEG.
	finalImage, err := bimg.NewImage(resizedImage).Convert(bimg.JPEG)

	if err != nil {
		return nil, err
	}

	return finalImage, nil
}

func ConvertImageToRotatedBuffer(imagePath string, size int) (buffer []byte, err error) {

	// Read the image file into a buffer using bimg
	buffer, err = bimg.Read(imagePath)
	if err != nil {
		return nil, err
	}

	// Create an image object
	image := bimg.NewImage(buffer)

	// Resize the image
	resizedImage, err := image.Resize(size, size)

	if err != nil {
		return nil, err
	}

	// Rotate the image 180 deg.
	rotatedImage, err := bimg.NewImage(resizedImage).Rotate(180)

	if err != nil {
		return nil, err
	}

	// Convert to JPEG.
	finalImage, err := bimg.NewImage(rotatedImage).Convert(bimg.JPEG)

	if err != nil {
		return nil, err
	}

	return finalImage, nil
}
