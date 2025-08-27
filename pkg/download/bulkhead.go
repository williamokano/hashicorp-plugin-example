package download

import (
	"context"
	"sync"
)

// Bulkhead provides concurrency control for downloads
type Bulkhead struct {
	semaphore chan struct{}
	wg        sync.WaitGroup
}

// NewBulkhead creates a new bulkhead with specified max concurrent operations
func NewBulkhead(maxConcurrent int) *Bulkhead {
	if maxConcurrent <= 0 {
		maxConcurrent = 1
	}
	
	return &Bulkhead{
		semaphore: make(chan struct{}, maxConcurrent),
	}
}

// Execute runs a function with concurrency control
func (b *Bulkhead) Execute(fn func() error) error {
	// Acquire semaphore
	b.semaphore <- struct{}{}
	defer func() { <-b.semaphore }()
	
	return fn()
}

// ExecuteAsync runs a function asynchronously with concurrency control
func (b *Bulkhead) ExecuteAsync(fn func() error) {
	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		b.Execute(fn)
	}()
}

// Wait waits for all async operations to complete
func (b *Bulkhead) Wait() {
	b.wg.Wait()
}

// BulkheadResult represents the result of an async operation
type BulkheadResult struct {
	Name  string
	Error error
}

// ParallelDownloader manages parallel plugin downloads
type ParallelDownloader struct {
	bulkhead      *Bulkhead
	maxConcurrent int
	results       chan BulkheadResult
	ctx           context.Context
}

// NewParallelDownloader creates a new parallel downloader
func NewParallelDownloader(maxConcurrent int) *ParallelDownloader {
	if maxConcurrent <= 0 {
		maxConcurrent = 4 // Default to 4 concurrent downloads
	}
	
	return &ParallelDownloader{
		bulkhead:      NewBulkhead(maxConcurrent),
		maxConcurrent: maxConcurrent,
		results:       make(chan BulkheadResult, 100),
		ctx:           context.Background(),
	}
}

// NewParallelDownloaderWithContext creates a new parallel downloader with context
func NewParallelDownloaderWithContext(ctx context.Context, maxConcurrent int) *ParallelDownloader {
	pd := NewParallelDownloader(maxConcurrent)
	pd.ctx = ctx
	return pd
}

// DownloadItem represents an item to download
type DownloadItem struct {
	Name     string
	Version  string
	URL      string
	DestPath string
}

// DownloadAll downloads multiple items in parallel
func (pd *ParallelDownloader) DownloadAll(items []DownloadItem, downloadFunc func(DownloadItem) error) ([]BulkheadResult, error) {
	var wg sync.WaitGroup
	results := make([]BulkheadResult, 0, len(items))
	resultChan := make(chan BulkheadResult, len(items))
	
	// Start downloads
	for _, item := range items {
		// Check context
		select {
		case <-pd.ctx.Done():
			return results, pd.ctx.Err()
		default:
		}
		
		wg.Add(1)
		go func(item DownloadItem) {
			defer wg.Done()
			
			err := pd.bulkhead.Execute(func() error {
				// Check context before download
				select {
				case <-pd.ctx.Done():
					return pd.ctx.Err()
				default:
				}
				
				return downloadFunc(item)
			})
			
			resultChan <- BulkheadResult{
				Name:  item.Name,
				Error: err,
			}
		}(item)
	}
	
	// Wait for all downloads to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()
	
	// Collect results
	for result := range resultChan {
		results = append(results, result)
	}
	
	return results, nil
}

// DownloadQueue manages a queue of downloads with progress reporting
type DownloadQueue struct {
	downloader *ParallelDownloader
	items      []DownloadItem
	onProgress func(completed, total int, current string)
	onError    func(name string, err error)
}

// NewDownloadQueue creates a new download queue
func NewDownloadQueue(maxConcurrent int) *DownloadQueue {
	return &DownloadQueue{
		downloader: NewParallelDownloader(maxConcurrent),
		items:      []DownloadItem{},
	}
}

// Add adds an item to the download queue
func (q *DownloadQueue) Add(item DownloadItem) {
	q.items = append(q.items, item)
}

// SetProgressCallback sets the progress callback
func (q *DownloadQueue) SetProgressCallback(fn func(completed, total int, current string)) {
	q.onProgress = fn
}

// SetErrorCallback sets the error callback
func (q *DownloadQueue) SetErrorCallback(fn func(name string, err error)) {
	q.onError = fn
}

// Execute executes all downloads in the queue
func (q *DownloadQueue) Execute(downloadFunc func(DownloadItem) error) error {
	if len(q.items) == 0 {
		return nil
	}
	
	var wg sync.WaitGroup
	completed := 0
	total := len(q.items)
	var mu sync.Mutex
	
	for _, item := range q.items {
		wg.Add(1)
		go func(item DownloadItem) {
			defer wg.Done()
			
			// Report progress
			mu.Lock()
			if q.onProgress != nil {
				q.onProgress(completed, total, item.Name)
			}
			mu.Unlock()
			
			// Execute download
			err := q.downloader.bulkhead.Execute(func() error {
				return downloadFunc(item)
			})
			
			mu.Lock()
			completed++
			mu.Unlock()
			
			if err != nil && q.onError != nil {
				q.onError(item.Name, err)
			}
		}(item)
	}
	
	wg.Wait()
	
	// Final progress report
	if q.onProgress != nil {
		q.onProgress(completed, total, "")
	}
	
	return nil
}

// GetMaxConcurrency returns the configured max concurrency
func (pd *ParallelDownloader) GetMaxConcurrency() int {
	return pd.maxConcurrent
}