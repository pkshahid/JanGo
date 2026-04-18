package management

import (
	"context"
	"fmt"
	"github.com/pkshahid/JanGo/core/settings"
	"github.com/pkshahid/JanGo/management"
	"github.com/pkshahid/JanGo/static"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path/filepath"
	"sync"
)

func init() {
	management.Register(&CollectStaticCommand{})
}

type CollectStaticCommand struct {
	noInput bool
	clear   bool
	dryRun  bool
	link    bool
}

func (c *CollectStaticCommand) Help() string {
	return "Collects static files into STATIC_ROOT."
}

func (c *CollectStaticCommand) Name() string {
	return "collectstatic"
}

func (c *CollectStaticCommand) AddFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&c.noInput, "noinput", false, "Do NOT prompt the user for input of any kind.")
	cmd.Flags().BoolVar(&c.clear, "clear", false, "Clear the existing files using the storage before trying to copy or link the original file.")
	cmd.Flags().BoolVar(&c.dryRun, "dry-run", false, "Do everything except modify the filesystem.")
	cmd.Flags().BoolVar(&c.link, "link", false, "Create a symbolic link to each file instead of copying.")
}

func (c *CollectStaticCommand) Execute(ctx context.Context, args []string) error {
	// arguments are handled by cobra flags

	s := settings.Get()
	if s.STATIC_ROOT == "" {
		fmt.Println("Error: STATIC_ROOT is not set in settings")
		return fmt.Errorf("STATIC_ROOT is not set in settings")
	}

	storage := static.GetStorage()

	if c.clear && !c.dryRun {
		fmt.Printf("Clearing %s...\n", s.STATIC_ROOT)
		os.RemoveAll(s.STATIC_ROOT)
		os.MkdirAll(s.STATIC_ROOT, 0755)
	}

	if !c.dryRun {
		os.MkdirAll(s.STATIC_ROOT, 0755)
	}

	// 1. Gather all files from all finders
	finders := static.GetFinders()
	fileMap := make(map[string]string) // relative path -> absolute source path

	for _, finder := range finders {
		files, err := finder.List()
		if err != nil {
			fmt.Printf("Error listing files in finder: %v\n", err)
			continue
		}
		for _, name := range files {
			if _, exists := fileMap[name]; !exists {
				absPath, err := finder.Find(name)
				if err == nil {
					fileMap[name] = absPath
				}
			}
		}
	}

	// 2. Process files concurrently
	fmt.Printf("Found %d static files to process.\n", len(fileMap))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 20) // limit concurrency

	var copiedCount int
	var skippedCount int
	var mu sync.Mutex

	for name, srcPath := range fileMap {
		wg.Add(1)
		go func(name, srcPath string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			destPath := storage.Path(name)

			// Check if file is unchanged (we can compare mtime or hash, let's compare size and mtime for speed, fallback to hash if needed)
			srcInfo, err := os.Stat(srcPath)
			if err != nil {
				return
			}

			destInfo, err := os.Stat(destPath)
			if err == nil && srcInfo.Size() == destInfo.Size() && srcInfo.ModTime().Equal(destInfo.ModTime()) {
				// File is probably unchanged
				mu.Lock()
				skippedCount++
				mu.Unlock()
				return
			}

			if !c.dryRun {
				err = c.copyFile(srcPath, destPath, c.link)
				if err != nil {
					fmt.Printf("Error copying %s: %v\n", name, err)
					return
				}
				// Set mod time to match
				os.Chtimes(destPath, srcInfo.ModTime(), srcInfo.ModTime())
			}

			mu.Lock()
			copiedCount++
			mu.Unlock()
			fmt.Printf("Copied %s\n", name)

		}(name, srcPath)
	}

	wg.Wait()

	// 3. If using ManifestStaticFilesStorage, trigger PostProcess
	if manifestStorage, ok := storage.(*static.ManifestStaticFilesStorage); ok && !c.dryRun {
		fmt.Println("Post-processing (hashing) files...")
		err := manifestStorage.PostProcess()
		if err != nil {
			fmt.Printf("Error during post-processing: %v\n", err)
		}
	}

	if c.dryRun {
		fmt.Printf("\nDry run complete. Would copy: %d, Skip: %d\n", copiedCount, skippedCount)
	} else {
		fmt.Printf("\nCollectstatic complete. Copied: %d, Skipped: %d\n", copiedCount, skippedCount)
	}

	return nil
}

func (c *CollectStaticCommand) copyFile(src, dst string, link bool) error {
	dir := filepath.Dir(dst)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if link {
		os.Remove(dst) // Remove if exists
		return os.Link(src, dst)
	}

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
	return err
}
