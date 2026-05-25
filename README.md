# ggvm - [Go Graphics][01] Virtual Machine

Intended to be a Stack based virtual machine, supporting push/pop, load/store, ALU instructions, and conditional/unconditional branching; but it will also have instructions which wrap the [fogleman/gg][01] library's functions. Essentially, a vm which takes input of an assembly file using a custom ISA, and is able to do everything which that ISA supports + output images.

## Why

Lately I've been using the gg library for an unrelated project, to generate custom plots. However, I found that to tweak a generated image, you must edit your Go code, and then recompile and run it, which gets tedious after a short while and is in my opinion, wasteful of compute resources. 

My initial plan was to write a simple wrapper; a line-by-line interpreter. It would keep a gg Context within the program, and each line of the input file would execute one function of the gg library, adding to the image. Of course the issue with this is that you lose the ability to loop, have functions, and do numerical computations which in a lot of cases is necessary. The alternative (which I've now settled upon) is a minimal VM. The VM, just like the line-by-line interpreter, will also keep a gg Context and have instructions which run the gg library functions; but it'll also have the standard stack-based vm push/pop, load/store, branch, etc instructions which bring back the ability to loop, recurse, and perform math. I've also wanted to write a VM for a long time :sweat_smile:.

## Plan

Below is all the things which I want the vm to be capable of doing, on top of everything already mentioned:

1. It should be capable of commenting out lines, or putting comments after code on a line. For instance `;this comment` could be its own line, or be after code.
2. It should be able to import code from other files. So for instance, if drawing a line is a primitive assembly instruction, and multiple lines can be used to draw an arrow, there could be an `arrow.ggvm` file which has a function that draws an arrow on the gg context, and consumes parameters for the arrow off the stack. The `arrow.ggvm` file could simply have a jump point called `arrow_func`, and code in the main script could just push values onto the stack, and then run `JUMP arrow.arrow_func`. All such imported libraries will be local, or use the \~/ggvm directory for the library; I guess those could be called system-wide libraries. Importing should be relative to local path, or the \~/ggvm path. Something like `import abcd/efg/library_name` to import `~/ggvm/abcd/efg/library_name.ggvm`, or `./abcd/efg/library_name.ggvm`. If there's a conflict, it should choose the local version.
3. I also want there to be a CLI version which interprets in real time. In this cli version, running `help <command>` should immediately print the parameters and usage of the assembly instruction, and be otherwise omitted from the actual code. If possible, this should work for external functions as well, printing comments directly above the function's jump-point. So `HELP arrow.arrow_func` should print the usage of the arrow function, after importing it. 
4. Instructions and function names (jump points) should *NOT* be case sensitive. File/module names *WILL* be case sensitive.
5. In the CLI version, only errors and help messages should be printed. When running a file, it should print everything, but not help messages (ignore them).
6. Tokenization will be done in a state machine. yet to make that.
7. Whenever there's an error, the stack and variables are NOT altered at all. And neither is the gg context.
8. Finally, if there's time: 
    - maybe make a standard library? The gg library comes with a *lot* of built-in functions, so I think there shouldnt be a need for it.

## ISA

- Basic VM instructions:
    - **IMPORT <filepath>:** Import the file specified by the filepath. It should work with/without the file extension at the end.
    - **PUSH <value>:** Pushes the next token onto the stack. This token could be a string, int64, or float64. A float will be if there's a decimal.
    - **POP:** Pops the value off the stack and discards it.
    - **DUP:** Duplicates the topmost value of the stack.
    - **SWAP:** Swaps the two topmost values on the stack.
    - **STORE <variable name>:** Stores the topmost value of the stack into a variable.
    - **LOAD <variable name>:** Loads the value from the specified variable onto the top of the stack.
    - **ADD:** Pops two off the stack and pushes the sum.
    - **SUB:** Pops two off the stack and pushes the difference (top - bottom).
    - **MUL:** Pops two off the stack and pushes the product.
    - **DIV:** Pops two off the stack and pushes the quotient (top - bottom).
    - **SQRT:** Pops one off the stack and pushes the square root.
    - **LT:** Pops two off the stack, and pushes 1 if top < bottom, and 0 if otherwise.
    - **EQ:** Pops two off the stack, and pushes 1 if top = bottom, and 0 if otherwise.
    - **JUMP <target>:** Jumps unconditionally to the target.
    - **JUMPZ <target>:** Pops the top off the stack, and jumps to the target if the value was 0.
    - **CASTINT:** Pops the top off the stack and pushes that value cast to an int.
    - **CASTFLOAT:** Pops the top off the stack and pushes that value cast to a float.
    - **HELP <instruction/function>:** If the instruction/function exists, print the help for it. If not, throw an error.

> [!NOTE]
> Going through all the gg instructions is taking too long. I'll update it later.

- gg instructions:
    - **ggNEWCONTEXT:** Pops width and height off the stack, and pushes Context pointer. This can be stored to a variable.
    - **ggACTIVECONTEXT <variable name>:** This will set the active context to the variable specified.
    - **ggSAVE <filename>:** Pops the gg context pointer from the stack and saves to filename. If this is split up by forward slashes, it will be saved to that directory.
    - **ggDEGREES:** Converts stack top from radians to degrees.
    - **ggRADIANS:** Converts stack top from degrees to radians.
    - **ggCLEAR**
    - **ggCLIP**
    - **ggRESETCLIP**
    - **ggCLOSEPATH**
    - **ggARC**
    - **ggCIRCLE**
    - **ggELLIPSE**
    - **ggIMAGE**
    - **ggLINE**
    - **ggPOINT**
    - **ggRECT**
    - **ggSTRING**
    - **ggMOVETO**
    - **ggLINETO**
    - **ggQUADTO**
    - **ggCUBICTO**
    - **ggPOP**
    - **ggPUSH**
    - **ggSCALEABT**
    - **ggRGBA**
    - **ggSHEAR**
    - **ggSTRLEN**
    - **ggFILL**
    - **ggFILLPRES**
    - **ggSETLINEWIDTH**
    - **ggTRANSLATE**
    - **ggINVERTMASK**
    - **ggFONTSIZE**
    - **ggFONTFILE**

<!-- Links -->
[01]: https://github.com/fogleman/gg
