$ErrorActionPreference = 'Stop'

$packageName = 'dashspace-cli'
$url64 = 'https://github.com/dashspace/cli/releases/download/v1.0.0/dashspace-1.0.0-windows-amd64.zip'
$checksum64 = 'CHECKSUM_TO_UPDATE'
$checksumType64 = 'sha256'

$packageArgs = @{
  packageName   = $packageName
  unzipLocation = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
  url64bit      = $url64
  checksum64    = $checksum64
  checksumType64= $checksumType64
}

Install-ChocolateyZipPackage @packageArgs

# Add to PATH
$installLocation = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
Install-ChocolateyPath "$installLocation"