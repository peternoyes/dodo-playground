# Dodo: 6502 Game System

![Dodo](/assets/dodo.png)

## Introduction

Dodo is an open source 8-bit 6502 portable game system. The concept of Dodo is to be a nostalgic system with real hardware and a great development environment. Rather than being built around a Raspberry PI, Dodo is constructed with real 8-bit chips. The system can be hand soldered easily at home because it uses all DIP components.

This website provides a way to quickly develop games in an online IDE with an integrated simulator. Finished games can be flashed to a game cartridge over a USB serial cable. Games developed in the playground have a unique hyperlink that can be shared.

### Links

- [dodolabs.io](http://www.dodolabs.io)
- Project on [hackaday.io](https://hackaday.io/project/9325-dodo-6502-game-system)
- Project on [Github](https://github.com/peternoyes/dodo)

## Development

The development stack for Dodo is built around cc65, the 6502 C compiler. Stored in ROM is an ABI (Application Binary Interface) written in Assembly that exposes a Game API. Games are stored in an external 8KB FRAM cartridge. 

## Skeleton Code

Below is a simple outline of a game's structure.


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

## 6502 Assembly

It is also possible to write Dodo games in 6502 assembly. There is a toggle in the navigation bar to specify the language preference. The assembly API is nearly identical to the C API.

### Function Names

The functions in assembly are simply the C names but all lowercase. For instance

``` cpp
CLEAR();
```

would be

``` assembly
jsr clear
```

### Calling Convention

The parameters in assembly are also the same as in C, except that they need to manually be pushed onto a stack. There are two functions for pushing parameters, pusha and pushax. For byte parameters, load the value int the A register and call pusha. For pointer parameters which are 16-bit, the upper and lower bytes need to be loaded into the A and X registers and then call pushax.

Example

``` assembly
lda #4          ; row
jsr pusha
lda #3          ; column
jsr pusha
jsr set_cursor 	; call set_cursor

lda #<message   ; pointer to message string
ldx #>message
jsr pushax      ; push pointer onto stack
jsr draw_string ; call draw_string

...

; Null terminated string
message: .byte "Hello, World", $0

```

### Assembly Skeleton

``` assembly
    .include "api.inc65"
    .setcpu "6502"
    .export main

main:
    ; clear the graphics in video memory
    jsr clear

loop:
    ; game logic

    ; push video memory to the OLED
    jsr display

    ; wait for the next interrupt
    jsr wait

    jmp loop

```

 
## Screen

Dodo provides a 128x64 Monochrome OLED screen that is mapped to system memory. The screen layout is organized by page. There are 8 pages each containing 128 bytes that represent a block of 128x8 pixels. Each byte contains a vertical slice of bitmap data where bit 0 is the top of the slice.

## Timing

Dodo's system clock runs at 1Mhz. The hardware includes a 6522 VIA Peripheral that fires an interrupt every 50,000 clock cycles to pump the game loop. A game that has its logic fit within the 50,000 cycle budget will run at 20 frames per scecond. Game music is also managed in the background by the interrupt.

## Sprites

Sprites are layed out in the same manner as the screen. Sprites must have their height be a multiple of 8 pixels, but the width is arbitrary. Sprites can be drawn flipped along the vertical axis as an option. Sprite drawing also supports a number of boolean drawing operations.

## Sound

Dodo supports a single channel of audio of either background music or a sound effect. Once background music is initialized it will repeat endlessly. When a sound effect is activated, the background music is silenced while the effect plays. Once the effect is complete, the music will resume.

Dodo supports two octaves of notes (more is possible in a future revision). The music data is stored in a byte array that contains alternating frequencies and durations. A frequency of 0 represents silence. The array is terminated by two consecutive 0's. Eachh duration equals 50ms.

A trick to play staccato notes is to slightly shorten the duration of a note and follow with a small gap of silence to fill the remaing space. For instance, if a staccato 'C' with a duration of 400ms is desired to be played on repeat here is the corresponding byte array:

``` cpp
static byte const _music[6] = { 238, 7, 0, 1, 0, 0};
```

In the above example 238 is the frequency for 'C'. The 7 accounts for 350ms of sound. The 0 represents that the duration will be silent. The 1 accounts for the remaining 50ms of silence. The two 0's at the end terminate the sequence.

Below are the frequncy values for the notes in each octave.

Note | Octave 1 | Octave 2
-----|----------|----------
B    | 251      | 125
C    | 238      | 118
C#   | 224      | 110
D    | 210      | 104
D#   | 199      | 99
E    | 188      | 93
F    | 177      | 88
F#   | 168      | 83
G    | 158      | 78
G#   | 149      | 74
A    | 140      | 69
A#   | 133      | 65

## API

### DRAW_SPRITE

``` cpp
DRAW_SPRITE(sprite, x, y, w, h, f, m);
```

Parameter    | Type       | Description
-------------|------------|-----------------------------
sprite       | *byte      | pointer to the sprite image data
x            | byte       | x coordinate
y            | byte       | y coordinate
w            | byte       | width of sprite
h            | byte       | height of sprite, must be multiple of 8
f            | byte       | boolean that specifies whether or not to flip horizontally
m            | mode       | drawing mode, see below

*Note: Performance will be maximized when y is a multiple of 8

Drawing Modes

Mode     | Description
---------|-------------
DRAW_NOP | normal, replaces everything underneath the sprite
DRAW_OR  | logical OR, fastest mode
DRAW_AND | logical AND
DRAW_XOR | logical XOR

When a game is designed with background graphics, it is common practice to have a mask for each sprite that is drawn using DRAW_AND with the sprite subsequently drawn using DRAW_OR. This method allows the sprite to be drawn with transparency.

The DRAW_XOR mode is useful for implementing a flashing cursor.

### DISPLAY

``` cpp
DISPLAY();
```

Pushes the contents of video memory to the display. This call is intended to be made once per game cycle.

### CLEAR_SPRITE

``` cpp
CLEAR_SPRITE(x, y, w, h);
```

Parameter    | Type       | Description
-------------|------------|-----------------------------
x            | byte       | x coordinate
y            | byte       | y coordinate
w            | byte       | width
h            | byte       | height, must be multiple of 8

Erases the rectangular portiion of the screen defined by the parameters. Note that background graphics will be erased as well.

### GET_PIXEL

```cpp
GET_PIXEL(x, y);
```

Parameter    | Type       | Description
-------------|------------|-----------------------------
x            | byte       | x coordinate
y            | byte       | y coordinate

Returns the color of the pixel at the specified coordinates, 0 for black, 1 for white.

### SET_PIXEL

``` cpp
SET_PIXEL(x, y, c);
```

Parameter    | Type       | Description
-------------|------------|-----------------------------
x            | byte       | x coordinate
y            | byte       | y coordinate
c            | byte       | color, 0 for black, 1 for white

Sets a pixel to a specific color

### DRAW_LINE

``` cpp
DRAW_LINE(x0, y0, x1, y1, c);
```

Bresenham line algorithm

Parameter    | Type       | Description
-------------|------------|-----------------------------
x0			 | byte 	  | x coordinate of first point
y0           | byte       | y coordinate of first point
x1           | byte       | x coordinate of second point
y1           | byte       | y coordinate of second point
c            | byte       | color, 0 for black, 1 for white

*Note: Computationally expensive, it is recommended to draw lines sparingly.

### DELAY_MS

``` cpp
DELAY_MS(delay);
```

Parameter    | Type       | Description
-------------|------------|-----------------------------
delay        | byte       | delay in milliseconds

*Note: Should be used sparingly such as to delay while showing a splash screen

### LED_ON

``` cpp
LED_ON();
```

Turns LED on (No effect in simulator)

### LED_OFF

``` cpp
LED_OFF();
```

Turns LED off

### WAIT

``` cpp
WAIT();
```

Waits for an interrupt to fire. WAIT() should be called at the end of the game loop in order to synchronize the frame rate to a consistent 20 FPS.

### LOAD_MUSIC

LOAD_MUSIC(music);

Parameter    | Type       | Description
-------------|------------|-----------------------------
music        | *byte      | pointer to music

See the sound section above for a description of the music format

### PLAY_EFFECT

Parameter    | Type       | Description
-------------|------------|-----------------------------
effect       | *byte      | pointer to effect

PLAY_EFFECT(effect);

The sound effects are stored in the same format as the music. For an affect it may be of benefit to use more frequencies than just those that correspond to notes. 

An effect will play on repeat until a subsequent call to PLAY_EFFECT with 0 passed in.

### CLEAR

``` cpp
CLEAR();
```

Clears the video memory to erase the screen

*Note: A call to DISPLAY() is required to see the results of a call to CLEAR()

### COPY_BACKGROUND

``` cpp
COPY_BACKGROUND(data, x, y, w, h, dir);
```

Parameter    | Type       | Description
-------------|------------|-----------------------------
data         | *byte      | pointer to byte array
x            | byte       | x coordinate
y            | byte       | y coordinate
w            | byte       | width
h            | byte       | height
dir          | byte       | direction, 0 = vmem -> buffer, 1 = buffer -> vmem

Copying the background back and forth between video memory and a buffer is useful for games with background graphics. This technique would be used instead of calling CLEAR_SPRITE(). Typically a game should copy the background where a sprite will be drawn, draw the sprite, call DISPLAY() to show the graphics, and then erase the sprite by copying the buffer back into video memory.

The buffer needs to be a page taller than the sprite. For instance, if the sprite is 24x16 pixels (2 pages tall, 48 total bytes). The buffer needs to be 24*24 pixels (3 pages tall, 72 total bytes)

### DRAW_STRING

``` cpp
DRAW_STRING(text);
```

Parameter    | Type       | Description
-------------|------------|-----------------------------
text         | *char      | ANSI string to be displayed

Draws the ANSI text string at the current cursor location. The screen supports 8 rows and 17 columns of text. 

### SET_CURSOR

``` cpp
SET_CURSOR(row, col);
```

Parameter    | Type       | Description
-------------|------------|-----------------------------
row          | byte       | row of cursor
col          | byte       | column of cursor

Moves the cursor for subsequent calls to DRAW_STRING

### READ_BUTTONS

``` cpp
READ_BUTTONS();
```

Returns a byte that is packed with the button state. For each bit that is unset the corresponding button is pushed. 

Bit Position | Mask     |Button
-------------|----------|-----------
1            | 1        | up
2            | 2        | down
3            | 4        | left
4            | 8        | right
5            | 16       | a
6            | 32       | b

For example:

``` cpp
buttons = READ_BUTTONS();
if ((buttons & 4) == 0) {
	move_left();
}
```