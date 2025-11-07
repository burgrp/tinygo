.section .init
.global _start
.type _start,@function

_start:

#if QINGKE_V2
    // On QingKe V2 (CH32V003...), mtvec.BASE must be 1 KB aligned regardless of mode.
    // To avoid wasting a 1 KB alignment gap for the interrupt handler,
    // this code merges both the reset and trap entry into a single
    // routine.  mtvec is set to this address in MODE=00 (direct).
    // The trap handler checks whether the interrupt is from PFIC
    // (peripheral interrupt controller) by reading RCC_RSTSCKR.
    // If so, it jumps to handleInterruptASM; otherwise, it continues
    // normal execution from the reset handler.
    li   t0, 0x40021024      // RCC_RSTSCKR
    lw   t1, 0(t0)
    li   t2, 0xFC000000      // PFIC_GISR_GACTSTA_Msk
    and  t1, t1, t2
    beqz t1, handleInterruptASM
    li   t1, 0x01000000
    sw   t1, 0(t0)
#endif

    // If we're on a multicore system, we need to wait for hart 0 to wake us up.
#if TINYGO_CORES > 1
    csrr a0, mhartid

    // Hart 0 stack
    bnez a0, 1f
    la sp,      _stack_top

1:
    // Hart 1 stack
    li a1, 1
    bne a0, a1, 2f
    la sp,      _stack1_top

2:
    // Hart 2 stack
    #if TINYGO_CORES >= 3
    li a1, 2
    bne a0, a1, 3f
    la sp,      _stack2_top
    #endif

3:
    // Hart 3 stack
    #if TINYGO_CORES >= 4
    li a1, 3
    bne a0, a1, 4f
    la sp,      _stack3_top
    #endif

4:
    // done

#if TINYGO_CORES > 4
#error only up to 4 cores are supported at the moment!
#endif

#else
    // Load the stack pointer.
    la sp,      _stack_top
#endif

    // Load the globals pointer. The program will load pointers relative to this
    // register, so it must be set to the right value on startup.
    // See: https://gnu-mcu-eclipse.github.io/arch/riscv/programmer/#the-gp-global-pointer-register
    // Linker relaxations must be disabled to avoid the initialization beign
    // relaxed with an uninitialized global pointer: mv gp, gp
    .option push
    .option norelax
    la gp,      __global_pointer$
    .option pop

    // Jump to runtime.main
    call main
