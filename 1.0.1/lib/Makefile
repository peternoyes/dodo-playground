C_SOURCES = main.c
ASM_SOURCES = crt0.s65 zeropage.s65

# Compilation of C files
%.o: %.c
	/home/cc65/bin/cc65 --cpu 6502 -O -t none -o $(@:.o=.s) $<
	/home/cc65/bin/ca65 --cpu 6502 -o $@ -l $(@:.o=.lst) $(<:.c=.s)

# Compilation of assembler files
%.o: %.s65
	/home/cc65/bin/ca65 --cpu 6502 -o $@ -l $(@:.o=.lst) $<

all: fram.bin

fram.bin: $(ASM_SOURCES:.s65=.o) $(C_SOURCES:.c=.o)
	/home/cc65/bin/cl65 -C fram.cfg -m fram.map -o $@ $^ dodo.lib

clean:
	rm -f fram.bin *.s *.o *.lst *.map
