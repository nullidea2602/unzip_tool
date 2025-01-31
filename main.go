package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Config struct {
	ZipDir   string `json:"zip_dir"`
	UnzipDir string `json:"unzip_dir"`
}

const configFile = "config.json"

func loadConfig() (Config, error) {
	config := Config{
		ZipDir:   "./zip_dir",
		UnzipDir: "./unzip_dir",
	}
	file, err := os.Open(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Create default config if missing
			configData, _ := json.MarshalIndent(config, "", "  ")
			os.WriteFile(configFile, configData, 0644)
			return config, nil
		}
		return config, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return config, err
	}
	return config, nil
}

func unzipFile(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(destDir, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		defer outFile.Close()
		
		zipFile, err := f.Open()
		if err != nil {
			return err
		}
		defer zipFile.Close()
		_, err = io.Copy(outFile, zipFile)
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	config, err := loadConfig()
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}
	fmt.Println("Using zip directory:", config.ZipDir)
	fmt.Println("Using unzip directory:", config.UnzipDir)

	if err := os.MkdirAll(config.UnzipDir, os.ModePerm); err != nil {
		fmt.Println("Error creating unzip directory:", err)
		return
	}

	files, err := os.ReadDir(config.ZipDir)
	if err != nil {
		fmt.Println("Error reading zip directory:", err)
		return
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".zip" {
			zipPath := filepath.Join(config.ZipDir, file.Name())
			fmt.Println("Extracting:", zipPath)
			if err := unzipFile(zipPath, config.UnzipDir); err != nil {
				fmt.Println("Error extracting", file.Name(), ":", err)
			} else {
				fmt.Println("Successfully extracted", file.Name())
			}
		}
	}
}
