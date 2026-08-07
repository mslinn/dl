package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"dl/pkg/config"
	"dl/pkg/downloader"
	"dl/pkg/remote"

	"github.com/lithammer/dedent"

	flag "github.com/spf13/pflag"
)

// Variables set at build time using ldflags
// Example: go build -ldflags "-X main.Version=2.0.0 -X main.Commit=abc123 -X main.BuildDate=2024-01-01"
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

// Args holds command line arguments for the dl tool
type Args struct {
	url        string
	debug      bool
	verbose    bool
	keepVideo  bool
	xrated     bool
	videoDest  string
	configPath string
}

func main() {
	args := parseArgs()

	if err := downloader.CheckYtDlp(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	cfg, err := config.Load(args.configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	destination, purpose, mediaType, format, err := determineDestinationAndMediaType(args, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if args.verbose {
		printDownloadInfo(args.url, format, destination, cfg)
	}

	localPath, err := downloadMedia(args, destination, mediaType, format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully downloaded to: %s\n", localPath)

	if err := copyToRemotes(cfg, localPath, purpose, args.verbose); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: some remote copies failed: %v\n", err)
	}

	fmt.Println("Done!")
}

func determineDestinationAndMediaType(args *Args, cfg *config.Config) (destination string, purpose remote.Purpose, mediaType downloader.MediaType, format string, err error) {
	if args.keepVideo || args.videoDest != "" || args.xrated {
		return determineVideoDestination(args, cfg)
	}

	return determineMP3Destination(cfg)
}

func determineVideoDestination(args *Args, cfg *config.Config) (destination string, purpose remote.Purpose, mediaType downloader.MediaType, format string, err error) {
	mediaType = downloader.Video
	format = "mp4"

	switch {
	case args.xrated:
		destination, err = cfg.GetXDestPath()
		purpose = remote.PurposeXRated
	case args.videoDest != "":
		destination = args.videoDest
	default:
		destination, err = cfg.GetVDestPath()
	}

	return destination, purpose, mediaType, format, err
}

func determineMP3Destination(cfg *config.Config) (destination string, purpose remote.Purpose, mediaType downloader.MediaType, format string, err error) {
	destination, err = cfg.GetMP3sPath()
	purpose = remote.PurposeMP3s
	mediaType = downloader.MP3
	format = "mp3"

	return destination, purpose, mediaType, format, err
}

func downloadMedia(args *Args, destination string, mediaType downloader.MediaType, format string) (string, error) {
	opts := &downloader.Options{
		URL:         args.url,
		Destination: destination,
		MediaType:   mediaType,
		Format:      format,
		Verbose:     args.verbose,
		XRated:      args.xrated,
	}

	dl := downloader.New(opts)
	return dl.Download()
}

func copyToRemotes(cfg *config.Config, localPath string, purpose remote.Purpose, verbose bool) error {
	if len(cfg.ActiveRemotes) == 0 {
		return nil
	}

	copier := remote.New(cfg, verbose)
	return copier.CopyToRemotes(localPath, purpose)
}

func parseArgs() *Args {
	args := &Args{}

	flag.StringVarP(&args.configPath, "config", "c", "~/dl.config", "Path to configuration file")
	flag.BoolVarP(&args.debug, "debug", "d", false, "Enable debug mode (alias for verbose)")
	flag.BoolVarP(&args.keepVideo, "keep-video", "k", false, "Download and keep video")
	flag.BoolVarP(&args.verbose, "verbose", "v", false, "Enable verbose output")
	flag.StringVarP(&args.videoDest, "video-dest", "V", "", "Download video to specified directory")
	flag.BoolVarP(&args.xrated, "xrated", "x", false, "Download x-rated video to xdest")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "dl v%s - Download videos and audio from various websites\n", Version)
		fmt.Fprintf(os.Stderr, dedent.Dedent(`
			Usage: %s [options] URL

			Downloads media from URLs using yt-dlp.
			By default, downloads audio as MP3, unless -k, -x, or -V options are provided.

			Options:
			-c, --config string Path to configuration file (default "~/dl.config")
			-d, --debug Enable debug mode (alias for verbose)
			-h, --help Display this help message
			-k, --keep-video Download and keep video
			-v, --verbose Enable verbose output
			-V, --video-dest string Download video to specified directory
			-x, --xrated Download x-rated video to xdest

			Configuration:
			Edit ~/dl.config to configure local and remote destinations.
			See README-GO.md for configuration details.

			Examples:
			dl https://www.youtube.com/watch?v=dQw4w9WgXcQ
			dl -v https://www.youtube.com/watch?v=dQw4w9WgXcQ
			dl -k https://www.youtube.com/watch?v=dQw4w9WgXcQ
			dl -V ~/Videos https://www.youtube.com/watch?v=dQw4w9WgXcQ
			dl -vV . https://www.youtube.com/watch?v=dQw4w9WgXcQ # Combined flags
		`), filepath.Base(os.Args[0]))
	}

	flag.Parse()

	// Get URL from remaining arguments
	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Error: URL required\n\n")
		flag.Usage()
		os.Exit(1)
	}
	args.url = flag.Arg(0)

	// Debug mode enables verbose
	if args.debug {
		args.verbose = true
	}

	return args
}

func printDownloadInfo(url, format, destination string, cfg *config.Config) {
	fmt.Printf("Downloading from: %s\n", url)
	fmt.Printf("Media type: %s\n", format)
	fmt.Printf("Destination: %s\n", destination)
	if len(cfg.ActiveRemotes) > 0 {
		fmt.Printf("Active remotes: %s\n", strings.Join(cfg.GetActiveRemoteNames(), ", "))
	}
	fmt.Println()
}
