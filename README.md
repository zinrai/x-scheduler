# x-scheduler

A command-line tool for scheduling X ( Twitter ) posts using YAML configuration files.

## Features

- **Declarative Configuration**: Define posts and schedules in YAML
- **Daily Batch Processing**: Execute all posts scheduled for today in a single run
- **Stateless Design**: No database or persistent state required
- **RFC 3339 Time Format**: Standard-compliant time specifications
- **Test Mode**: Test posts immediately with dry-run capability

### System Requirements

- Linux/Unix system with cron (optional)
- [xurl](https://github.com/xdevplatform/xurl) command installed and configured

## Installation

### Prerequisites

`x-scheduler` uses [xurl](https://github.com/xdevplatform/xurl) for X API authentication and posts, which provides.

### Build

```bash
$ go build -o x-scheduler cmd/main.go
```

## Configuration

### YAML Configuration File

See `example.yaml`.

#### Configuration Fields

- `content` (required): The text content of your post
- `scheduled_at` (required): When to post in RFC 3339 format
- `enabled` (optional): Set to `true` to enable the post (default: `false`)
- `test` (optional): Set to `true` to execute immediately for testing (default: `false`)
- `dry_run` (optional): Set to `true` to simulate posting without actually posting (requires `test: true`)

## Usage

### Validate Configuration

Check your configuration file for errors:

```bash
$ x-scheduler -validate config.yaml
```

Example output:
```
Configuration validation successful
Total posts: 6
Enabled posts: 4
Future posts for today: 2
Poster: xurl command available

Upcoming posts for today:
  08:00: Good morning! Ready to tackle the day ahead!
  17:00: Weekly development update: Shipped 3 features this...
```

### Execute Posts

Execute all posts scheduled for today:

```bash
$ x-scheduler -execute config.yaml
```

Example output:
```
[INFO] Starting execution
[INFO] Skipping past post: Yesterday's post (scheduled at 10:00:00)
[INFO] Found 3 posts scheduled for execution
[INFO] Queuing post: Good morning! (in 30m0s at 08:00:00)
[INFO] Queuing immediate post: Testing API connection
[INFO] Test post: Testing API connection
✓ Test post successful: Testing API connection
[INFO] Waiting 29m45s until execution time (08:00:00)
[INFO] Posting: Good morning! Ready to tackle the day ahead!
[INFO] Post successful: Good morning! Ready to tackle the day ahead!
[INFO] Execution completed: 2 successful, 0 failed
```

### Command Line Options

```
  -execute    Execute posts scheduled for today
  -validate   Validate configuration file
  -verbose    Enable verbose logging
  -version    Show version information
  -help       Show help message
```

## How It Works

x-scheduler uses a **daily batch processing model**:

1. **Execution**: `x-scheduler -execute config.yaml`
   - Loads configuration at execution time
   - Filters posts to include only today's future posts
   - Automatically skips posts scheduled in the past
   - Sorts posts by execution time
   - Queues posts in a channel-based job queue
   - Processes posts sequentially, waiting until each post's scheduled time
   - Executes test posts immediately regardless of schedule
   - Posts to X API via xurl with detailed error logging
   - Terminates after all posts are processed

2. **Scheduling**:
   - **00:00 execution**: Processes all posts scheduled for the day
   - **10:00 execution**: Skips 08:00 and 09:00 posts, processes 12:00+ posts
   - **Manual execution**: Run anytime to process remaining posts for the day

## Scheduling Options

### Daily Cron (Recommended)

Execute once per day at midnight.

```bash
0 0 * * * /usr/local/bin/x-scheduler -execute /path/to/config.yaml
```

### Manual Execution

Run manually anytime to process today's remaining posts.

```bash
$ x-scheduler -execute config.yaml
```

## Handling Post Failures

When x-scheduler fails to post a tweet, it does not automatically retry. Instead, it logs detailed error information from xurl to help you detect and resolve issues.

Failed posts are not automatically retried.

- Fix the issue and run x-scheduler again the same day
- Move the failed post to a future date
- Check xurl configuration and authentication

## Flexible Past Post Handling

x-scheduler handles past posts gracefully:

- **Automatic skipping**: Past posts are automatically skipped during execution
- **History preservation**: You can keep past posts in your configuration file as history
- **No manual cleanup**: No need to manually disable or remove past posts
- **Informative logging**: Past posts are logged as skipped with timestamps

This design allows you to:
- Maintain a complete history of your scheduled posts
- Add new future posts without worrying about past entries
- Use the same configuration file over time without constant maintenance
- Run x-scheduler multiple times per day without duplicate posts

## License

This project is licensed under the MIT License - see the [LICENSE](https://opensource.org/license/mit) for details.
