package main

import (
    "os"
    "bytes"
    "image"
    "image/color"
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

  http.HandleFunc("/image", httpImage)
  http.HandleFunc("/local", localImage)
  http.HandleFunc("/random", random)
  http.HandleFunc("/gen", gen)
  http.ListenAndServe(":" + port, nil)
}

func localImage(w http.ResponseWriter, r *http.Request) {
  var (
    filename string
    format string
    bits int
    img image.Image
    dat []byte
    err error
  )

  // Read the img GET parameter
  if filename = r.URL.Query().Get("img"); filename == "" {
    print("GET parameter img malformed")
    panic("failed")
  }

  // Get the image type requested (BMP, RAW)
  if format = r.URL.Query().Get("format"); format == "" {
    // default: bmp
    format = "bmp"
  }


  if dat, err = ioutil.ReadFile(filename); err != nil {
    print("Can't open file" + filename)
    panic("failed")
  }


  // Create Image object
  if img, _, err = image.Decode(bytes.NewBuffer(dat)); err != nil {
    print("unable to decode image")
    panic(err)
  }

  // Get the image type requested (BMP, RAW)
  if bits, err = strconv.Atoi(r.URL.Query().Get("bits")); bits == 0 && err != nil {
    // default: 18
    bits = 18
  }

  img = processImage(img, format == "bmp")
  serveImage(w, img, format == "bmp", bits)
}

func httpImage(w http.ResponseWriter, r *http.Request) {
  var (
    img string
    format string
    bits int
    err error
  )

  // Read the img GET parameter
  if img = r.URL.Query().Get("img"); img == "" {
    print("GET parameter img malformed")
    panic("failed")
  }

  // Get the image type requested (BMP, RAW)
  if format = r.URL.Query().Get("format"); format == "" {
    // default: bmp
    format = "bmp"
  }

  // Get the image type requested (BMP, RAW)
  if bits, err = strconv.Atoi(r.URL.Query().Get("bits")); bits == 0 && err != nil {
    // default: 18
    bits = 18
  }

  image := getImageFromURL(w, img)
  image = processImage(image, format == "bmp")
  serveImage(w, image, format == "bmp", bits)
}

func random(w http.ResponseWriter, r *http.Request) {
  print("serving random image")

  var (
    format string
    bits int
    err error
  )

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
      print("can't create request")
      panic(err)
    }

    // Include auth client-id in headers
    req.Header.Add("Authorization", "Client-ID 4299fbb6e4a5cc4")

    // Execute request and get response
    if resp, err = client.Do(req); err != nil {
      print("request failed")
      panic(err)
    }

    // Read content from response
    if content, err = ioutil.ReadAll(resp.Body); err != nil {
      print("reading response failed")
      panic(err)
    }

    // Unmarshal JSON to Struct
    if err = json.Unmarshal(content, &imgur); err != nil {
      print("json unmarshal failed")
      panic(err)
    }
  }

  // Get the image at the current index, and increment index
  data := imgur.Data[count]
  count = count + 1


  // Get the image type requested (BMP, RAW)
  if format = r.URL.Query().Get("format"); format == "" {
    // default: bmp
    format = "bmp"
  }

  // Get the image type requested (BMP, RAW)
  if bits, err = strconv.Atoi(r.URL.Query().Get("bits")); bits == 0 && err != nil {
    // default: 18
    bits = 18
  }

  image := getImageFromURL(w, data.Link)
  image = processImage(image, format == "bmp")
  serveImage(w, image, format == "bmp", bits)
}

func gen(w http.ResponseWriter, r *http.Request) {
  format := "raw"


  img := image.NewRGBA(image.Rect(0, 0, 240, 320))
  bounds := img.Bounds()


  for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
    for x := bounds.Min.X; x < bounds.Max.X; x++ {
      g := uint8((x * 254) / 240)
      img.Set(x, y, color.RGBA{g, 0x00, 0x00, 0xff})
    }
  }

  bits := 18




  serveImage(w, img, format == "bmp", bits)
}





func getImageFromURL(w http.ResponseWriter, img string) image.Image {
  var (
    err error
    content []byte
    input image.Image
    resp *http.Response
  )

  // Request Image
  if resp, err = http.Get(img); err != nil {
    print("requesting image failed")
    panic(err)
  }

  // Read content to byte[]
  defer resp.Body.Close()
  if content, err = ioutil.ReadAll(resp.Body); err != nil {
    print("unable to read content from request")
    panic(err)
  }

  // Create Image object
  if input, _, err = image.Decode(bytes.NewBuffer(content)); err != nil {
    print("unable to decode image")
    panic(err)
  }

  return input
}



func processImage(input image.Image, flipImage bool) image.Image {
  dst := input

  // Resize if needed
  if dst.Bounds().Dx() > 240 || dst.Bounds().Dy() > 320 {
    dst = imaging.Fill(dst, 240, 320, imaging.Center, imaging.Lanczos)
  }

  if flipImage {
    //dst = imaging.FlipH(dst)
    dst = imaging.FlipV(dst)
  }

  return dst
}



func serveImage(w http.ResponseWriter, img image.Image, asBmp bool, bits int) {
  // Create buffer from image
  buffer := new(bytes.Buffer)
  var mimeType string

  if asBmp {
    mimeType = "image/bmp"
    if err := imaging.Encode(buffer, img, imaging.BMP); err != nil {
      print("unable to encode image")
      panic(err)
    }
  } else {
    var mb uint16
    bounds := img.Bounds()
    //mimeType = "application/octet-stream"
    mimeType = "text/plain"

    // two first bytes are for the width
    binary.Write(buffer, binary.BigEndian, uint16(bounds.Max.X))

    // next two bytes for the height
    binary.Write(buffer, binary.BigEndian, uint16(bounds.Max.Y))

    // color bits (18bit 262k, 16bit 65k, 8bit 180)
    binary.Write(buffer, binary.BigEndian, uint16(bits))


    for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
      for x := bounds.Min.X; x < bounds.Max.X; x++ {
        r, g, b, _ := img.At(x, y).RGBA()

        if bits == 16 {
          mb = (uint16(r) & 0xF800) | ((uint16(g) & 0xFC00) >> 5) | (uint16(b) >> 11)
          binary.Write(buffer, binary.BigEndian, mb)
        } else if bits == 18 {
          binary.Write(buffer, binary.BigEndian, uint8(r))
          binary.Write(buffer, binary.BigEndian, uint8(g))
          binary.Write(buffer, binary.BigEndian, uint8(b))
        } else {
          print("bitrate not supported yet: " + strconv.FormatInt(int64(bits), 16))
          panic("here")
        }

      }
    }

    plain := base64.StdEncoding.EncodeToString(buffer.Bytes())
    buffer = bytes.NewBuffer([]byte(plain))
  }



  // Serve data
  w.Header().Set("Content-Type", mimeType)
  w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))


  if _, err := w.Write(buffer.Bytes()); err != nil {
    print("unable to write image")
    panic(err)
  }
}
