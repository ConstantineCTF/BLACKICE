param($ip, $port)
$key = 0xC3
$encIP = ""
foreach($c in $ip.ToCharArray()){
    $encIP += [String]::Format("{0:x2}",([byte][char]$c -bxor $key))
}
$encPort = ""
foreach($c in $port.ToCharArray()){
    $encPort += [String]::Format("{0:x2}",([byte][char]$c -bxor $key))
}
Write-Host "    Encrypted!" -ForegroundColor Green
"$encIP|$encPort" | Out-File -Encoding ASCII enc.tmp
