package progress_test

import (
	"fmt"
	"os"
	"time"

	"github.com/jmylchreest/tinct/internal/ui/progress"
)

// ExampleSpinner demonstrates a spinner for long-running operations.
func ExampleSpinner() {
	spinner := progress.NewSpinner("Processing...").WithWriter(os.Stdout)
	spinner.Start()

	// Simulate work
	time.Sleep(2 * time.Second)

	spinner.Stop("Processing complete ✓")
	// Output: Processing complete ✓
}

// ExampleSpinner_update demonstrates updating spinner message.
func ExampleSpinner_update() {
	spinner := progress.NewSpinner("Step 1...")
	spinner.Start()

	time.Sleep(500 * time.Millisecond)
	spinner.UpdateMessage("Step 2...")

	time.Sleep(500 * time.Millisecond)
	spinner.UpdateMessage("Step 3...")

	time.Sleep(500 * time.Millisecond)
	spinner.Stop("All steps complete ✓")
}

// ExampleProgressBar demonstrates a progress bar for downloads.
func ExampleProgressBar() {
	total := int64(1024 * 1024 * 10) // 10 MB
	bar := progress.NewProgressBar(total, "Downloading plugin").WithWriter(os.Stdout)

	// Simulate download in chunks
	for downloaded := int64(0); downloaded < total; downloaded += 1024 * 512 {
		bar.Set(downloaded)
		time.Sleep(50 * time.Millisecond)
	}

	bar.Finish("Download complete ✓")
	// Output: Download complete ✓
}

// ExampleStatus demonstrates a simple status line.
func ExampleStatus() {
	status := progress.NewStatus("Checking repositories...").WithWriter(os.Stdout)

	time.Sleep(500 * time.Millisecond)
	status.Update("Downloading manifest...")

	time.Sleep(500 * time.Millisecond)
	status.Update("Verifying checksums...")

	time.Sleep(500 * time.Millisecond)
	status.Finish("Repository sync complete ✓")
	// Output: Repository sync complete ✓
}

// Example showing combined usage in a realistic scenario.
func Example_combined() {
	fmt.Println("Installing plugins...")

	// Step 1: Checking with spinner
	spinner := progress.NewSpinner("Checking plugin repository...").WithWriter(os.Stdout)
	spinner.Start()
	time.Sleep(time.Second)
	spinner.Stop("Repository checked ✓")

	// Step 2: Download with progress bar
	bar := progress.NewProgressBar(1024*1024*5, "Downloading plugin").WithWriter(os.Stdout)
	for i := int64(0); i < 1024*1024*5; i += 1024 * 256 {
		bar.Set(i)
		time.Sleep(10 * time.Millisecond)
	}
	bar.Finish("Plugin downloaded ✓")

	// Step 3: Installing with spinner
	spinner2 := progress.NewSpinner("Installing plugin...").WithWriter(os.Stdout)
	spinner2.Start()
	time.Sleep(time.Second)
	spinner2.Stop("Installation complete ✓")

	fmt.Println("\nAll done!")
	// Output:
	// Installing plugins...
	// Repository checked ✓
	// Plugin downloaded ✓
	// Installation complete ✓
	//
	// All done!
}
