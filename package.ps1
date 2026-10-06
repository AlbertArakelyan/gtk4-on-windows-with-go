# Builds todo.exe and bundles it with the GTK 4 runtime from msys2 (UCRT64)
# into .\dist, so it runs on Windows machines without msys2 installed.
$ErrorActionPreference = "Stop"

$msys = if ($env:MSYS2_ROOT) { $env:MSYS2_ROOT } else { "C:\msys64" }
$ucrt = Join-Path $msys "ucrt64"
$ldd = Join-Path $msys "usr\bin\ldd.exe"
$dist = Join-Path $PSScriptRoot "dist"

& (Join-Path $PSScriptRoot "build.ps1")
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

if (Test-Path $dist) { Remove-Item $dist -Recurse -Force }
New-Item -ItemType Directory $dist | Out-Null
Copy-Item (Join-Path $PSScriptRoot "todo.exe") $dist

# Copies every UCRT64 DLL that the given binaries load (ldd resolves them transitively).
function Copy-Dependencies([string[]]$binaries) {
    $env:PATH = "$ucrt\bin;$env:PATH"
    foreach ($bin in $binaries) {
        foreach ($line in & $ldd $bin) {
            if ($line -match "=> /ucrt64/bin/(\S+\.dll)") {
                $target = Join-Path $dist $matches[1]
                if (-not (Test-Path $target)) { Copy-Item (Join-Path "$ucrt\bin" $matches[1]) $target }
            }
        }
    }
}

# gdk-pixbuf image loaders (PNG, SVG, ...) are loaded at runtime, not linked,
# so ldd on todo.exe doesn't see them. loaders.cache uses paths relative to dist.
$loaders = "lib\gdk-pixbuf-2.0\2.10.0"
New-Item -ItemType Directory (Join-Path $dist "$loaders\loaders") | Out-Null
Copy-Item "$ucrt\$loaders\loaders\*.dll" (Join-Path $dist "$loaders\loaders")
Copy-Item "$ucrt\$loaders\loaders.cache" (Join-Path $dist $loaders)

Copy-Dependencies (@(Join-Path $dist "todo.exe") + (Get-ChildItem (Join-Path $dist "$loaders\loaders\*.dll")).FullName)

# Icon themes (for icons such as user-trash-symbolic) and compiled GSettings schemas.
New-Item -ItemType Directory (Join-Path $dist "share\icons"), (Join-Path $dist "share\glib-2.0\schemas") | Out-Null
foreach ($theme in "Adwaita", "AdwaitaLegacy", "hicolor") {
    Copy-Item "$ucrt\share\icons\$theme" (Join-Path $dist "share\icons") -Recurse
}
Copy-Item "$ucrt\share\glib-2.0\schemas\gschemas.compiled" (Join-Path $dist "share\glib-2.0\schemas")

$size = (Get-ChildItem $dist -Recurse -File | Measure-Object Length -Sum).Sum / 1MB
$dlls = (Get-ChildItem $dist -Filter *.dll).Count
"Bundled into $dist ($dlls DLLs, {0:N1} MB)" -f $size
