package main

import (
  "fmt"
  "os"
  "io"

  "image"
  "image/color"
  "image/draw"
  "image/png"

  "github.com/youpy/go-wav"
  "github.com/mjibson/go-dsp/spectral"
)

func main() {

    inputFilePath := os.Args[1]
    outputFilePath := os.Args[2]


    /////////////
    // Read audio data from wav file

    file, _ := os.Open(inputFilePath)
    reader := wav.NewReader(file)

    defer file.Close()

    var audioData []float64

    for {
        samples, err := reader.ReadSamples()
        if err == io.EOF {
            break
        }

        for _, sample := range samples {
            sampleFloat := reader.FloatValue(sample, 0)
            audioData = append(audioData, sampleFloat)
        }

    }

    format, _ := reader.Format()

    // fmt.Println(len(audioData))



    ////////////
    // Do spectral analysis using

    var min, max float64

    const blockSize = 1024
    const imageWidth = 1000

    nSamples := len(audioData)
    scaleFactor := nSamples / imageWidth
    overlap := blockSize - scaleFactor

    inputSegments := spectral.Segment(audioData, blockSize, overlap)
    var outputSegments [][]float64

    spectralOptions := &spectral.PwelchOptions{Scale_off: true, NFFT: blockSize}

    var frequencies []float64

    for _, segment := range inputSegments {
        var spectralData []float64
        spectralData, frequencies = spectral.Pwelch(segment, float64(format.SampleRate), spectralOptions)
        
        min, max = minAndMax(spectralData)

        outputSegments = append(outputSegments, spectralData)         
    }

    // fmt.Println(outputSegments)
    fmt.Println("freqs: ", frequencies)
    fmt.Println("Min: ", min, "Max: ", max)

    fmt.Println("numSegements: ", len(outputSegments))
    fmt.Println("numFreqs: ", len(outputSegments[0]))


    //////////////////
    // Print to a png

    // Create a blank white image
    imageHeight := blockSize / 2
    imgRect := image.Rect(0, 0, imageWidth, imageHeight)
    img := image.NewGray(imgRect)
    draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.ZP, draw.Src)


    for x, segment := range outputSegments {
        for y, value := range segment {

            normalizedValue := normalizeToUint8(min, max, value)
            color := &color.Gray{normalizedValue}
            // fmt.Println(x, y, value, normalizedValue, uint8(normalizedValue), color)

            fill := &image.Uniform{color}

            draw.Draw(img, image.Rect(x, imageHeight - y, x + 1, imageHeight - y + 1), fill, image.ZP, draw.Src)
        
        } 
    }

    out, err := os.Create(outputFilePath)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    err = png.Encode(out, img)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

}


func minAndMax(slice []float64) (min, max float64) {
    min, max = slice[0], slice[0]
    for _, value := range slice {
        if value < min {
            min = value
        }
        if value > max {
            max = value
        }
    }
    return
}

func normalizeToUint8(min, max, value float64) (normalizedValue uint8) {
    distance := max - min
    normalizedValue = uint8(((value / distance) * 255))
    return
}