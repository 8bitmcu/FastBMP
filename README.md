## FastBMP - A fast ILI9341 Driver for the ESP8266 Wi-Fi modules

This projects aims at driving the ILI9341 using the ESP8266 module as fast as possible. Based off [Adafruit's ILI9341](https://github.com/adafruit/Adafruit_ILI9341) display library, it is much faster at displaying an image, thanks to the following tweaks:

* Using SPI transactions to send on average 240 pixels at the time, instead of one at the time per transaction
* Using SPI at it's full 80mhz as supported by the ILI9341 and ESP8266
* Reducing overhead by using custom pixel-based image format instead of BMP

