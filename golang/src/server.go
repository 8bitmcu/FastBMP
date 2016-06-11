package main

import (
    "os"
    "time"
    "bytes"
    "image"
    "strconv"
    "net/http"
    "io/ioutil"
    "encoding/json"
    "encoding/base64"
    "encoding/binary"
    "github.com/disintegration/imaging"
)

type Imgur struct {
  Data[] struct {
    Id string `json:"id"`
    Title string `json:"title"`
    Type string `json:"type"`
    Animated bool `json:"animated"`
    Link string `json:"link"`
  } `json:"data"`

}


var imgur Imgur
var count = -1


func main() {
  // first argument is the port of the server, optional
  port := "8080"

  if len(os.Args) > 1 {
    port = os.Args[1]
  }

  println(time.Now().Format("15:04:05 Mon January _2 2006") + ". Server up and running on port " + port)

  http.HandleFunc("/bitmap", httpImage)
  http.ListenAndServe(":" + port, nil)
}

func httpImage(w http.ResponseWriter, r *http.Request) {
  var (
    url string
    source string
    bits int
    err error
    img image.Image
  )

  // Read the source GET parameter (local, http, imgur)
  if source = r.URL.Query().Get("source"); source == "" {
    println("GET parameter source malformed")
    panic("failed")
  }

  // Read the url GET parameter (used in local, http sources)
  if url = r.URL.Query().Get("url"); url == "" && source != "imgur" {
    println("GET parameter url malformed")
    panic("failed")
  }

  // Get the requested bitrate, defaults to 16
  if bits, err = strconv.Atoi(r.URL.Query().Get("bits")); !(bits == 16 || bits == 18) || err != nil {
    bits = 16
  }

  // Get image
  if source == "local" {
    img = getLocalImage(url)
  } else if source == "http" {
    img = getImageFromURL(url)
  } else if source == "imgur" {
    img = getImageFromImgur()
  } else {
    println("source '" + source + "' not supported")
    panic("failed")
  }

  // Process image (crop)
  if img.Bounds().Dx() > 240 || img.Bounds().Dy() > 320 {
    img = imaging.Fill(img, 240, 320, imaging.Center, imaging.Lanczos)
  }

  // Log request to console
  log := time.Now().Format("15:04:05") + ": Requested " + source
  if source != "imgur" {
    log += ": " + url
  }
  println(log + ", bits: " + strconv.Itoa(bits))

  // Serve image
  serveImage(w, img, bits)
}



func getLocalImage(filename string) image.Image {
  var (
    img image.Image
    dat []byte
    err error
  )

  // Read file from disk
  if dat, err = ioutil.ReadFile(filename); err != nil {
    println("Can't open file" + filename)
    panic(err)
  }

  // Create Image object
  if img, _, err = image.Decode(bytes.NewBuffer(dat)); err != nil {
    println("unable to decode image")
    panic(err)
  }

  return img
}


func getImageFromImgur() image.Image {

  //if the index overflows or isn't set yet
  if count >= len(imgur.Data) || count == -1 {
    var (
      err error
      req *http.Request
      resp *http.Response
      content []byte
    )

    count = 0
    client := &http.Client{}

    // Create imgur.com api request
    if req, err = http.NewRequest("GET", "https://api.imgur.com/3/gallery/r/subaru", nil); err != nil {
      println("can't create request")
      panic(err)
    }

    // Include auth client-id in headers
    req.Header.Add("Authorization", "Client-ID 4299fbb6e4a5cc4")

    // Execute request and get response
    if resp, err = client.Do(req); err != nil {
      println("request failed")
      panic(err)
    }

    // Read content from response
    if content, err = ioutil.ReadAll(resp.Body); err != nil {
      println("reading response failed")
      panic(err)
    }

    // Unmarshal JSON to Struct
    if err = json.Unmarshal(content, &imgur); err != nil {
      println("json unmarshal failed")
      panic(err)
    }
  }

  // Get the image at the current index, and increment index
  data := imgur.Data[count]
  count = count + 1

  return getImageFromURL(data.Link)
}


func getImageFromURL(img string) image.Image {
  var (
    err error
    content []byte
    input image.Image
    resp *http.Response
  )

  // Request Image
  if resp, err = http.Get(img); err != nil {
    println("requesting image failed")
    panic(err)
  }

  // Read content to byte[]
  defer resp.Body.Close()
  if content, err = ioutil.ReadAll(resp.Body); err != nil {
    println("unable to read content from request")
    panic(err)
  }

  // Create Image object
  if input, _, err = image.Decode(bytes.NewBuffer(content)); err != nil {
    println("unable to decode image")
    panic(err)
  }

  return input
}


func serveImage(w http.ResponseWriter, img image.Image, bits int) {
  // Create buffer from image
  buffer := new(bytes.Buffer)
  bounds := img.Bounds()

  // two first bytes are for the width
  binary.Write(buffer, binary.BigEndian, uint16(bounds.Max.X))

  // next two bytes for the height
  binary.Write(buffer, binary.BigEndian, uint16(bounds.Max.Y))

  // color bits (18bit 262k, 16bit 65k)
  binary.Write(buffer, binary.BigEndian, uint16(bits))

  // loop through each pixel and convert them to rgb565 or rgb666
  for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
    for x := bounds.Min.X; x < bounds.Max.X; x++ {
      r, g, b, _ := img.At(x, y).RGBA()

      if bits == 16 {
        binary.Write(buffer, binary.BigEndian, (uint16(r) & 0xF800) | ((uint16(g) & 0xFC00) >> 5) | (uint16(b) >> 11))
      } else if bits == 18 {
        binary.Write(buffer, binary.BigEndian, uint8(r))
        binary.Write(buffer, binary.BigEndian, uint8(g))
        binary.Write(buffer, binary.BigEndian, uint8(b))
      }

    }
  }

  // Encode the whole buffers content to base64
  buffer = bytes.NewBuffer([]byte(base64.StdEncoding.EncodeToString(buffer.Bytes())))

  // Write response headers and data
  w.Header().Set("Content-Type", "text/plain")
  w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))

  if _, err := w.Write(buffer.Bytes()); err != nil {
    println("unable to write image")
    panic(err)
  }
}
