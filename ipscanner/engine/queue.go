package engine

import (
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/shahradelahi/cloudflare-warp/ipscanner/model"
	"github.com/shahradelahi/cloudflare-warp/log"
)

type IPQueue struct {
	queue        []statute.IPInfo
	maxQueueSize int
	mu           sync.Mutex
	available    chan struct{}
	maxTTL       time.Duration
	rttThreshold time.Duration
	inIdealMode  bool
	reserved     statute.IPInfQueue
}

func NewIPQueue(opts *statute.ScannerOptions) *IPQueue {
	var reserved statute.IPInfQueue
	return &IPQueue{
		queue:        make([]statute.IPInfo, 0),
		maxQueueSize: opts.IPQueueSize,
		maxTTL:       opts.IPQueueTTL,
		rttThreshold: opts.MaxDesirableRTT,
		available:    make(chan struct{}, opts.IPQueueSize),
		reserved:     reserved,
	}
}

func (q *IPQueue) Enqueue(info statute.IPInfo) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, existingIP := range q.queue {
		if existingIP.AddrPort == info.AddrPort {
			return false
		}
	}

	defer func() {
		log.Debugw("IP queue state change", zap.Int("current_size", len(q.queue)))
		for _, ipInfo := range q.queue {
			log.Debugw(
				"IP queue item details",
				zap.Time("created", ipInfo.CreatedAt),
				zap.Stringer("addr", ipInfo.AddrPort),
				zap.Duration("rtt", ipInfo.RTT),
			)
		}
	}()

	log.Debug("Enqueue: Sorting queue by RTT")
	sort.Slice(q.queue, func(i, j int) bool {
		return q.queue[i].RTT < q.queue[j].RTT
	})

	if len(q.queue) == 0 {
		log.Debug("Enqueue: empty queue adding first available item")
		q.queue = append(q.queue, info)
		return false
	}

	if info.RTT <= q.rttThreshold {
		log.Debug("Enqueue: the new item's RTT is less than at least one of the members.")
		if len(q.queue) >= q.maxQueueSize && info.RTT < q.queue[len(q.queue)-1].RTT {
			log.Debug("Enqueue: the queue is full, remove the item with the highest RTT.")
			q.queue = q.queue[:len(q.queue)-1]
		} else if len(q.queue) < q.maxQueueSize {
			log.Debug("Enqueue: Insert the new item in a sorted position.")
			index := sort.Search(len(q.queue), func(i int) bool { return q.queue[i].RTT > info.RTT })
			q.queue = append(q.queue[:index], append([]statute.IPInfo{info}, q.queue[index:]...)...)
		} else {
			log.Debug("Enqueue: The Queue is full but we keep the new item in the reserved queue.")
			q.reserved.Enqueue(info)
		}
	}

	log.Debug("Enqueue: Checking if any member has a higher RTT than the threshold.")
	for _, member := range q.queue {
		if member.RTT > q.rttThreshold {
			return false // If any member has a higher RTT than the threshold, return false.
		}
	}

	log.Debug("Enqueue: All members have an RTT lower than the threshold.")
	if len(q.queue) < q.maxQueueSize {
		// the queue isn't full dont wait
		return false
	}

	q.inIdealMode = true
	// ok wait for expiration signal
	log.Debug("Enqueue: All members have an RTT lower than the threshold. Waiting for expiration signal.")
	return true
}

func (q *IPQueue) Dequeue() (statute.IPInfo, bool) {
	defer func() {
		log.Debugw("IP queue state change", zap.Int("current_size", len(q.queue)))
		for _, ipInfo := range q.queue {
			log.Debugw(
				"IP queue item details",
				zap.Time("created", ipInfo.CreatedAt),
				zap.Stringer("addr", ipInfo.AddrPort),
				zap.Duration("rtt", ipInfo.RTT),
			)
		}
	}()
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queue) == 0 {
		return statute.IPInfo{}, false
	}

	info := q.queue[len(q.queue)-1]
	q.queue = q.queue[0 : len(q.queue)-1]

	q.available <- struct{}{}

	return info, true
}

func (q *IPQueue) Init() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.inIdealMode {
		q.available <- struct{}{}
		return
	}
}

func (q *IPQueue) Expire() {
	q.mu.Lock()
	defer q.mu.Unlock()

	log.Debug("Expire: In ideal mode")
	defer func() {
		log.Debugw("IP queue state change", zap.Int("current_size", len(q.queue)))
		for _, ipInfo := range q.queue {
			log.Debugw(
				"IP queue item details",
				zap.Time("created", ipInfo.CreatedAt),
				zap.Stringer("addr", ipInfo.AddrPort),
				zap.Duration("rtt", ipInfo.RTT),
			)
		}
	}()

	shouldStartNewScan := false
	resQ := make([]statute.IPInfo, 0)
	for i := 0; i < len(q.queue); i++ {
		if time.Since(q.queue[i].CreatedAt) > q.maxTTL {
			log.Debug("Expire: Removing expired item from queue")
			shouldStartNewScan = true
		} else {
			resQ = append(resQ, q.queue[i])
		}
	}
	q.queue = resQ
	log.Debug("Expire: Adding reserved items to queue")
	for i := 0; i < q.maxQueueSize && i < q.reserved.Size(); i++ {
		q.queue = append(q.queue, q.reserved.Dequeue())
	}
	if shouldStartNewScan {
		q.available <- struct{}{}
	}
}

func (q *IPQueue) AvailableIPs(desc bool) []statute.IPInfo {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Create a separate slice for sorting
	sortedQueue := make([]statute.IPInfo, len(q.queue))
	copy(sortedQueue, q.queue)

	// Sort by RTT ascending/descending
	sort.Slice(sortedQueue, func(i, j int) bool {
		if desc {
			return sortedQueue[i].RTT > sortedQueue[j].RTT
		}
		return sortedQueue[i].RTT < sortedQueue[j].RTT
	})

	return sortedQueue
}

func (q *IPQueue) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.queue)
}
