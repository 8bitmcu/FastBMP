#include <ESP8266WiFi.h>
#include <SPI.h>
#include <math.h>

#include "Adafruit_GFX.h"
#include "Adafruit_ILI9340.h"
#include "ArduinoJson.h"
#include "Base_64.h"

#include "wificonf.h"

#define GO_host     "192.168.0.19"
#define GO_port     8088
#define ILI9340_CS  15
#define ILI9340_DC  2
#define ILI9340_RST 4

WiFiClient client;
Adafruit_ILI9340 tft = Adafruit_ILI9340(ILI9340_CS, ILI9340_DC, ILI9340_RST);

// init LCD and connect to wifi
void setup() {
  tft.begin();

  // Set the orientation to portrait
  tft.setRotation(2);
  tft.setCursor(0, 0);
  
  tft.fillScreen(ILI9340_BLACK);
  tft.setTextColor(ILI9340_WHITE);

  tft.print("Wi-Fi init ");
  tft.print(SSID_NAME);
  WiFi.begin(SSID_NAME, SSID_PASS);

  // Try to connect, and time out after a minute
  unsigned long tout = millis() + 60 * 1000;
  unsigned long ldot = tout;
  while(WiFi.status() != WL_CONNECTED && tout > millis()) {
    if(millis() - ldot > 500) {
      tft.print(".");
      ldot = millis();
    }
    yield();
  }

  // Connected to wifi?
  if(WiFi.status() != WL_CONNECTED) {
    tft.setTextColor(ILI9340_RED);
    tft.println("\nConnection failed!");
    return;
  }
  else {
    tft.setTextColor(ILI9340_WHITE);
    tft.print("\nDHCP assigned ip: ");
    tft.println(WiFi.localIP());
  }

  // Connect to server
  if (!client.connect(GO_host, GO_port)) {
    tft.setTextColor(ILI9340_RED);
    tft.println("Cannot connect to server");
    return;
  }
  else {
    tft.println("Connected to server");
  }
}


void loop(void) {

  //request("/bitmap?source=imgur");
  request("/bitmap?source=http&url=https://pixabay.com/static/uploads/photo/2013/07/12/12/58/tv-test-pattern-146649_960_720.png");
  
  drawBMP(0, 0);

  delay(2000);
  tft.setCursor(0, 0);
  tft.fillScreen(ILI9340_WHITE);
}


void request(char* url) {
  client.print(String("GET ") + url + " HTTP/1.1\r\n" +
               "Host: " + GO_host + "\r\n" +
               "Connection: close\r\n\r\n");

  while (!client.available()) {
    // todo: timeout
    yield();
  }

  // ignore headers (assume 200)
  while (client.available() && client.readStringUntil('\r') != "\n");
}


void read_b64(uint8_t *output, uint16_t outputLen) {
  uint16_t bufSize = (outputLen/3)*4;
  uint16_t idx = 0;
  uint8_t buf[bufSize];
  
  while(idx < bufSize) {
    idx += client.read(&buf[idx], bufSize-idx);
    yield();
  }

  base64_decode((char *) output,(char *) buf, bufSize);
}


void drawBMP(uint16_t x, uint16_t y) {
  uint8_t int16[6];
  uint16_t buffSize = 60;
  uint16_t w, h, bits;

  // dunno whats wrong with the first byte
  client.readBytes(int16, 1);

  // first two bytes are for the width
  // next two bytes are for the height
  read_b64(int16, 6);
  w = (int16[0] << 8) | int16[1];
  h = (int16[2] << 8) | int16[3];
  bits = (int16[4] << 8) | int16[5];
  
  uint16_t dataLeft = (w * h * ceil(bits/2));
  uint16_t blockSz = dataLeft < buffSize ? dataLeft : buffSize;
  uint8_t block[blockSz];

  // set pixel format
  if(bits == 16) {
    tft.writecommand(ILI9340_PIXFMT);
    tft.writedata(0x55);
  }
  else if(bits == 18) {
    tft.writecommand(ILI9340_PIXFMT);
    tft.writedata(0x56);
  }

  //set draw area on the screen
  tft.setAddrWindow(x, y, x+w-1, y+h-1);
  
  while(dataLeft > 0) {
    
    // maximum block size left
    //if(blockSz > dataLeft) blockSz = dataLeft;
    
    // read block
    read_b64(block, blockSz);

    // push to screen
    tft.pushData(block, blockSz);

    // adjust data left
    dataLeft -= blockSz;
  }
}

