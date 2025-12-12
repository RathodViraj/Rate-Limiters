# Rate Limiter Strategies

This project implements and compares four different rate limiting strategies commonly used in modern web applications. Each strategy has its own approach to controlling request flow with specific trade-offs.

## Table of Contents

1. [Token Bucket](#token-bucket)
2. [Leaky Bucket](#leaky-bucket)
3. [Fixed Window Counter](#fixed-window-counter)
4. [Sliding Window Log](#sliding-window-log)
5. [Sliding Window Counter](#sliding-window-counter)

---

## Token Bucket

### How It Works

The Token Bucket uses a virtual "bucket" filled with tokens at a constant rate:
- Tokens are added to the bucket at a fixed rate (e.g., 1 token per 100ms)
- Each request consumes a token to proceed
- Bucket has a maximum capacity (prevents unlimited bursts)
- If no tokens available, request is rejected or queued

**Example**: With 1 token per second and bucket capacity of 10:
- Bucket starts with 0-10 tokens
- Tokens accumulate at 1 token/second (capped at 10)
- Each request uses 1 token
- Can handle bursts up to 10 requests when bucket is full

### Implementation Details
- **Storage**: In-memory state with timestamp tracking
- **Scope**: Global (applies to all users/requests)
- **Refill Rate**: Configurable tokens per second
- **Bucket Capacity**: Configurable maximum tokens

### Pros
- **Allows Bursts**: Can handle sudden traffic spikes (up to bucket capacity)
- **Smooth Recovery**: Gradually refills tokens, smoothing traffic over time
- **Simple & Efficient**: Lightweight implementation with good performance
- **Flexible**: Can adjust token consumption for different request types
- **No Dependencies**: Pure in-memory, no external services needed
- **Predictable Throughput**: Guaranteed average throughput over time

### Cons
- **Burst Unfairness**: Early bursts can consume all tokens, starving later requests
- **Implementation Complexity**: Slightly more complex than fixed window
- **Capacity Tuning**: Requires careful tuning of bucket capacity
- **Per-Client Complexity**: Harder to implement per-user limits in distributed systems
- **Token Timing**: Edge cases with token expiration/refresh timing

---

## Leaky Bucket

### How It Works

The Leaky Bucket uses a queue with a fixed leak rate. Think of it as a bucket with a hole at the bottom:
- Incoming requests are added to a queue (bucket)
- Requests are processed from the queue at a constant, predictable rate (leaking)
- If the queue is full, new requests are rejected

**Example**: With a queue size of 8 and leak rate of 100ms per request:
- First 8 requests: Queued
- 9th request onwards: Rejected
- Requests are processed at a steady 100ms interval

### Implementation Details
- **Storage**: In-memory buffered channel
- **Scope**: Global (applies to all users/requests)
- **Processing**: Background goroutine processes queue at fixed intervals
- **Queue Capacity**: Configurable (default 8)
- **Leak Rate**: Configurable time interval between request processing

### Pros
- **Smooth Traffic**: Provides very smooth, predictable outflow at a constant rate
- **Burst Handling**: Queues handle bursts gracefully up to queue capacity
- **Fair Processing**: Requests are processed in FIFO order
- **Predictable Latency**: Request processing time is consistent
- **No Dependencies**: Pure in-memory implementation

### Cons
- **Queue Memory**: Must store queued requests in memory (limited by queue size)
- **Latency for Queued Requests**: Requests in queue experience variable latency
- **Inflexible Rate**: Can't easily adapt to changing traffic patterns
- **Lost Requests**: Requests exceeding queue capacity are completely dropped
- **Resource Intensive**: Background goroutine continuously running

---

## Fixed Window Counter

### How It Works

The Fixed Window Counter divides time into fixed-size windows (e.g., 5-second windows). For each window:
- A counter tracks the number of requests received
- When the limit is reached, subsequent requests are rejected with a 429 status
- When the window expires, the counter resets to zero

**Example**: With a limit of 3 requests per 5 seconds:
- Requests 0-2: Accepted
- Requests 3-12: Rejected
- After 5 seconds: Counter resets, next window allows 3 more requests

### Implementation Details
- **Storage**: In-memory map with mutex locking
- **Scope**: Global (applies to all users/requests)
- **Time Precision**: Second-level granularity

### Pros
- **Simple Implementation**: Very straightforward to understand and implement
- **Low Memory**: Minimal memory overhead (only stores one counter per window)
- **Fast Lookups**: O(1) complexity for checking limits
- **No Dependencies**: Pure in-memory, no external services needed

### Cons
- **Burst at Window Boundaries**: Clients can exploit the window boundary to send double the limit in a short time (e.g., 3 requests at 4.9s and 3 requests at 5.1s)
- **Not Truly Fair**: Within a window, all requests are equal - doesn't account for burst patterns
- **Rough Time Boundaries**: Requests near the window edge may feel unfair
- **Poor for Short Bursts**: Can't distinguish between sustained traffic and sudden spikes

---

## Sliding Window Log

### How It Works

The Sliding Window Log tracks individual request timestamps in a rolling window:
- Each request timestamp is recorded in a log
- Old timestamps outside the window are removed
- Current request count is compared against the limit

**Example**: With a limit of 6 requests per 10 seconds:
- Records timestamps of requests: [100, 101, 102, 103, 104, 105]
- At time 110s, timestamps older than 100s are removed
- New request check: count of requests in [100, 110] against limit

### Implementation Details
- **Storage**: Redis sorted sets (timestamp-based storage)
- **Scope**: Global (applies to all users/requests)
- **Precision**: Millisecond-level accuracy
- **Window**: True sliding window, not fixed buckets

### Pros
- **No Boundary Issues**: Smooth handling across any time boundary
- **Accurate**: True sliding window provides the most accurate rate limiting
- **Flexible Time Windows**: Works well with any time window specification
- **Per-Minute/Hour Limits**: Naturally handles metrics like "10 requests per minute"
- **Historical Data**: Can query exact request history for analytics

### Cons
- **High Memory**: Stores every request timestamp, uses significant memory
- **Redis Dependency**: Requires external Redis for distributed systems
- **Slow Operations**: Redis operations for every request add latency
- **Scaling Issues**: Memory grows with traffic volume (O(n) where n = number of requests)
- **Network Overhead**: Each request requires Redis network call
- **Garbage Collection**: Requires periodic cleanup of old timestamps

---

## Sliding Window Counter

### How It Works

The Sliding Window Counter is a hybrid approach that combines elements of Fixed Window and Sliding Window Log:
- Uses a fixed number of buckets (windows) that slide together
- Each bucket tracks request count for a specific time period
- Current and previous bucket counts are evaluated within the current window
- Provides more accuracy than fixed window with less overhead than full logging

**Example**: With a limit of 6 requests per 10 seconds using 10 buckets:
- At time 5s: Check current bucket + partial weight of previous bucket
- At time 10.5s: Window slides, old buckets are discarded
- Smoother rate limiting than pure fixed window

### Implementation Details
- **Storage**: Redis (for distributed systems, tracks buckets and timestamps)
- **Scope**: Per-client (tracks by IP address)
- **Precision**: Configurable bucket size
- **Window**: Sliding window with fixed number of buckets

### Pros
- **Better Accuracy**: Smoother than fixed window counter
- **Lower Memory**: Less memory intensive than sliding window log
- **No Boundary Issues**: Handles window boundaries better than fixed window
- **Distributed Ready**: Uses Redis for multi-instance deployments
- **Per-User Tracking**: Can easily track per-IP or per-client limits

### Cons
- **Redis Dependency**: Requires external Redis for operation
- **Moderate Complexity**: More complex than fixed window but simpler than sliding window log
- **Network Latency**: Redis calls add latency for each request
- **Bucket Tuning**: Requires tuning the number and size of buckets
- **Less Accurate**: Still not as accurate as true sliding window log

---

## When to Use Each Strategy

### Use Token Bucket When:
- You need to **allow controlled bursts**
- You want smooth average throughput with flexibility
- You have variable request costs (some requests worth more tokens)
- You need good performance without external dependencies
- **Examples**: API rate limiting with burst allowance, bandwidth throttling, load leveling

### Use Leaky Bucket When:
- You need **smooth, predictable traffic flow**
- Processing requests at a constant rate is critical
- You can tolerate queued requests with variable latency
- System has capacity to queue requests temporarily
- **Examples**: Video streaming, print queues, network packet processing

### Use Fixed Window Counter When:
- You need the **simplest possible implementation**
- Memory is extremely limited
- Slight bursts at boundaries are acceptable
- You're rate limiting per IP or user (fixed windows naturally partition by time)
- **Examples**: Basic API throttling, simple DDoS protection

### Use Sliding Window Log When:
- You need **maximum accuracy** and fairness
- You're tracking metrics for billing/analytics
- Memory is not a constraint
- You have a distributed system with Redis available
- **Examples**: API billing, SaaS quota management, precise request tracking

### Use Sliding Window Counter When:
- You need **better accuracy than fixed window** without full logging
- You want a **middle ground** between simplicity and accuracy
- You need per-user/IP rate limiting
- You're in a **distributed system** with Redis available
- **Examples**: API rate limiting, micro-service quota management, balanced rate limiting

---

## Testing

Each strategy includes a test file that demonstrates how it works:

```bash
# Run token bucket test
go test -v ./rate-limiters/token-bucket/

# Run leaky bucket test
go test -v ./rate-limiters/leaky-bucket/

# Run fixed window counter test
go test -v ./rate-limiters/fixed-window-counter/

# Run sliding window log test
go test -v ./rate-limiters/sliding-window-log/

# Run sliding window counter test
go test -v ./rate-limiters/sliding-window-counter/
```

**Note**: Tests make HTTP requests to a server running on `http://localhost:8080/home`. Start the server first:

```bash
go run cmd/main.go
```

---

## Implementation Notes

- **Global Limiting**: Token Bucket and Leaky Bucket implementations apply limits globally to all requests
- **Per-User Limiting**: Fixed Window Counter, Sliding Window Log, and Sliding Window Counter track per IP/user
- **Distributed Systems**: Sliding Window Log and Sliding Window Counter require Redis for distributed rate limiting
- **In-Memory**: Token Bucket and Leaky Bucket use pure in-memory implementations
- **Metrics**: Sliding Window Log naturally supports detailed analytics and metrics

---

## References
- https://bytebytego.com/courses/system-design-interview/design-a-rate-limiter
- https://youtu.be/CVItTb_jdkE?si=mAn-B9bXuUTx2JFY


