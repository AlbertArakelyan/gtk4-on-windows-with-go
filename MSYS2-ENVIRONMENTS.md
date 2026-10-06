# MSYS2 environments: which terminal to use

MSYS2 is a single install (`C:\msys64`) that contains several separate **environments**.
Each environment has its own folder, its own compiler and its own copies of the libraries.
Opening a terminal only decides which of these folders is first on your `PATH`.

| Terminal | Folder | Package prefix | Compiler | Built programs depend on |
|---|---|---|---|---|
| **MSYS** | `/usr` | none (e.g. `make`, `git`) | gcc (msys-only programs) | `msys-2.0.dll` (POSIX layer) |
| **UCRT64** | `/ucrt64` | `mingw-w64-ucrt-x86_64-` | gcc | Universal C Runtime (modern Windows) |
| **MINGW64** | `/mingw64` | `mingw-w64-x86_64-` | gcc | `msvcrt.dll` (legacy C runtime) |
| **CLANG64** | `/clang64` | `mingw-w64-clang-x86_64-` | clang / LLVM | Universal C Runtime |
| **CLANGARM64** | `/clangarm64` | `mingw-w64-clang-aarch64-` | clang / LLVM | Universal C Runtime, ARM64 Windows |

On disk, `/ucrt64` is `C:\msys64\ucrt64`, `/usr` is `C:\msys64\usr`, and so on.

## Which one to choose

- **MSYS**: for the Unix-style tools themselves: `pacman`, bash scripts, `grep`, `sed`, `make`, `autoconf` and so on.
  Don't use it to build apps you'll ship, because anything built there needs `msys-2.0.dll`.
- **UCRT64**: **the default for building real Windows programs.** msys2 recommends it, and it's what this project uses (GTK 4, gcc).
- **MINGW64**: only for old projects or older Windows versions that need the legacy `msvcrt` runtime.
  It's the same idea as UCRT64 with a different runtime underneath.
- **CLANG64**: when you want clang/LLVM tooling (clang-tidy, LLVM sanitizers, lld) instead of gcc.
- **CLANGARM64**: only when building on or for ARM64 Windows.

## Rules of thumb

1. **Install packages with your environment's prefix.** For UCRT64 that's `mingw-w64-ucrt-x86_64-<name>`.
   `pacman` works from any terminal and puts each package into its own environment's folder.
2. **Don't mix environments.** Something built in UCRT64 must link against UCRT64 libraries.
   Mixing them gives linker errors or crashes at startup.
3. **When you ship an app**, the DLLs it needs come from that environment's `bin` folder (`C:\msys64\ucrt64\bin`).
   Run `ldd ./app.exe` in the matching terminal to list them.
4. **Check where you are**: `echo $MSYSTEM` prints `MSYS`, `UCRT64`, ... and `which gcc` shows which compiler you'd get.

## Common tools: which package to install

The examples use the UCRT64 prefix. For another environment, swap the prefix (see the table).

### Compilers: gcc / g++

```sh
pacman -S mingw-w64-ucrt-x86_64-gcc        # provides gcc.exe and g++.exe (gfortran is a separate package)
pacman -S mingw-w64-ucrt-x86_64-toolchain  # group: gcc, g++, gdb, make, binutils, pkgconf, ...
```

- `g++` comes in the same package as `gcc`; there's no separate `g++` package.
- The `gcc` package **without** a prefix belongs to MSYS and builds programs that need `msys-2.0.dll`. You usually don't want it.
- For clang, use the CLANG64 environment (`mingw-w64-clang-x86_64-clang`) rather than installing clang into UCRT64.

### make

There are two different `make` programs, and both are valid:

| Package | Command | Where it comes from | Use it when |
|---|---|---|---|
| `make` | `make` | MSYS (`/usr/bin`) | Unix-style Makefiles and autotools projects (they expect a POSIX shell, `rm`, `cp`, ...) |
| `mingw-w64-ucrt-x86_64-make` | `mingw32-make` | UCRT64 | Makefiles written for `cmd.exe`, or CMake's "MinGW Makefiles" generator |

Inside the UCRT64 terminal, the MSYS `make` running UCRT64's `gcc` is completely normal. `make` only orchestrates the build; the compiler decides what the output links against.

### cmake and ninja

```sh
pacman -S mingw-w64-ucrt-x86_64-cmake mingw-w64-ucrt-x86_64-ninja
```

- **Always use the environment's `cmake`**, not the MSYS one. MSYS `cmake` thinks it's on a Unix-like system and finds the wrong compilers and libraries.
- Recommended setup is CMake with Ninja:
  ```sh
  cmake -G Ninja -B build -DCMAKE_BUILD_TYPE=Release
  cmake --build build
  ```
- Generator choices:
  - `-G Ninja`: fastest, the usual choice.
  - `-G "MSYS Makefiles"`: uses MSYS `make`, from inside an msys2 terminal.
  - `-G "MinGW Makefiles"`: uses `mingw32-make`, e.g. from PowerShell.

### autotools (`./configure && make`)

```sh
pacman -S base-devel autotools mingw-w64-ucrt-x86_64-toolchain
```

Run `./configure` from the **UCRT64 terminal**, so it detects UCRT64's `gcc` and `pkg-config`.
The build tools themselves (autoconf, automake, libtool, MSYS `make`) come from MSYS.

### pkg-config

```sh
pacman -S mingw-w64-ucrt-x86_64-pkgconf
```

Use the environment's own `pkg-config`, so it reports paths to that environment's libraries.
That's why `build.ps1` sets `PKG_CONFIG=C:\msys64\ucrt64\bin\pkg-config.exe`.

### gdb (debugger)

```sh
pacman -S mingw-w64-ucrt-x86_64-gdb
```

Debug a program with the `gdb` from the environment it was built in.

### Libraries (GTK, SDL, Boost, OpenSSL, ...)

Always install the environment-prefixed version:

```sh
pacman -S mingw-w64-ucrt-x86_64-gtk4
pacman -S mingw-w64-ucrt-x86_64-SDL2
pacman -S mingw-w64-ucrt-x86_64-boost
```

Find packages with `pacman -Ss <name>`; the results show which environment each one belongs to.

### python

- `mingw-w64-ucrt-x86_64-python` is a native Windows Python. Use it for Python bindings to UCRT64 libraries (e.g. PyGObject with GTK).
- `python` without a prefix is MSYS Python. It's fine for scripts and build tools but isn't a native Windows Python.

### git

The MSYS `git` package works well inside msys2 terminals.
If you already use Git for Windows, keep using that instead; you don't need both.

### Go, Rust and other toolchains from outside msys2

You don't install these through msys2, but they can use its C compiler and libraries:

- **Go with cgo** (like this project): put `C:\msys64\ucrt64\bin` first on `PATH` and set `CGO_ENABLED=1`.
  Point `CC` and `PKG_CONFIG` at the UCRT64 tools. `build.ps1` does all of this.
- **Rust**: use the `x86_64-pc-windows-gnu` target, with `C:\msys64\ucrt64\bin` on `PATH` so it finds UCRT64's linker.

## Using msys2 tools from PowerShell, cmd or VS Code

You don't have to open an msys2 terminal. Put **one** environment's `bin` folder on `PATH`:

```powershell
$env:PATH = "C:\msys64\ucrt64\bin;$env:PATH"
```

- Don't add several environments' `bin` folders at once; you'll get a mix of DLLs.
- Watch out for other compilers already on `PATH`, e.g. a scoop- or Chocolatey-installed `gcc`. Whichever is first wins.
- In VS Code you can add an "MSYS2 UCRT64" terminal profile that starts
  `C:\msys64\usr\bin\bash.exe --login` with the environment variable `MSYSTEM=UCRT64`.

## Useful pacman commands

| Command | What it does |
|---|---|
| `pacman -Syu` | Update everything (run it again if it closes the terminal partway) |
| `pacman -Ss <name>` | Search packages |
| `pacman -S <pkg>` | Install a package |
| `pacman -S --needed <pkg>` | Install only if it's missing or outdated |
| `pacman -Rns <pkg>` | Remove a package and the dependencies nothing else uses |
| `pacman -Q` | List installed packages |
| `pacman -Qo <file>` | Show which package owns a file |
| `pacman -Ql <pkg>` | List the files a package installed |
