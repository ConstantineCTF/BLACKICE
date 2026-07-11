package main
import ("net/http"; "log")
func main() {
    log.Println("Server running on http://0.0.0.0:8000")
    log.Fatal(http.ListenAndServe(":8000", http.FileServer(http.Dir("."))))
}