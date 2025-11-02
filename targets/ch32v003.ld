MEMORY
{
    FLASH (rx) : ORIGIN = 0x00000000, LENGTH = 16K
    RAM (rwx)  : ORIGIN = 0x20000000, LENGTH = 2K
}

REGION_ALIAS("FLASH_TEXT", FLASH)

_stack_size = 512;

INCLUDE "targets/riscv.ld"