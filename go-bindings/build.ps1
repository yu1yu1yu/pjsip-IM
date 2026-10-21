[CmdletBinding()]
param([switch]$Test)
$ErrorActionPreference = 'Stop'
$taskRoot = $PSScriptRoot
foreach ($tool in @('go', 'gcc', 'g++', 'cmake', 'ninja')) {
    if (-not (Get-Command $tool -ErrorAction SilentlyContinue)) {
        throw "Missing tool: $tool. See go-bindings/README.md."
    }
}
function Invoke-Checked {
    param([string]$Program, [string[]]$Arguments)
    & $Program @Arguments
    if ($LASTEXITCODE -ne 0) { throw "$Program failed (exit $LASTEXITCODE)" }
}
Push-Location $taskRoot
try {
    # Keep tool caches local; do not modify global Go configuration.
    $previousCache = $env:GOCACHE
    $previousCgo = $env:CGO_ENABLED
    $previousCC = $env:CC
    $env:GOCACHE = Join-Path $taskRoot '.cache/go-build'
    $env:CGO_ENABLED = '1'
    $env:CC = (Get-Command gcc).Source
    Invoke-Checked cmake @('-S', '.', '-B', 'build', '-G', 'Ninja',
        '-DCMAKE_BUILD_TYPE=Release', '-DCMAKE_C_COMPILER=gcc', '-DCMAKE_CXX_COMPILER=g++')
    Invoke-Checked cmake @('--build', 'build', '--target', 'pjgo', '--parallel', '3')
    Invoke-Checked go @('build', '-o', 'build/bin/sip-im.exe', './cmd/sip-im')
    if ($Test) {
        $previousPath = $env:PATH
        try {
            $env:PATH = "$(Join-Path $taskRoot 'build/bin');$env:PATH"
            Invoke-Checked go @('test', '-count=1', './...')
            Invoke-Checked go @('vet', './...')
        } finally { $env:PATH = $previousPath }
    }
    Write-Host "Built: $taskRoot/build/bin/sip-im.exe"
} finally {
    $env:GOCACHE = $previousCache
    $env:CGO_ENABLED = $previousCgo
    $env:CC = $previousCC
    Pop-Location
}
