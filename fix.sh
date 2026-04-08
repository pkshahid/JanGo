sed -i 's/import (/import (\n\t"strings"/g' sessions/backends/backends.go
sed -i 's/stat, err := os.Stat(filePath)/_, err := os.Stat(filePath)/g' sessions/backends/backends.go
