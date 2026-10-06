# Developer docs: GTK 4 desktop development on Windows with Go

The detailed guide for this project. For the short overview, see [README.md](README.md).

The project is a small **todo app** written in **Go** with **GTK 4**, using the [gotk4](https://github.com/diamondburned/gotk4) bindings, built and run on **Windows**.
Todos live only in memory: there's no database and nothing is saved, so closing the app clears the list.

This document explains how GTK development on Windows works: what to install, why each piece is needed, how the build works, and how to ship the app to other computers.

![The todo app](docs/screenshot.png)

---

## Contents

1. [What the app does](#1-what-the-app-does)
2. [How GTK on Windows works (the big picture)](#2-how-gtk-on-windows-works-the-big-picture)
3. [Project files](#3-project-files)
4. [Setting up a Windows machine from scratch](#4-setting-up-a-windows-machine-from-scratch)
5. [PATH: making Windows find msys2's tools and DLLs](#5-path-making-windows-find-msys2s-tools-and-dlls)
6. [gotk4: the Go bindings](#6-gotk4-the-go-bindings)
7. [The scripts: build.ps1, run.ps1, package.ps1 (and make)](#7-the-scripts-buildps1-runps1-packageps1-and-make)
8. [The dist folder: shipping the app to other computers](#8-the-dist-folder-shipping-the-app-to-other-computers)
9. [How the code works (main.go)](#9-how-the-code-works-maingo)
10. [Where everything was installed](#10-where-everything-was-installed)
11. [Troubleshooting](#11-troubleshooting)
12. [Everyday cheat sheet](#12-everyday-cheat-sheet)

---

## 1. What the app does

- Type a task and press **Enter** or click **Add**.
- Tick the **checkbox** to mark a task done; it gets struck through and greyed out.
- Click the **trash** button to delete a task.
- Filter with **All / Active / Done**.
- See **"N items left"**, and remove all finished tasks with **Clear completed**.

---

## 2. How GTK on Windows works (the big picture)

GTK is a **C library**. It isn't written in Go, and it isn't part of Windows.
To use it from Go on Windows you need four things:

```
┌──────────────────────────────────────────────────────────────────────┐
│ main.go  (your Go code)                                              │
│    │ calls                                                           │
│    ▼                                                                 │
│ gotk4    (Go package that wraps GTK's C functions; uses cgo)         │
│    │ cgo compiles C glue code with ──► gcc         (from msys2)      │
│    │ cgo finds GTK headers/libs  with ──► pkg-config (from msys2)    │
│    ▼                                                                 │
│ GTK 4 C library: headers, .dll.a import libs, and .dll files         │
│                                       (all from msys2)               │
└──────────────────────────────────────────────────────────────────────┘
```

| Piece | What it is | Why it's needed | Where it comes from |
|---|---|---|---|
| **Go** | The Go compiler | Compiles `main.go` and gotk4 | [go.dev](https://go.dev/dl/) (was already installed) |
| **gcc** | A C compiler | gotk4 uses **cgo**, so Go has to compile C code, and that needs a C compiler | msys2, `mingw-w64-ucrt-x86_64-gcc` |
| **pkg-config** | A tool that prints compiler and linker flags for a library | cgo runs `pkg-config --cflags --libs gtk4` to find GTK's headers and libraries | msys2, `mingw-w64-ucrt-x86_64-pkgconf` |
| **GTK 4** + its libraries | The GUI toolkit: glib, gobject, gio, cairo, pango, harfbuzz, gdk-pixbuf, … | The thing we actually use | msys2, `mingw-w64-ucrt-x86_64-gtk4` |
| **gobject-introspection** | GTK's metadata library | gotk4 asks pkg-config for `gobject-introspection-1.0`; without it the build fails | msys2, `mingw-w64-ucrt-x86_64-gobject-introspection` |

**Why msys2?** Linux distributions ship GTK through their package manager. Windows has no such thing.
[MSYS2](https://www.msys2.org/) fills that gap: a Windows package manager (`pacman`, borrowed from Arch Linux) with ready-built Windows versions of gcc, GTK and thousands of other libraries.
It's also the approach the GTK project itself recommends for Windows.

**Two phases to keep apart:**

1. **Build time:** Go, gcc and pkg-config compile `main.go` into `todo.exe`. GTK's *headers* and *import libraries* are used here.
2. **Run time:** `todo.exe` is small; the GTK code lives in ~60 **DLLs** (`libgtk-4-1.dll`, `libglib-2.0-0.dll`, …).
   Windows has to **find those DLLs** whenever the app starts, either next to the exe or through `PATH`.
   This is why sections 5 and 8 exist.

---

## 3. Project files

| File | What it is |
|---|---|
| `main.go` | The whole app (UI + logic). See [section 9](#9-how-the-code-works-maingo). |
| `go.mod` / `go.sum` | Go module definition. Pins `github.com/diamondburned/gotk4/pkg v0.4.1`. |
| `build.ps1` | Builds `todo.exe` with msys2's compiler. See [section 7](#7-the-scripts-buildps1-runps1-packageps1). |
| `run.ps1` | Starts `todo.exe` with msys2's DLL folder on `PATH`. |
| `package.ps1` | Builds and bundles the app + GTK runtime into `dist\`. |
| `Makefile` | Optional shortcuts: `mingw32-make build / run / package / clean / help` call the scripts above. See [7.4](#74-calling-the-scripts-via-make). |
| `todo.exe` | Build output (created by `build.ps1`). Ignored by git. |
| `dist\` | Self-contained, shippable copy of the app (created by `package.ps1`). Ignored by git. |
| `.gitignore` | Keeps `todo.exe` and `dist\` out of git: they're large, generated, and can be recreated at any time. |
| `docs\screenshot.png` | Screenshot used in the docs. |
| `MSYS2-ENVIRONMENTS.md` | Reference: msys2's terminals (UCRT64, MINGW64, MSYS, …), and which packages to use for make, cmake, g++, etc. |
| `README.md` | Short overview of the project. |
| `DEV_DOCS.md` | This file: the detailed developer guide. |

Versions used: Go 1.25.1 · gotk4 v0.4.1 · GTK 4.24.1 · gcc 16.2.0 · pkgconf 3.0.7 · gobject-introspection 1.86.0.

---

## 4. Setting up a Windows machine from scratch

Do these steps once per computer.

### 4.1 Install Go

Download from <https://go.dev/dl/> and install. Check in PowerShell:

```powershell
go version
```

### 4.2 Install MSYS2 with the official installer

1. Go to <https://www.msys2.org/> and download `msys2-x86_64-<date>.exe`.
2. Run it and keep the default folder **`C:\msys64`**. The scripts in this repo assume that path; see 4.6 if you choose another.
3. When it finishes, it opens an msys2 terminal.

> **Note:** we first tried installing msys2 through **scoop** (`scoop install msys2`). It worked, but it was
> removed in favour of the official installer, because that's the standard and documented setup and it puts
> msys2 in the well-known `C:\msys64`. Use the official installer.

### 4.3 Update msys2

msys2 has several terminals (MSYS, UCRT64, MINGW64, CLANG64, …); see `MSYS2-ENVIRONMENTS.md`.
**For this project always use "MSYS2 UCRT64"** (Start menu, or `C:\msys64\ucrt64.exe`).

Update everything first:

```sh
pacman -Syu
```

- `-S` = install/sync, `y` = refresh the package lists, `u` = upgrade installed packages.
- The first update often upgrades msys2's core and **closes the terminal**. That's expected:
  **open the UCRT64 terminal again and run `pacman -Syu` a second time** until it says "there is nothing to do".

### 4.4 Install the compiler, pkg-config and GTK 4

In the **MSYS2 UCRT64** terminal:

```sh
pacman -S --needed \
  mingw-w64-ucrt-x86_64-gcc \
  mingw-w64-ucrt-x86_64-pkgconf \
  mingw-w64-ucrt-x86_64-gtk4 \
  mingw-w64-ucrt-x86_64-gobject-introspection
```

- `--needed` skips packages that are already up to date.
- The **`mingw-w64-ucrt-x86_64-`** prefix means "the UCRT64 build of this package". Every package for this project must use that prefix.
  A package **without** a prefix (e.g. plain `gcc`) belongs to the MSYS environment and produces programs that need `msys-2.0.dll`, which you don't want.
- `gtk4` automatically pulls in its dependencies: glib, cairo, pango, harfbuzz, gdk-pixbuf, the Adwaita icon theme, librsvg, …
- **`gobject-introspection` is easy to forget.** Our first build failed without it:
  ```
  Package gobject-introspection-1.0 was not found in the pkg-config search path.
  ```

### 4.5 Check the installation

Still in the UCRT64 terminal:

```sh
which gcc                                              # → /ucrt64/bin/gcc
gcc --version                                          # → gcc.exe (Rev4, Built by MSYS2 project) 16.2.0
pkg-config --modversion gtk4                           # → 4.24.1
pkg-config --modversion gobject-introspection-1.0      # → 1.86.0
```

> **"I don't see gcc in the msys2 terminal!"** You're probably in the plain **MSYS2 MSYS** terminal.
> It only searches `/usr/bin`. gcc lives in `/ucrt64/bin`, which only the **UCRT64** terminal puts on `PATH`.
> On disk it's `C:\msys64\ucrt64\bin\gcc.exe`.

### 4.6 (Optional) msys2 in a different folder

If msys2 isn't in `C:\msys64`, set an environment variable before running the scripts:

```powershell
$env:MSYS2_ROOT = "D:\tools\msys64"
```

All three scripts read `MSYS2_ROOT` and fall back to `C:\msys64`.

### 4.7 Put msys2 on your Windows PATH

See the next section. This is what lets you double-click `todo.exe` and use `gcc` from any terminal.

---

## 5. PATH: making Windows find msys2's tools and DLLs

`PATH` is the list of folders Windows searches for **programs** (`gcc.exe`, `pkg-config.exe`, …) **and for DLLs** when an exe starts.
All of msys2's UCRT64 programs and GTK DLLs are in one folder:

```
C:\msys64\ucrt64\bin
```

### 5.1 What was done on this machine

1. **Removed scoop's `gcc` and `make`** (`scoop uninstall gcc make`).
   scoop's gcc was a *different* MinGW gcc with its own runtime DLLs (`libwinpthread-1.dll`, `libgcc_s_seh-1.dll`, …),
   and its folder was **first** on the user `PATH`. That causes two kinds of trouble:
   - **Compile errors / wrong toolchain:** cgo may pick up scoop's gcc, which can't see msys2's GTK headers and libraries and may disagree with them.
   - **Runtime crashes:** when `todo.exe` starts, Windows may load scoop's same-named runtime DLLs instead of msys2's.

   Removing it also deleted the `C_INCLUDE_PATH` and `CPLUS_INCLUDE_PATH` variables scoop's gcc had set.
2. **Added `C:\msys64\ucrt64\bin` to the end of the user `PATH`.**
   Now `gcc`, `pkg-config` and the GTK DLLs are found from any terminal and from Explorer,
   so you can double-click `todo.exe`.

### 5.2 How to do it yourself (GUI)

1. Press **Win**, type **"environment variables"**, open **"Edit environment variables for your account"**.
2. Select **Path** under *User variables* → **Edit** → **New** → `C:\msys64\ucrt64\bin` → **OK**.
3. **Close and reopen** every terminal and VS Code. Programs only read `PATH` when they start.

Or in PowerShell (doesn't touch other entries; keeps `%VAR%`-style entries intact):

```powershell
$key = [Microsoft.Win32.Registry]::CurrentUser.OpenSubKey('Environment', $true)
$raw = $key.GetValue('Path', '', 'DoNotExpandEnvironmentNames')
$key.SetValue('Path', $raw.TrimEnd(';') + ';C:\msys64\ucrt64\bin', 'ExpandString')
$key.Close()
```

### 5.3 If you keep another gcc (e.g. scoop's)

You don't have to uninstall it, but **order matters: the first match on `PATH` wins.**

- If scoop's gcc folder comes **before** `C:\msys64\ucrt64\bin`, plain `gcc` runs scoop's gcc, and Windows may load scoop's runtime DLLs into GTK apps.
- Our `build.ps1` is safe either way: it puts `C:\msys64\ucrt64\bin` first **for the build only** and sets `CC` and `PKG_CONFIG` to msys2's tools explicitly.
- For running apps, the `dist` folder (section 8) is immune because Windows checks the exe's own folder first.

Check what you'd get:

```powershell
Get-Command gcc, pkg-config | Select Source
```

### 5.4 Never put several msys2 environments on PATH

Use only `C:\msys64\ucrt64\bin`. Adding `C:\msys64\mingw64\bin`, `C:\msys64\usr\bin`, etc. as well mixes incompatible DLLs.

---

## 6. gotk4: the Go bindings

[gotk4](https://github.com/diamondburned/gotk4) is **generated** Go code that wraps GTK's C API using cgo.
Each C function such as `gtk_button_new_with_label()` becomes a Go function such as `gtk.NewButtonWithLabel()`.

### 6.1 How it was added

```powershell
go mod init gotk4-todo
go get github.com/diamondburned/gotk4/pkg@latest   # → v0.4.1
go mod tidy
```

Imports used in `main.go`:

```go
"github.com/diamondburned/gotk4/pkg/gio/v2"   // gio.ApplicationFlagsNone
"github.com/diamondburned/gotk4/pkg/glib/v2"  // glib.MarkupEscapeText
"github.com/diamondburned/gotk4/pkg/gtk/v4"   // all widgets
```

gotk4 is **not** "installed on Windows". It's a normal Go dependency that lives in the Go module cache (section 10).
The only things installed system-wide are msys2 and its packages.

### 6.2 The first build is slow (~5–10 minutes), later builds take seconds

- The `gtk/v4` package is huge: about 120,000 lines of generated Go in `gtk.go` alone.
- Because of cgo, gcc also has to compile a large amount of C glue code, which runs on **one CPU core**.
- glib, gio, gdk, pango, cairo, … are compiled the same way.
- Windows starts processes slowly, and **Windows Defender** scans every temporary file gcc writes.

Go **caches** the compiled packages. You pay this cost again only after upgrading gotk4, Go or gcc, or after `go clean -cache`.

**Speed-ups:**
- Add Defender exclusions for the Go build cache (`go env GOCACHE`, usually `%LOCALAPPDATA%\go-build`) and for `C:\msys64`.
- In a new project, warm the cache once: `go build github.com/diamondburned/gotk4/pkg/gtk/v4`.

### 6.3 Harmless warnings

Every build prints:

```
cgo-generated-wrappers:7:13: warning: conflicting types for built-in function 'free' ...
```

That comes from cgo's own generated code with gcc 16, not from this project. Ignore it.

### 6.4 Finding gotk4 functions

The generated source is the most reliable reference, e.g.
`%USERPROFILE%\go\pkg\mod\github.com\diamondburned\gotk4\pkg@v0.4.1\gtk\v4\gtk.go`.
Search it for the C name (`gtk_list_box_remove`) to find the Go method (`(*ListBox).Remove`).
API docs: <https://pkg.go.dev/github.com/diamondburned/gotk4/pkg/gtk/v4>.
GTK's own docs: <https://docs.gtk.org/gtk4/>.

---

## 7. The scripts: build.ps1, run.ps1, package.ps1 (and make)

All three are PowerShell scripts. Run them from the project folder (or through make, see [7.4](#74-calling-the-scripts-via-make)):

```powershell
cd C:\Users\alber\Desktop\gotk4-test
.\build.ps1
```

If PowerShell refuses to run scripts, allow local scripts once:
`Set-ExecutionPolicy -Scope CurrentUser RemoteSigned`.

### 7.1 `build.ps1`: compile the app

**Why it exists:** building gotk4 requires cgo with **msys2's** gcc and pkg-config.
Typing four environment variables before every `go build` is tedious and error-prone, especially when another gcc (like scoop's) is installed.

**What it does:**

```powershell
$msys = if ($env:MSYS2_ROOT) { $env:MSYS2_ROOT } else { "C:\msys64" }
$ucrt = Join-Path $msys "ucrt64"
$env:PATH        = "$ucrt\bin;$env:PATH"          # msys2 tools + DLLs first (this process only)
$env:CGO_ENABLED = "1"                             # cgo must be on (gotk4 needs it)
$env:CC          = "$ucrt\bin\gcc.exe"             # exactly which C compiler cgo uses
$env:PKG_CONFIG  = "$ucrt\bin\pkg-config.exe"      # exactly which pkg-config cgo uses
go build -ldflags "-H windowsgui" -o todo.exe .
```

- The environment changes only affect this one PowerShell process. Nothing permanent changes.
- `-ldflags "-H windowsgui"` builds a **GUI-subsystem** exe, so **no black console window** opens next to the app.
  Remove it while debugging if you want to see GTK's warnings in a console.

**Usage:**

```powershell
.\build.ps1        # build todo.exe
.\build.ps1 run    # build, then start it
```

`run` starts the app with `Start-Process`, so the script returns immediately instead of waiting for the window to close.
Without that, `mingw32-make run` could hang until you closed the app.

### 7.2 `run.ps1`: start the app with msys2's DLLs available

**Why it exists:** `todo.exe` needs the GTK DLLs from `C:\msys64\ucrt64\bin`.
Before that folder was added to `PATH`, starting `todo.exe` directly failed immediately with exit code
**`0xC0000135` = STATUS_DLL_NOT_FOUND**. `run.ps1` adds the folder to `PATH` just for the launch.

```powershell
$env:PATH = "C:\msys64\ucrt64\bin;$env:PATH"
Start-Process todo.exe
```

**Now optional:** with `C:\msys64\ucrt64\bin` on your user `PATH` (section 5), you can double-click `todo.exe` instead.
It's still useful on a machine where msys2 isn't on `PATH`.

### 7.3 `package.ps1`: build a self-contained `dist\` folder

**Why it exists:** to run the app on computers **without msys2**. See the next section for what it copies and why.

```powershell
.\package.ps1
# → Bundled into ...\dist (61 DLLs, 93.6 MB)
```

It runs `build.ps1` first, **deletes and recreates `dist\`** every time, so run it again after every code change you want to ship.

### 7.4 Calling the scripts via make

The scripts can also be run through **make**, using the `Makefile` in the project root.
The Makefile is only a thin wrapper: every target just calls the matching script, so the build logic lives in one place, the `.ps1` files.

```powershell
mingw32-make build      # = .\build.ps1
mingw32-make run        # = .\build.ps1 run
mingw32-make package    # = .\package.ps1
mingw32-make clean      # delete todo.exe and dist\
mingw32-make help       # list the targets
```

**Which make, and why it's called `mingw32-make`:**
msys2 offers two different make packages (see `MSYS2-ENVIRONMENTS.md`):

| Package | Command | Lives in | Works from |
|---|---|---|---|
| **`mingw-w64-ucrt-x86_64-make`** ← installed | `mingw32-make` | `C:\msys64\ucrt64\bin` | PowerShell, cmd, VS Code, … (that folder is on `PATH`) |
| `make` | `make` | `C:\msys64\usr\bin` | only inside msys2 terminals |

The UCRT64 package was installed because this project is driven from PowerShell and only `C:\msys64\ucrt64\bin` is on `PATH`:

```sh
pacman -S --needed mingw-w64-ucrt-x86_64-make
```

The `32` in `mingw32-make` is a historical name. It's a normal 64-bit GNU Make 4.4.1, named that way so it doesn't clash with the MSYS `make`.

**Want to type just `make`?** Add an alias to your PowerShell profile (`notepad $PROFILE`):

```powershell
Set-Alias make mingw32-make
```

**How the Makefile works:**

```make
PS := powershell -NoProfile -ExecutionPolicy Bypass

build:
	$(PS) -File build.ps1
```

- make doesn't run PowerShell by itself: `mingw32-make` runs recipe lines through `cmd.exe`, or `sh.exe` if one is on `PATH`.
  So each recipe starts `powershell` explicitly. That line works the same under both.
- `-NoProfile` skips your PowerShell profile for faster, predictable runs; `-ExecutionPolicy Bypass` lets the scripts run even if script execution is restricted.
- All targets are `.PHONY`: they're commands, not files, so make always runs them.
- `help` prints the `## ` comment lines from the Makefile.
- Recipe lines **must start with a TAB**, not spaces, or make fails with `missing separator`.
- **Never end a Makefile line with `\`**, not even in a comment: make treats it as "continues on the next line".
  A comment ending in `dist\` once silently swallowed the `clean:` line below it.

**Is make needed?** No. It's a convenience. `.\build.ps1` and `mingw32-make build` do exactly the same thing.

---

## 8. The dist folder: shipping the app to other computers

### 8.1 Is it required?

**No, not for development on your own machine.** With msys2 on `PATH`, `todo.exe` runs fine.
You need `dist\` when:

- **Giving the app to someone else.** They don't have msys2, so they'd get "DLL not found".
- **Insulating it from msys2 updates.** `pacman -Syu` replaces DLLs in place; the bundle keeps the versions you tested.
- **Avoiding DLL clashes.** Windows looks in the **exe's own folder first**, before `PATH`, so a bundled app can't pick up a wrong same-named DLL from another program.

Static linking isn't a practical alternative for GTK on Windows: GTK and its many libraries are designed to be used as DLLs.
An installer (Inno Setup, NSIS, WiX) is just a wrapper around a folder like `dist\`.

### 8.2 What ends up in dist and why

```
dist\
├── todo.exe
├── libgtk-4-1.dll, libglib-2.0-0.dll, libcairo-2.dll, …     (61 DLLs)
├── lib\gdk-pixbuf-2.0\2.10.0\
│   ├── loaders\*.dll       ← image-format plugins (PNG, JPEG, SVG, …)
│   └── loaders.cache       ← list of those plugins (paths are relative, so it's copied as-is)
└── share\
    ├── icons\Adwaita, AdwaitaLegacy, hicolor   ← icon themes (trash icon, window buttons, …)
    └── glib-2.0\schemas\gschemas.compiled      ← GTK/GLib settings definitions
```

How `package.ps1` assembles it:

1. **Linked DLLs:** it runs msys2's `ldd` (`C:\msys64\usr\bin\ldd.exe`) on `todo.exe`.
   `ldd` lists **every** DLL the exe loads, including indirect ones, e.g.
   `libgtk-4-1.dll => /ucrt64/bin/libgtk-4-1.dll`. Every DLL under `/ucrt64/bin` is copied next to `todo.exe`.
   Windows system DLLs (`KERNEL32.dll`, `ucrtbase.dll`, …) are skipped because every Windows has them.
2. **Plugin DLLs:** gdk-pixbuf's image loaders are **loaded at runtime, not linked**, so `ldd todo.exe` can't see them.
   They're copied explicitly, then `ldd` runs on them too, because e.g. the SVG loader needs `librsvg` and its dependencies.
3. **Data files:** icon themes and compiled settings schemas aren't code. GTK loads them by path at runtime, so they're copied as folders.

**How GTK finds `share\` and `lib\` at runtime:** on Windows, GTK and GLib work out their *install prefix* from the location of their own DLL.
If the DLL sits in a folder called `bin`, the prefix is its parent; otherwise, as in our flat `dist\`, it's the DLL's own folder.
Then GTK looks for `<prefix>\share\icons`, `<prefix>\lib\gdk-pixbuf-2.0\…`, and so on.
**That's why the folder structure inside `dist\` must stay exactly as it is.**

### 8.3 Verified

`dist\todo.exe` was started with **every msys2 folder removed from `PATH`**. The window rendered correctly, Adwaita icons included.
GTK only logged harmless messages: a D-Bus warning that's normal on Windows, and Vulkan info about an OBS capture hook.

### 8.4 Shipping

Zip the **whole** `dist` folder. The receiver unzips it anywhere and runs `todo.exe`. No installation, no msys2.

---

## 9. How the code works (main.go)

### 9.1 Startup

```go
app := gtk.NewApplication(appID, gio.ApplicationFlagsNone)
app.ConnectActivate(func() { newTodoApp(app).window.Present() })
app.Run(os.Args)
```

- A `gtk.Application` owns the main loop. `Run` blocks until the last window closes.
- `appID` (`com.example.gotk4todo`) is a unique reverse-DNS name GTK uses to identify the app.
- The **activate** signal fires when the app starts, and that's where the window is built.

### 9.2 Data model

```go
type todo struct {
    text  string
    done  bool
    row   *gtk.ListBoxRow   // the row showing this todo
    label *gtk.Label        // its text label
}
type todoApp struct {
    window  *gtk.ApplicationWindow
    entry   *gtk.Entry
    list    *gtk.ListBox
    counter *gtk.Label
    todos   []*todo          // in-memory list; lost on exit
    filter  filter           // All / Active / Done
}
```

The `todos` slice is the source of truth. Each todo keeps pointers to its widgets so the UI can be updated.

### 9.3 Widget tree

```
ApplicationWindow "Todo"
└── Box (vertical)
    ├── Box (horizontal)                  ← input row
    │   ├── Entry  "What needs to be done?"
    │   └── Button "Add"  (.suggested-action = blue)
    ├── ScrolledWindow
    │   └── ListBox (.boxed-list, placeholder "Nothing to show")
    │       └── ListBoxRow  (one per todo)
    │           └── Box (horizontal)
    │               ├── CheckButton
    │               ├── Label  (task text)
    │               └── Button (user-trash-symbolic icon, .flat)
    └── Box (horizontal)                  ← footer
        ├── Label "N items left"  (.dim-label)
        ├── Box (.linked) [All][Active][Done]  ← grouped ToggleButtons
        └── Button "Clear completed"
```

GTK 4 layout basics used here:
- `gtk.Box` stacks children vertically or horizontally; `Append` adds a child.
- `SetHExpand(true)` / `SetVExpand(true)` lets a widget take the leftover space.
- `AddCSSClass("…")` applies GTK theme styles: `suggested-action`, `flat`, `dim-label`, `linked`, `boxed-list`.

### 9.4 Signals (events)

GTK calls your Go functions through **signals**, connected with `Connect…` methods:

| Signal | Code | Action |
|---|---|---|
| Entry **activate** (Enter pressed) | `t.entry.ConnectActivate(t.addFromEntry)` | add a todo |
| Add **clicked** | `addButton.ConnectClicked(t.addFromEntry)` | add a todo |
| CheckButton **toggled** | `check.ConnectToggled(...)` | flip `done`, refresh |
| Trash **clicked** | `deleteButton.ConnectClicked(...)` | `remove(item)` |
| Filter ToggleButton **toggled** | `btn.ConnectToggled(...)` | set filter, refresh |
| Clear completed **clicked** | `clearButton.ConnectClicked(t.clearCompleted)` | remove done todos |

The filter buttons behave like radio buttons because of `btn.SetGroup(first)`: only one can be active at a time.

### 9.5 The functions

- **`addFromEntry`**: trims the entry text, ignores empty input, clears the entry, calls `add`.
- **`add`**: builds a row (checkbox, label, trash button), appends it to the `ListBox` and the `todos` slice, refreshes.
- **`remove`**: removes the todo from the slice and its row from the `ListBox`.
- **`clearCompleted`**: removes every done todo, from both the slice and the UI.
- **`refresh`**: the single place that syncs the UI with the data:
  - done tasks get `<s>…</s>` (strikethrough) markup and the `dim-label` style;
  - rows are shown or hidden according to the filter (`SetVisible`);
  - the "N items left" counter is updated.

  The text goes through `glib.MarkupEscapeText` first, so typing `<b>` or `&` doesn't break the markup.
  When every row is hidden or the list is empty, the ListBox shows its **placeholder** "Nothing to show".

### 9.6 Adding persistence later

Persistence was intentionally left out. To add it later, load the `todos` on startup (calling `add` for each) and save them in `add`, `remove`, `clearCompleted` and the checkbox handler.
JSON in `%APPDATA%`, SQLite or anything else would work.

---

## 10. Where everything was installed

| What | Location | How to remove |
|---|---|---|
| MSYS2 (+ gcc, pkgconf, GTK 4, gobject-introspection, make, …) | `C:\msys64` | Windows "Installed apps" → MSYS2 → Uninstall |
| msys2 on user `PATH` | user `Path` variable, entry `C:\msys64\ucrt64\bin` | remove the entry (section 5.2) |
| Go module cache (gotk4 source) | `%USERPROFILE%\go\pkg\mod` | `go clean -modcache` |
| Go build cache (compiled gotk4) | `%LOCALAPPDATA%\go-build` (`go env GOCACHE`) | `go clean -cache` |
| This project | this folder | delete the folder |

**Nothing** was installed into `C:\Windows` or other system folders. GTK and its libraries exist only under `C:\msys64` and, after packaging, as copies inside `dist\`.

---

## 11. Troubleshooting

| Symptom | Cause | Fix |
|---|---|---|
| `todo.exe` does nothing / closes instantly; exit code `0xC0000135`; or a "…dll was not found" dialog | Windows can't find the GTK DLLs | Add `C:\msys64\ucrt64\bin` to `PATH` (section 5), use `.\run.ps1`, or run `dist\todo.exe` |
| `Package gobject-introspection-1.0 was not found in the pkg-config search path` | Package missing | `pacman -S mingw-w64-ucrt-x86_64-gobject-introspection` |
| `Package gtk4 was not found …` | GTK missing, or the wrong pkg-config is used | Install `mingw-w64-ucrt-x86_64-gtk4`; build with `build.ps1` (it sets `PKG_CONFIG`) |
| `cgo: C compiler "gcc" not found` | No gcc on `PATH` | Use `build.ps1` (sets `CC`), or put `C:\msys64\ucrt64\bin` on `PATH` |
| Strange compile/link errors, `undefined reference`, header errors | A different gcc (scoop, Chocolatey, TDM, …) is being used | `Get-Command gcc` should point to `C:\msys64\ucrt64\bin\gcc.exe`; use `build.ps1` |
| App crashes at startup on your machine but `dist\` works | Same-named DLLs from another program earlier on `PATH` | Make sure no other MinGW `bin` folder precedes `C:\msys64\ucrt64\bin` |
| `gcc` not found in the msys2 terminal | You're in the **MSYS** terminal | Open **MSYS2 UCRT64** |
| `pacman -Syu` closed the terminal | Normal core update | Reopen the terminal and run `pacman -Syu` again |
| First `go build` takes 5–10 minutes | gotk4 + cgo compile (section 6.2) | Wait once; later builds use the cache. Add Defender exclusions. |
| `warning: conflicting types for built-in function 'free'` | cgo + gcc 16 | Harmless, ignore |
| Icons missing (empty squares) in `dist\` | `share\icons` or the pixbuf loaders missing or moved | Re-run `.\package.ps1`; don't rearrange `dist\` |
| Scripts won't run ("running scripts is disabled") | PowerShell execution policy | `Set-ExecutionPolicy -Scope CurrentUser RemoteSigned` |
| Want to see GTK warnings | GUI-subsystem build hides the console | Remove `-ldflags "-H windowsgui"` in `build.ps1`, rebuild, run from a terminal |

---

## 12. Everyday cheat sheet

```powershell
# Develop
.\build.ps1 run          # build + start
.\todo.exe               # start (msys2 is on PATH)

# Ship
.\package.ps1            # → dist\ ; zip it and share

# Same via make
mingw32-make run | build | package | clean | help
```

```sh
# In the "MSYS2 UCRT64" terminal
pacman -Syu                                  # update everything
pacman -Ss <name>                            # search packages
pacman -S --needed mingw-w64-ucrt-x86_64-<name>   # install a UCRT64 package
pacman -Qo /ucrt64/bin/libgtk-4-1.dll        # which package owns a file
```

Further reading:
- `MSYS2-ENVIRONMENTS.md` in this repo: which msys2 terminal to use, plus make, cmake, g++, gdb, …
- MSYS2: <https://www.msys2.org/> · GTK on Windows: <https://www.gtk.org/docs/installations/windows/>
- gotk4: <https://github.com/diamondburned/gotk4> · examples: <https://github.com/diamondburned/gotk4-examples>
- GTK 4 API: <https://docs.gtk.org/gtk4/>
