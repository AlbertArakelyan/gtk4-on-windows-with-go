# Building on Linux

> **Status: not tested yet.** This project has only been built and run on Windows so far.
> These notes describe how the Linux build is expected to work; the Linux template is on the [roadmap](README.md#roadmap).

## Short version

On Linux you **don't need msys2, PATH changes or a `dist` folder** for development.
Install the GTK 4 development package with your distro's package manager, then `go build`.

## 1. Install the dependencies

| Distro | Command |
|---|---|
| Debian / Ubuntu | `sudo apt install libgtk-4-dev gcc pkg-config` |
| Fedora | `sudo dnf install gtk4-devel gcc pkgconf` |
| Arch | `sudo pacman -S gtk4 gcc pkgconf` |

Plus Go from <https://go.dev/dl/> or your package manager.

- **gcc:** gotk4 uses cgo, so a C compiler is needed, just like on Windows.
- **pkg-config:** cgo uses it to find GTK's headers and libraries.
- **GTK 4 development package:** the GTK headers and libraries. It pulls in glib, cairo, pango, etc.
- **gobject-introspection:** on Windows this had to be installed separately.
  On Linux it usually comes in as a dependency of the GTK package; if the build says
  `Package gobject-introspection-1.0 was not found`, install `libgirepository1.0-dev` (Debian/Ubuntu) or `gobject-introspection-devel` (Fedora).

Check:

```sh
pkg-config --modversion gtk4
```

## 2. Build and run

```sh
go build -o todo .
./todo
```

- No `CC`, `PKG_CONFIG` or PATH setup is needed: the system's `gcc` and `pkg-config` are already the right ones.
- No `-ldflags "-H windowsgui"`: that flag is Windows-only (it hides the console window). Linux has no such distinction.
- The **first build is slow** here too (several minutes), because gotk4 compiles a lot of C code. Later builds use Go's cache and take seconds.

## 3. Why there's no DLL problem on Linux

On Windows, `todo.exe` needs ~60 GTK DLLs, and Windows only finds them next to the exe or through `PATH`.
That's why this project has `run.ps1`, the PATH change and the `dist` folder.

On Linux, GTK is installed into the system's standard library folders (e.g. `/usr/lib`), which the dynamic loader always searches.
The binary finds GTK automatically.

Check which shared libraries the binary uses:

```sh
ldd ./todo
```

## 4. Sharing the app with other people

### Users don't need the `-dev` package

The `-dev` / `-devel` package contains headers for **building**.
To **run** the app, users only need the runtime library (`libgtk-4-1` on Debian/Ubuntu, `gtk4` on Fedora/Arch).
Most GNOME-based desktops already have it.

### The GTK version matters

Your binary needs **at least the GTK version it was built against**, and the same goes for glibc.
Built on a new distro (e.g. GTK 4.18), it may fail on an older one (Ubuntu 22.04 ships GTK 4.6) with a missing-symbol error.
The usual fix is to build on the **oldest distro you want to support** (e.g. in a container).

### Ways to ship

| Method | What it means | Notes |
|---|---|---|
| **Flatpak** | The app runs on the GNOME runtime, which provides GTK | The common choice for GTK apps; works on any distro; publish on Flathub |
| **`.deb` / `.rpm`** | Distro packages that declare GTK as a dependency | The package manager installs GTK; one package per distro family |
| **AppImage** | One file with the libraries bundled inside | The Linux equivalent of the Windows `dist` folder; bundling GTK this way is fiddly |
| **Plain binary** | Just `./todo` | Fine for yourself or other developers who have GTK installed |

## 5. Windows vs Linux at a glance

| | Windows | Linux |
|---|---|---|
| Where GTK comes from | msys2 (UCRT64) | distro package manager |
| C compiler | msys2's gcc | system gcc |
| Build command | `.\build.ps1` (sets CC, PKG_CONFIG, PATH) | `go build -o todo .` |
| Finding GTK at runtime | `PATH` or DLLs next to the exe | automatic (system library folders) |
| Shipping | `dist` folder (`package.ps1`) or an installer | Flatpak, `.deb`/`.rpm` or AppImage |

## macOS (for comparison)

macOS is closer to Windows:

- **Development** is easy: `brew install gtk4 pkg-config`, then `go build`.
- **Sharing** needs a `.app` bundle that contains the GTK libraries, similar to the Windows `dist` folder.
