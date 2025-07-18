$ErrorActionPreference = 'Stop'

$packageName = 'dashspace-cli'
$installLocation = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"

# Remove from PATH
Uninstall-ChocolateyPath "$installLocation"

# Remove binary
Remove-Item "$installLocation\dashspace.exe" -Force -ErrorAction SilentlyContinue