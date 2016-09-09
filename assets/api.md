# Documentation

## Introduction

Dodo is an 8-bit 65C02 system. Stored in ROM is an ABI (Application Binary Interface) that exposes a Game API. Games are stored in an external 8KB ROM cartridge. 

### Screen

Dodo provides a 128x64 Monochrome OLED screen that is mapped to system memory. The screen layout is organized by page. There are 8 pages each containing 128 bytes that represent 128x8 pixels. Each byte contains a vertical slice of bitmap data where bit 0 is the top of the slice.

### Timing

Dodo's system clock runs at 1Mhz. Dodo contains a 6522 VIA Peripheral that fires an interrupt every 50,000 clock cycles, or every 50ms, to pump the game loop. A game that has its logic fit within the 50,000 cycle budget will run at 20 frames per scecond. Game music is also managed in the background by the interrupt.

#### Sample Game Loop

``` cpp
#include <stdio.h>
#include <stdlib.h>
#include "api.h"

int main() {
	// Initialize the API
	api_init();

	// Clear the graphics in video memory
	CLEAR();

	for (;;) {
		// Game Logic

		// Push video memory to the OLED
		DISPLAY();

		// Wait for next interrupt
		WAIT();
	}

	return 0;
}
```

## Sprites

## Sound

## API

### DRAW_SPRITE

``` cpp
DRAW_SPRITE(sprite, x, y, w, h, f, m);
```

sprite: pointer to the sprite image data
x: byte, x coordinate (performance does not vary based on value)
y: byte, y coorindate (performance is best when a multiple of 8)
w: byte, width (may be any value)
h: byte, height (must be a multiple of 8)
f: byte, boolean that specifies whether or not to flip the image horizontally
m: byte, drawing mode (mode may be one of the values below)

DRAW_NOP, Normal Drawing Mode, replaces everything underneath the bitmap
DRAW_OR, Draws using logical OR, fastest mode
DRAW_AND, Draws using logical AND
DRAW_XOR, Draws using logical XOR

When a game is designed with background graphics, it is common practice to have a mask for each bitmap drawn first using DRAW_AND and then the real bitmap is subsequently drawing using DRAW_OR. This method allows the sprite to be drawn with transparency.

The DRAW_XOR mode is useful for implementing a flashing cursor.

### DISPLAY

DISPLAY()

Pushes the contents of video memory to the display. This call is intended to be made once per game cycle.

### CLEAR_SPRITE

CLEAR_SPRITE(x, y, w, h)

x: byte, x coordinate
y: byte, y coordinate
w: byte, width
h: byte, height (must be multiple of 8)

### SET_PIXEL

SET_PIXEL(x, y, c)

x: byte, x coorinate of pixel
y: byte, y coodinate of pixel
c: byte, color, 0 for black, 1 for white

### DRAW_LINE

DRAW_LINE(x0, y0, x1, y1, c)

Bresenham line algorithm

x0: byte, first x coordinate of line
y0: byte, first y coordinate of line
x1: byte, second x coordinate of line
y1: byte, second y coordinate of line
c: byte, color, 0 for black, 1 for white

Computationally expensive, it is recommended to draw lines sparingly.

### DELAY_MS

DELAY_MS(delay)

delay: byte, number of milliseconds to pause for

Should be used to delay while showing a splash screen

### LED_ON

LED_ON()

Turns LED on

### LED_OFF

LED_OFF()

Turns LED off

### WAIT

WAIT()

Waits for the interrupt to fire. WAIT() should be called at the end of the game loop in order to synchronize the frame rate to a consistent 20 FPS.

### LOAD_MUSIC

LOAD_MUSIC(music)

music: byte, pointer to music

### PLAY_EFFECT

PLAY_EFFECT(effect)
SPI_ENABLE()
SPI_DISABLE()
SPI_WRITE(v)

### CLEAR

CLEAR()

Clears the video memory to erase the screen

### COPY_BACKGROUND

COPY_BACKGROUND(data, x, y, w, h, dir)

data: pointer to buffer
x: 
y:
w:
h:
dir: direction, 0 = vmem -> buffer, 1 = buffer -> vmem

Copying the background back and forth between video memory and a buffer is useful for games with background graphics. Typically a game should copy the background where a sprite will be drawn, draw the sprite, call DISPLAY to show the graphics, and then erase the sprite by copying the buffer back into video memory.

The buffer needs to be a page taller than the sprite. For instance, if the sprite is 24x16 pixels (2 pages tall, 48 bytes). The buffer needs to be 24*24 pixels (3 pages tall, 72 bytes)


DRAW_STRING(text)
SET_CURSOR(row, col)
READ_BUTTONS()