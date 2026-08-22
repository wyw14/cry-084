$ErrorActionPreference = 'Stop'
Get-ChildItem "$PSScriptRoot\..\migrations\*.sql" | Sort-Object Name | ForEach-Object {
  psql $env:DATABASE_URL -v ON_ERROR_STOP=1 -f $_.FullName
}
