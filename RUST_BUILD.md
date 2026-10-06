# GTK 4 with Rust on Windows (compared to Go)

> **Status: not tested.** This project is written in Go. These are notes on how the same setup would look for a Rust GTK app,
> based on the [gtk-rs book](https://gtk-rs.org/gtk4-rs/stable/latest/book/installation_windows.html).
> Check the book's current instructions before setting it up.

## Short version

Mostly the same idea as Go: GTK is still a **C library**, so on Windows you still need msys2 (or another way to get GTK),
**pkg-config**, and the **GTK DLLs** on `PATH` or bundled next to the exe.
The differences are in how the bindings work and which Windows toolchain you pick.

## What stays the same

- **GTK comes from msys2:** the same GTK 4 and pkg-config packages, installed with `pacman`.
- **pkg-config:** the Rust bindings ([gtk4-rs](https://gtk-rs.org/)) use it at build time to find GTK.
- **DLLs at run time:** the exe needs the GTK DLLs, through `PATH` or copied next to it.
  The same `dist` approach works, including the `ldd` trick from `package.ps1`.
- **Linux:** same as Go. `sudo apt install libgtk-4-dev`, then `cargo build` (see [LINUX_BUILD.md](LINUX_BUILD.md)).

## What's different

### No cgo

Go needs **cgo**, which compiles a lot of C glue code with gcc. That's why the first Go build takes several minutes.
gtk4-rs consists of generated and hand-written Rust bindings that only **link** to GTK, so there's much less C work.
The first build is still slowish because Rust compiles many crates, but for a different reason.

### Rust has two Windows toolchains

| | **GNU** (`x86_64-pc-windows-gnu`) | **MSVC** (`x86_64-pc-windows-msvc`, rustup's default) |
|---|---|---|
| GTK from | msys2 `pacman` | built yourself with [gvsbuild](https://github.com/wingtk/gvsbuild) (Visual Studio) |
| Linker | gcc from msys2 | Visual Studio's linker |
| Setup | `rustup toolchain install stable-gnu`, then `rustup default stable-gnu` (or use it per project) | Install Visual Studio Build Tools and Python, run gvsbuild (slow, builds GTK from source), set `PATH`, `PKG_CONFIG_PATH` and `LIB` |
| Feels like | the Go setup in this repo | the "native Windows" route |

Plain `cargo build` with the default MSVC toolchain **won't find msys2's GTK**. Check which toolchain you're on with:

```powershell
rustup show
```

### The msys2 environment catch (GNU route)

The gtk-rs book's GNU instructions use the **MINGW64** packages (`mingw-w64-x86_64-gtk4`, with `C:\msys64\mingw64\bin` on `PATH`), **not UCRT64**.
rustup's GNU toolchain has traditionally been built against the old `msvcrt` runtime, and mixing it with UCRT64 libraries can cause problems.

Two ways to handle it:

1. **Follow the book:** use the MINGW64 packages and put `C:\msys64\mingw64\bin` on `PATH`.
2. **Stay on UCRT64:** use msys2's own Rust, which matches UCRT64:
   ```sh
   pacman -S mingw-w64-ucrt-x86_64-rust
   ```

Remember the rule from [MSYS2-ENVIRONMENTS.md](MSYS2-ENVIRONMENTS.md): **only one environment's `bin` folder on `PATH`.**
This machine has `C:\msys64\ucrt64\bin` on `PATH`, so either use msys2's Rust, or swap that entry for `mingw64\bin` if you follow the book.

## Minimal setup

```toml
# Cargo.toml
[dependencies]
gtk = { version = "0.9", package = "gtk4", features = ["v4_12"] }
```

- `package = "gtk4"` imports the `gtk4` crate under the shorter name `gtk`.
- The `v4_xx` feature turns on APIs up to that GTK version. Keep it **at or below** the GTK installed on your system and your users' systems.
- Check the current version of the `gtk4` crate on [crates.io](https://crates.io/crates/gtk4); `0.9` may be out of date.

## Go vs Rust at a glance

| | Go (this repo) | Rust |
|---|---|---|
| Bindings | gotk4 (cgo) | gtk4-rs (links to GTK, no C compilation) |
| Needs gcc | yes, cgo compiles C code | GNU toolchain: as the linker; MSVC toolchain: no |
| GTK on Windows | msys2 UCRT64 | msys2 (GNU) or gvsbuild (MSVC) |
| pkg-config | yes | yes |
| DLLs at run time | `PATH` or `dist` folder | same |
| Slow first build because | cgo compiles C glue | many Rust crates to compile |
