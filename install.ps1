# Install delta from a GitHub release.
#
#   irm https://raw.githubusercontent.com/ronz204/deltas-ai/main/install.ps1 | iex
#
# Environment:
#   DELTA_VERSION      version to install, without the leading v (default: the latest release)
#   DELTA_INSTALL_DIR  where to put the binary (default: %LOCALAPPDATA%\Programs\delta)
#   DELTA_BASE_URL     where the archives and checksums.txt are (default: the GitHub release).
#                      A URL or a local folder. Meant for testing.
#   DELTA_SKIP_PATH    set to 1 to leave the user PATH alone. Meant for testing.

function Install-Delta {
    $ErrorActionPreference = 'Stop'
    $ProgressPreference = 'SilentlyContinue'
    [Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12

    $repo = 'ronz204/deltas-ai'

    switch ($env:PROCESSOR_ARCHITECTURE) {
        'AMD64' { $arch = 'amd64' }
        'ARM64' { $arch = 'arm64' }
        default { throw "install: unsupported architecture $($env:PROCESSOR_ARCHITECTURE)" }
    }

    $version = $env:DELTA_VERSION
    if (-not $version) {
        try {
            $latest = Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest"
            $version = $latest.tag_name
        } catch {
            $version = $null
        }
        if (-not $version) {
            throw "install: no stable release of $repo was found (pre-releases do not count). Set DELTA_VERSION to install one, for example `$env:DELTA_VERSION = '0.1.0-rc1'."
        }
    }
    $version = $version.TrimStart('v')

    $archive = "delta_${version}_windows_${arch}.zip"
    $base = $env:DELTA_BASE_URL
    if (-not $base) { $base = "https://github.com/$repo/releases/download/v$version" }

    $tmp = Join-Path ([IO.Path]::GetTempPath()) ("delta-install-" + [guid]::NewGuid())
    New-Item -ItemType Directory -Path $tmp | Out-Null

    try {
        function Get-File($name) {
            $dest = Join-Path $tmp $name
            if (Test-Path -LiteralPath $base -PathType Container) {
                Copy-Item -LiteralPath (Join-Path $base $name) -Destination $dest
            } else {
                Invoke-WebRequest -Uri "$base/$name" -OutFile $dest -UseBasicParsing
            }
            return $dest
        }

        Write-Host "Installing delta $version for windows/$arch"
        $zip = Get-File $archive
        $sums = Get-File 'checksums.txt'

        $line = Get-Content $sums | Where-Object { $_ -match "\s$([regex]::Escape($archive))$" } | Select-Object -First 1
        if (-not $line) { throw "install: $archive is not listed in checksums.txt" }
        $expected = ($line -split '\s+')[0].ToLower()
        $actual = (Get-FileHash -Algorithm SHA256 -LiteralPath $zip).Hash.ToLower()
        if ($expected -ne $actual) {
            throw "install: checksum mismatch for $archive (expected $expected, got $actual)"
        }

        Expand-Archive -LiteralPath $zip -DestinationPath $tmp -Force

        $dir = $env:DELTA_INSTALL_DIR
        if (-not $dir) { $dir = Join-Path $env:LOCALAPPDATA 'Programs\delta' }
        New-Item -ItemType Directory -Path $dir -Force | Out-Null
        Copy-Item -LiteralPath (Join-Path $tmp 'delta.exe') -Destination (Join-Path $dir 'delta.exe') -Force

        & (Join-Path $dir 'delta.exe') version
        if ($LASTEXITCODE -ne 0) { throw 'install: the installed binary does not run' }
        Write-Host "Installed to $(Join-Path $dir 'delta.exe')"

        if ($env:DELTA_SKIP_PATH -ne '1') {
            $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
            $entries = @($userPath -split ';' | Where-Object { $_ })
            if ($entries -notcontains $dir) {
                [Environment]::SetEnvironmentVariable('Path', (($entries + $dir) -join ';'), 'User')
                Write-Host ''
                Write-Host "Added $dir to your user PATH. Restart open terminals to pick it up."
            }
        }
    }
    finally {
        Remove-Item -LiteralPath $tmp -Recurse -Force -ErrorAction SilentlyContinue
    }
}

Install-Delta
