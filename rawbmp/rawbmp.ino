

/***************************************************
  This is an example sketch for the Adafruit 2.2" SPI display.
  This library works with the Adafruit 2.2" TFT Breakout w/SD card
  ----> http://www.adafruit.com/products/1480

  Check out the links above for our tutorials and wiring diagrams
  These displays use SPI to communicate, 4 or 5 pins are required to
  interface (RST is optional)
  Adafruit invests time and resources providing this open source code,
  please support Adafruit and open-source hardware by purchasing
  products from Adafruit!

  Written by Limor Fried/Ladyada for Adafruit Industries.
  MIT license, all text above must be included in any redistribution
 ****************************************************/

#include "SPI.h"
#include "Adafruit_GFX.h"
#include "Adafruit_ILI9340.h"
#include <ESP8266WiFi.h>
#include <ArduinoJson.h>
#include <math.h>
#include "Base_64.h"
#include "ds8.h"
#include "wificonf.h"

Adafruit_ILI9340 tft = Adafruit_ILI9340(15, 2, 4);


void setup() {

  Serial.begin(9600);
  while (!Serial);

  tft.begin();
}


void wifiConnect(int timeout) {
  tft.setTextColor(ILI9340_WHITE);
  
  tft.print("wifi init ");
  tft.print(SSID_NAME);
  WiFi.begin(SSID_NAME, SSID_PASS);

  unsigned long tout = millis() + timeout;
  unsigned long ldot = tout;
  while(WiFi.status() != WL_CONNECTED && tout > millis()) {
    if(millis() - ldot > 500) {
      tft.print(".");
      ldot = millis();
    }

    yield();
  }

  if(WiFi.status() != WL_CONNECTED) {
    tft.setTextColor(ILI9340_RED);
    tft.println(" BAD!");
    return;
  }
  else {
    tft.setTextColor(ILI9340_GREEN);
    tft.println(" OK!");

    tft.setTextColor(ILI9340_WHITE);
    tft.print("  ip: ");
    tft.println(WiFi.localIP());

    tft.print("  signal: ");
    tft.print(WiFi.RSSI());
    tft.println("dBm");
  }
}



uint16_t read16(WiFiClient & f) {
  uint8_t d[2];
  f.readBytes(d, 2);
  return (d[1] << 8) | d[0];
}

uint32_t read32(WiFiClient & f) {
  uint8_t d[4];
  f.readBytes(d, 4);
  return (d[3] << 24) | (d[2] << 16) | (d[1] << 8) | d[0];
}



void read_b64(WiFiClient &f, uint8_t *output, uint16_t outputLen) {
  uint16_t bufSize = (outputLen/3)*4;
  uint8_t buf[bufSize];
  
  size_t idx = 0;
  while(idx < bufSize) {
    idx += f.read(&buf[idx], bufSize-idx);
    yield();
  }

  base64_decode((char *) output,(char *) buf, bufSize);
}


void rawDraw(WiFiClient &bmpFile, uint16_t x, uint16_t y) {
  uint8_t int16[6];
  uint16_t buffSize = 240;
  uint16_t w, h, bits;

  // dunno whats wrong with the first byte
  bmpFile.readBytes(int16, 1);

  // first two bytes are for the width
  // next two bytes are for the height
  read_b64(bmpFile, int16, 6);
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
    //if(dataLeft < 240) blockSz = dataLeft;
    
    // read block
    read_b64(bmpFile, block, blockSz);

    // push to screen
    tft.pushData(block, blockSz);

    // adjust data left
    dataLeft -= blockSz;
  }

  
  tft.setCursor(0, 0);
  tft.println();
  tft.print("done");
}




void getImage(char* host, int16_t port, char* url) {
  tft.setTextColor(ILI9340_WHITE);
  tft.print("con ");
  tft.print(host);
  tft.print(":");
  tft.print(port);

  WiFiClient client;
  if (!client.connect(host, port)) {
    tft.setTextColor(ILI9340_RED);
    tft.println(" BAD!");
    return;
  }
  else {
    tft.setTextColor(ILI9340_GREEN);
    tft.println(" OK!");
  }

  String req = String("GET ") + url + " HTTP/1.1\r\n" +
               "Host: " + host + "\r\n" +
               "Connection: close\r\n\r\n";
               
  tft.setTextColor(ILI9340_WHITE);
  tft.println("req");
  tft.println(req);




  client.print(req);
  

  unsigned long time = millis();
  while (!client.available()) {
    //todo: timeout
    yield();
  }
  yield();



  tft.println("res");

  bool headers = true;
  bool bmp = false, raw = false;

  while (client.available() && headers) {

    String line = client.readStringUntil('\r');

    tft.println(String("  ") + line);
    
    if(line.indexOf("HTTP") && line.indexOf("200")) {
      
    }

    if(line.indexOf("Content-Type") > 0) {
      raw = line.indexOf("application/octet-stream") > 0;
    }

    if (line == "\n") headers = false;
  }


  rawDraw(client, 0, 0);


  tft.setCursor(0, 0);
}




void loop(void) {

  tft.setRotation(2);
  tft.fillScreen(ILI9340_BLACK);

  tft.setFont(&Nintendo_DS_BIOS8pt7b);
  tft.setCursor(0, 0);
  tft.println();


  //introduce();
  wifiConnect(60 * 5 * 1000);

  while(1) {

    
    getImage("192.168.0.19", 8088, "/random?format=raw&bits=16");
    //getImage("192.168.0.22", 8088, "/local?img=subpixelroll.png&format=raw");

    unsigned long time = millis();

    while(time + 1000*10 > millis()) {
      delay(2000);
      yield();
      wdt_disable();
    }
  }
}

