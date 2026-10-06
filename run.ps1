# Starts todo.exe with the msys2 UCRT64 GTK DLLs on PATH.
$msys = if ($env:MSYS2_ROOT) { $env:MSYS2_ROOT } else { "C:\msys64" }
$env:PATH = "$(Join-Path $msys 'ucrt64\bin');$env:PATH"
Start-Process (Join-Path $PSScriptRoot "todo.exe")
