package main

import (
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func main() {
	if _, err := os.Stat("uploads"); os.IsNotExist(err) {
		os.Mkdir("uploads", os.ModePerm)
	}

	http.HandleFunc("/time", handleDataUpdate)
	http.HandleFunc("/execute", handleExecutionInput)
	http.HandleFunc("/upload", handleUpload)
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	log.Println("🚀 server running at http://localhost:9000")
	log.Fatal(http.ListenAndServe(":9000", nil))
}

func handleDataUpdate(w http.ResponseWriter, r *http.Request) {
	// Get the current time
	localTime := time.Now().Format(time.RFC1123)
	cpuData := "CPU Usage: " + strconv.Itoa(getCpuUsage()) + "% "
	memoryData := "Memory Usage: " + strconv.Itoa(getMemoryUsage()) + "% "

	// Write the current time back to the client
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"time": "%s", "cpuData": "%s", "memoryData": "%s"}`, localTime, cpuData, memoryData)))
}

func handleExecutionInput(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	command := strings.Split(r.FormValue("command"), " ")
	result := ""
	// COMMAND FOR EXECUTION
	cmd1 := exec.Command("ls")
	if len(command) == 1 {
		cmd1 = exec.Command(command[0])
	} else if len(command) == 2 {
		cmd1 = exec.Command(command[0], command[1])
	} else if len(command) == 3 {
		cmd1 = exec.Command(command[0], command[1], command[2])
	}
	// EXECUTING COMMAND AND SAVING OUTPUT
	if output, err := cmd1.Output(); err != nil {
		fmt.Println("Error", "")
	} else {
		result = string(output)
	}
	// REMOVING \n or /t
	response := ""
	for i := 0; i < len(result); i++ {
		//if !unicode.IsSpace(rune(result[i])) {
		if result[i] != '\n' {
			response += string(result[i])
		} else {
			response += "<br>"
		}
	}
	fmt.Println(result)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"response": "%s"}`, response)))
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		w.Write([]byte(`{"error": "Error uploading file"}`))
		return
	}
	defer file.Close()

	filePath := filepath.Join("uploads", header.Filename)
	dst, err := os.Create(filePath)
	if err != nil {
		w.Write([]byte(`{"error": "Error creating file"}`))
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		w.Write([]byte(`{"error": "Error copying file"}`))
		return
	}

	response := fmt.Sprintf("File %s saved successfully!", header.Filename)
	w.Write([]byte(fmt.Sprintf(`{"response": "%s"}`, response)))
}
func getCpuUsage() int {
	percent, err := cpu.Percent(time.Second, false)
	if err != nil {
		log.Fatal(err)
	}
	return int(math.Ceil(percent[0]))
}

func getMemoryUsage() int {
	memory, err := mem.VirtualMemory()
	if err != nil {
		log.Fatal(err)
	}
	return int(math.Ceil(memory.UsedPercent))
}
