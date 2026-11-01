[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [ValidateSet('focus', 'install', 'test', 'vet', 'docs', 'build', 'cas-observe', 'cas-provenance', 'cas-binding', 'cas-install', 'cas-collect', 'cas-acquire', 'trust-audit', 'ctl-check', 'ctl-identifiers', 'ctl-member', 'auth-dev')]
    [string]$Stage,
    [ValidateSet('unspecified', 'desktop-native', 'tool')]
    [string]$Context = 'unspecified'
)
$ErrorActionPreference = 'Stop'
& python (Join-Path $PSScriptRoot 'dev.py') $Stage --context $Context
exit $LASTEXITCODE
