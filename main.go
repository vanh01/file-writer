package main

import (
	"bufio"
	"fmt"
	"os"
	"sync"
)

// FileWriter represents a file writer service
type FileWriter struct {
	file     *os.File
	writer   *bufio.Writer
	writeCh  chan []byte
	closeCh  chan struct{}
	wg       sync.WaitGroup
	shutdown bool
}

// NewFileWriter creates a new FileWriter
func NewFileWriter(filePath string) (*FileWriter, error) {
	var file *os.File
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		file, err = os.Create(filePath)
		if err != nil {
			return nil, err
		}
	} else {
		file, err = os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
	}

	return &FileWriter{
		file:     file,
		writer:   bufio.NewWriter(file),
		writeCh:  make(chan []byte),
		closeCh:  make(chan struct{}),
		shutdown: false,
	}, nil
}

// Start starts the FileWriter
func (fw *FileWriter) Start() {
	go fw.writeRoutine()
}

// Write writes data to the file
func (fw *FileWriter) Write(data []byte) {
	if !fw.shutdown {
		fw.wg.Add(1)
		fw.writeCh <- data
	}
}

func (fw *FileWriter) WaitForAllWriter() {
	fw.wg.Wait()
}

// Close stops the FileWriter and closes the file
func (fw *FileWriter) Close() {
	if !fw.shutdown {
		close(fw.closeCh)
		fw.wg.Wait()
		fw.file.Close()
		fw.shutdown = true
	}
}

func (fw *FileWriter) writeRoutine() {
	for {
		select {
		case data := <-fw.writeCh:
			fmt.Print("Write " + string(data))
			_, err := fw.writer.Write(data)
			if err != nil {
				fmt.Println("Error writing to file:", err)
			}
			err = fw.writer.Flush()
			fw.wg.Done()
			if err != nil {
				fmt.Println("Error writing to file:", err)
			}
		case <-fw.closeCh:
			return
		}
	}
}

func main() {
	filePath := "output.txt"

	// Create FileWriter
	fileWriter, err := NewFileWriter(filePath)
	if err != nil {
		fmt.Println("Error creating FileWriter:", err)
		return
	}

	// Start FileWriter
	fileWriter.Start()

	// Write data concurrently
	for i := 0; i < 10; i++ {
		data := []byte(fmt.Sprintf("Line %d\n", i))
		go fileWriter.Write(data)
	}

	fileWriter.WaitForAllWriter()
	<-fileWriter.closeCh

	fmt.Println("FileWriter has completed.")
}
