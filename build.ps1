# Builds the todo app against GTK 4 from msys2 (UCRT64).
# Override the msys2 location with $env:MSYS2_ROOT if it isn't installed in C:\msys64.
$msys = if ($env:MSYS2_ROOT) { $env:MSYS2_ROOT } else { "C:\msys64" }
$ucrt = Join-Path $msys "ucrt64"
$env:PATH = "$ucrt\bin;$env:PATH"
$env:CGO_ENABLED = "1"
$env:CC = "$ucrt\bin\gcc.exe"
$env:PKG_CONFIG = "$ucrt\bin\pkg-config.exe"

# -H windowsgui hides the console window; drop it to see GTK warnings.
go build -ldflags "-H windowsgui" -o todo.exe .
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

# Start-Process returns immediately instead of waiting on the GUI app (e.g. under make).
if ($args -contains "run") { Start-Process (Join-Path $PSScriptRoot "todo.exe") }
