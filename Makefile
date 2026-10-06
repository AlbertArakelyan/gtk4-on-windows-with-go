# Thin wrapper around the PowerShell scripts, which hold the actual build logic.
# Use with msys2's UCRT64 make:  mingw32-make <target>
# (installed by: pacman -S mingw-w64-ucrt-x86_64-make)

PS := powershell -NoProfile -ExecutionPolicy Bypass

.PHONY: build run package clean help

## build:   compile todo.exe (build.ps1)
build:
	$(PS) -File build.ps1

## run:     build, then start the app (build.ps1 run)
run:
	$(PS) -File build.ps1 run

## package: build and bundle the app with GTK into dist\ (package.ps1)
package:
	$(PS) -File package.ps1

## clean:   delete todo.exe and the dist folder
clean:
	$(PS) -Command "Remove-Item -Recurse -Force -ErrorAction SilentlyContinue todo.exe, dist"

## help:    list the targets
help:
	@$(PS) -Command "Select-String -Path Makefile -Pattern '^## ' | ForEach-Object { $$_.Line.Substring(3) }"
